import { useState, useEffect } from 'react';
import {
  TextInput,
  Button,
  Group,
  Stack,
  Paper,
  Text,
  Progress,
  ActionIcon,
  Badge,
  Tooltip,
  Loader,
  Alert,
  ScrollArea,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconDownload,
  IconPlayerPlay,
  IconPlayerPause,
  IconRefresh,
  IconTrash,
  IconCheck,
  IconAlertCircle,
  IconLink,
  IconX,
  IconSearch,
} from '@tabler/icons-react';
import { useDownloadStore } from '../stores';
import { GetVideoInfo, AddDownload, ValidateURL } from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime';

export function DownloadPage() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [videoInfo, setVideoInfo] = useState<app.VideoInfoResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  
  const { 
    downloads, 
    addDownload, 
    removeDownload, 
    pauseDownload, 
    resumeDownload, 
    retryDownload, 
    updateProgress 
  } = useDownloadStore();
  
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

  // Listen for download progress events from backend
  useEffect(() => {
    const cancel = EventsOn('download:progress', (data: any) => {
      if (data && data.id && data.progress !== undefined) {
        updateProgress(data.id, data.progress);
      }
    });
    return () => cancel();
  }, [updateProgress]);

  const handleFetchInfo = async () => {
    if (!url.trim()) {
      setError('Please enter a YouTube URL');
      return;
    }
    
    setLoading(true);
    setError(null);
    setVideoInfo(null);
    
    try {
      const isValid = await ValidateURL(url);
      if (!isValid) {
        setError('Invalid YouTube URL. Please enter a valid YouTube video URL.');
        return;
      }
      
      const info = await GetVideoInfo(url);
      if (!info || !info.id) {
        setError('Could not fetch video information. Please check the URL and try again.');
        return;
      }
      setVideoInfo(info);
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch video info');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    if (!videoInfo || !url) {
      setError('No video info available');
      return;
    }
    
    setAdding(true);
    setError(null);
    
    try {
      // Find best format
      const format = videoInfo.formats?.find(f => 
        f.resolution?.includes('1080') || f.resolution?.includes('720')
      ) || videoInfo.formats?.[0];
      
      const formatId = format?.format_id || 'best';
      const quality = format?.quality || 'best';
      
      console.log('Adding download:', { url, formatId, quality, title: videoInfo.title });
      
      // Call backend to add download
      const id = await AddDownload(url, formatId, quality);
      console.log('Download added with ID:', id);
      
      // Add to local store for immediate UI update
      addDownload(url, {
        id: videoInfo.id,
        title: videoInfo.title,
        channel: videoInfo.channel,
        channelId: videoInfo.channel_id,
        duration: videoInfo.duration,
        description: videoInfo.description,
        thumbnail: videoInfo.thumbnail,
        formats: videoInfo.formats as any,
      }, {
        formatId: format?.format_id || 'best',
        ext: format?.ext || 'mp4',
        resolution: format?.resolution || '',
        fps: format?.fps || 0,
        vcodec: format?.vcodec || '',
        acodec: format?.acodec || '',
        filesize: format?.filesize || 0,
        quality: format?.quality || 'best',
      });
      
      // Clear the form after successful add
      setUrl('');
      setVideoInfo(null);
    } catch (err: any) {
      console.error('Failed to add download:', err);
      setError(err?.message || 'Failed to add download');
    } finally {
      setAdding(false);
    }
  };

  const handleClear = () => {
    setUrl('');
    setVideoInfo(null);
    setError(null);
  };

  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed': return 'green';
      case 'downloading': return 'blue';
      case 'error': return 'red';
      case 'paused': return 'yellow';
      default: return 'gray';
    }
  };

  const formatDuration = (seconds: number) => {
    if (!seconds) return 'Unknown';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <Stack gap="lg">
      <Text size="xl" fw={700} c={dark ? '#fff' : '#000'}>Downloads</Text>
      
      {/* URL Input */}
      <Paper 
        p="md" 
        withBorder 
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <TextInput
            placeholder="Paste YouTube URL here..."
            leftSection={<IconLink size={16} />}
            value={url}
            onChange={(e) => setUrl(e.currentTarget.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleFetchInfo()}
            disabled={loading}
            size="md"
            styles={{
              input: {
                background: dark ? '#141517' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            rightSection={
              url ? (
                <Tooltip label="Clear">
                  <ActionIcon onClick={handleClear} color="gray" variant="subtle">
                    <IconX size={16} />
                  </ActionIcon>
                </Tooltip>
              ) : undefined
            }
          />
          
          <Group justify="flex-end">
            <Tooltip label="Get video information">
              <Button
                onClick={handleFetchInfo}
                loading={loading}
                leftSection={<IconSearch size={16} />}
                disabled={!url.trim() || loading}
                color="yted"
              >
                Get Info
              </Button>
            </Tooltip>
          </Group>
          
          {error && (
            <Alert color="red" icon={<IconAlertCircle size={16} />}>
              {error}
            </Alert>
          )}
          
          {videoInfo && (
            <Paper 
              p="md" 
              withBorder 
              bg={dark ? '#1a1b1e' : '#f8f9fa'}
              style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
            >
              <Group align="flex-start" wrap="nowrap">
                {videoInfo.thumbnail ? (
                  <img
                    src={videoInfo.thumbnail}
                    alt={videoInfo.title}
                    style={{ 
                      width: 160, 
                      height: 90, 
                      objectFit: 'cover', 
                      borderRadius: 8,
                    }}
                  />
                ) : (
                  <Paper 
                    w={160} 
                    h={90}
                    bg={dark ? '#2c2e33' : '#e9ecef'}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: 8,
                    }}
                  >
                    <IconVideo size={32} color={dark ? '#5c5f66' : '#adb5bd'} />
                  </Paper>
                )}
                <Stack gap="xs" style={{ flex: 1 }}>
                  <Text fw={600} lineClamp={2} c={dark ? '#fff' : '#000'}>
                    {videoInfo.title || 'Unknown Title'}
                  </Text>
                  <Text size="sm" c={dark ? 'dimmed' : 'gray.7'}>
                    {videoInfo.channel || 'Unknown Channel'}
                  </Text>
                  <Text size="sm" c={dark ? 'dimmed' : 'gray.7'}>
                    Duration: {formatDuration(videoInfo.duration)}
                  </Text>
                  <Group justify="flex-end" mt="xs">
                    <Tooltip label="Add to download queue">
                      <Button 
                        size="sm" 
                        onClick={handleDownload} 
                        color="yted"
                        loading={adding}
                        leftSection={<IconDownload size={16} />}
                      >
                        Download
                      </Button>
                    </Tooltip>
                  </Group>
                </Stack>
              </Group>
            </Paper>
          )}
        </Stack>
      </Paper>

      {/* Download Queue */}
      <Group justify="space-between">
        <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>
          Queue ({downloads.length})
        </Text>
        {downloads.length > 0 && (
          <Tooltip label="Clear all downloads">
            <Button 
              size="xs" 
              variant="subtle" 
              color="gray"
              onClick={() => downloads.forEach(d => removeDownload(d.id))}
            >
              Clear All
            </Button>
          </Tooltip>
        )}
      </Group>
      
      <ScrollArea h="calc(100vh - 450px)">
        <Stack gap="sm">
          {downloads.length === 0 ? (
            <Paper 
              p="xl" 
              withBorder
              bg={dark ? '#25262b' : '#fff'}
              style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
            >
              <Text c={dark ? 'dimmed' : 'gray.6'} ta="center">
                No downloads yet. Paste a YouTube URL above to get started.
              </Text>
            </Paper>
          ) : (
            downloads.map((download) => (
              <Paper 
                key={download.id} 
                p="sm" 
                withBorder
                bg={dark ? '#25262b' : '#fff'}
                style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
              >
                <Group justify="space-between" align="flex-start">
                  <Group gap="sm" align="flex-start" wrap="nowrap">
                    {download.thumbnail ? (
                      <img
                        src={download.thumbnail}
                        alt={download.title || 'Video'}
                        style={{ 
                          width: 80, 
                          height: 45, 
                          objectFit: 'cover', 
                          borderRadius: 4,
                        }}
                      />
                    ) : (
                      <Paper 
                        w={80} 
                        h={45}
                        bg={dark ? '#2c2e33' : '#e9ecef'}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          borderRadius: 4,
                        }}
                      >
                        <IconVideo size={20} color={dark ? '#5c5f66' : '#adb5bd'} />
                      </Paper>
                    )}
                    <Stack gap={4}>
                      <Text size="sm" fw={500} lineClamp={1} style={{ maxWidth: 300 }} c={dark ? '#fff' : '#000'}>
                        {download.title || 'Loading...'}
                      </Text>
                      <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
                        {download.channel || download.url}
                      </Text>
                      <Group gap="xs">
                        <Badge 
                          size="sm" 
                          color={getStatusColor(download.status)}
                          leftSection={
                            download.status === 'completed' ? <IconCheck size={12} /> :
                            download.status === 'downloading' ? <Loader size={12} /> :
                            download.status === 'error' ? <IconAlertCircle size={12} /> :
                            download.status === 'paused' ? <IconPlayerPause size={12} /> :
                            <IconDownload size={12} />
                          }
                        >
                          {download.status}
                        </Badge>
                        {download.quality && (
                          <Badge size="sm" variant="outline" color={dark ? 'gray' : 'dark'}>
                            {download.quality}
                          </Badge>
                        )}
                      </Group>
                    </Stack>
                  </Group>
                  
                  <Stack gap="xs" align="flex-end">
                    <Group gap={4}>
                      {download.status === 'downloading' && (
                        <Tooltip label="Pause download">
                          <ActionIcon 
                            size="sm" 
                            onClick={() => pauseDownload(download.id)}
                            variant="light"
                            color="yellow"
                          >
                            <IconPlayerPause size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'paused' && (
                        <Tooltip label="Resume download">
                          <ActionIcon 
                            size="sm" 
                            onClick={() => resumeDownload(download.id)}
                            variant="light"
                            color="green"
                          >
                            <IconPlayerPlay size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'error' && (
                        <Tooltip label="Retry download">
                          <ActionIcon 
                            size="sm" 
                            onClick={() => retryDownload(download.id)}
                            variant="light"
                            color="blue"
                          >
                            <IconRefresh size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      <Tooltip label="Remove from queue">
                        <ActionIcon 
                          size="sm" 
                          color="red" 
                          onClick={() => removeDownload(download.id)}
                          variant="light"
                        >
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Tooltip>
                    </Group>
                    
                    {download.status === 'downloading' && (
                      <Tooltip label={`${Math.round(download.progress)}% complete`}>
                        <Progress
                          value={download.progress}
                          size="sm"
                          w={100}
                          color="yted"
                          radius="xs"
                        />
                      </Tooltip>
                    )}
                    <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
                      {Math.round(download.progress)}%
                    </Text>
                  </Stack>
                </Group>
                
                {download.errorMessage && (
                  <Alert color="red" mt="sm" p="xs" bg={dark ? '#2c1b1b' : '#fff5f5'}>
                    <Text size="xs">{download.errorMessage}</Text>
                  </Alert>
                )}
              </Paper>
            ))
          )}
        </Stack>
      </ScrollArea>
    </Stack>
  );
}

// Icon placeholder
function IconVideo({ size, color }: { size: number; color?: string }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color || "currentColor"} strokeWidth="2">
      <rect x="2" y="2" width="20" height="20" rx="2" />
      <polygon points="10 8 16 12 10 16" fill="currentColor" />
    </svg>
  );
}
