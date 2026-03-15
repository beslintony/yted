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
  const [downloadId, setDownloadId] = useState<string | null>(null);
  const [downloading, setDownloading] = useState(false);
  
  const { downloads, removeDownload, pauseDownload, resumeDownload, retryDownload, updateProgress } = useDownloadStore();
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

  // Listen for download progress events
  useEffect(() => {
    const cancel = EventsOn('download:progress', (data: any) => {
      if (data && data.id && data.progress !== undefined) {
        updateProgress(data.id, data.progress);
      }
    });
    return () => cancel();
  }, []);

  const handleFetchInfo = async () => {
    if (!url.trim()) return;
    
    setLoading(true);
    setError(null);
    setVideoInfo(null);
    setDownloadId(null);
    
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
    if (!videoInfo || !url) return;
    
    setDownloading(true);
    setError(null);
    
    try {
      // Use best quality format by default
      const formatId = videoInfo.formats?.find(f => f.resolution?.includes('1080') || f.resolution?.includes('720'))?.format_id 
        || videoInfo.formats?.[0]?.format_id 
        || 'best';
      const quality = 'best';
      
      const id = await AddDownload(url, formatId, quality);
      setDownloadId(id);
      
      // Don't clear the form - let user see what they downloaded
      // setUrl('');
      // setVideoInfo(null);
    } catch (err: any) {
      setError(err?.message || 'Failed to add download');
    } finally {
      setDownloading(false);
    }
  };

  const handleClear = () => {
    setUrl('');
    setVideoInfo(null);
    setError(null);
    setDownloadId(null);
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

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed': return <IconCheck size={16} />;
      case 'downloading': return <Loader size={16} />;
      case 'error': return <IconAlertCircle size={16} />;
      case 'paused': return <IconPlayerPause size={16} />;
      default: return <IconDownload size={16} />;
    }
  };

  const formatDuration = (seconds: number) => {
    if (!seconds) return 'Unknown';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  // Find the current download in the store
  const currentDownload = downloadId ? downloads.find(d => d.id === downloadId) : null;

  return (
    <Stack spacing="lg">
      <Text size="xl" fw={700} c={dark ? '#fff' : '#000'}>Downloads</Text>
      
      {/* URL Input */}
      <Paper 
        p="md" 
        withBorder 
        sx={{ 
          background: dark ? '#25262b' : '#fff',
          borderColor: dark ? '#373a40' : '#dee2e6',
        }}
      >
        <Stack spacing="md">
          <TextInput
            placeholder="Paste YouTube URL here..."
            icon={<IconLink size={16} />}
            value={url}
            onChange={(e) => setUrl(e.currentTarget.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleFetchInfo()}
            disabled={loading || downloading}
            size="md"
            sx={{
              '& input': {
                background: dark ? '#141517' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            rightSection={
              url && (
                <ActionIcon onClick={handleClear} color="gray" variant="subtle">
                  <IconX size={16} />
                </ActionIcon>
              )
            }
          />
          
          <Group position="right">
            <Button
              onClick={handleFetchInfo}
              loading={loading}
              leftIcon={<IconDownload size={16} />}
              disabled={!url.trim() || loading}
              color="red"
            >
              Get Info
            </Button>
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
              sx={{ 
                background: dark ? '#1a1b1e' : '#f8f9fa',
                borderColor: dark ? '#373a40' : '#dee2e6',
              }}
            >
              <Group align="flex-start" noWrap>
                {videoInfo.thumbnail ? (
                  <img
                    src={videoInfo.thumbnail}
                    alt={videoInfo.title}
                    style={{ 
                      width: 160, 
                      height: 90, 
                      objectFit: 'cover', 
                      borderRadius: 8,
                      background: dark ? '#2c2e33' : '#e9ecef',
                    }}
                  />
                ) : (
                  <Paper 
                    w={160} 
                    h={90} 
                    sx={{ 
                      background: dark ? '#2c2e33' : '#e9ecef',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: 8,
                    }}
                  >
                    <IconVideo size={32} color={dark ? '#5c5f66' : '#adb5bd'} />
                  </Paper>
                )}
                <Stack spacing="xs" sx={{ flex: 1 }}>
                  <Text fw={600} lineClamp={2} c={dark ? '#fff' : '#000'}>
                    {videoInfo.title || 'Unknown Title'}
                  </Text>
                  <Text size="sm" c={dark ? 'dimmed' : 'gray.7'}>
                    {videoInfo.channel || 'Unknown Channel'}
                  </Text>
                  <Text size="sm" c={dark ? 'dimmed' : 'gray.7'}>
                    Duration: {formatDuration(videoInfo.duration)}
                  </Text>
                  {videoInfo.formats && videoInfo.formats.length > 0 && (
                    <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
                      Available qualities: {videoInfo.formats
                        .filter(f => f.resolution && f.resolution !== 'audio only')
                        .map(f => f.resolution)
                        .filter((v, i, a) => a.indexOf(v) === i)
                        .slice(0, 5)
                        .join(', ')}
                    </Text>
                  )}
                  <Group position="right" mt="xs">
                    {currentDownload ? (
                      <Badge 
                        size="lg" 
                        color={getStatusColor(currentDownload.status)}
                        leftSection={getStatusIcon(currentDownload.status)}
                      >
                        {currentDownload.status === 'downloading' 
                          ? `Downloading ${Math.round(currentDownload.progress)}%`
                          : currentDownload.status}
                      </Badge>
                    ) : (
                      <Button 
                        size="sm" 
                        onClick={handleDownload} 
                        color="red"
                        loading={downloading}
                        leftIcon={<IconDownload size={16} />}
                      >
                        Download
                      </Button>
                    )}
                  </Group>
                </Stack>
              </Group>
              
              {currentDownload && currentDownload.status === 'downloading' && (
                <Progress
                  value={currentDownload.progress}
                  size="sm"
                  mt="md"
                  color="red"
                  radius="xs"
                />
              )}
            </Paper>
          )}
        </Stack>
      </Paper>

      {/* Download Queue */}
      <Group position="apart">
        <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>
          Queue ({downloads.length})
        </Text>
        {downloads.length > 0 && (
          <Button 
            size="xs" 
            variant="subtle" 
            color="gray"
            onClick={() => downloads.forEach(d => removeDownload(d.id))}
          >
            Clear All
          </Button>
        )}
      </Group>
      
      <ScrollArea sx={{ height: 'calc(100vh - 450px)' }}>
        <Stack spacing="sm">
          {downloads.length === 0 ? (
            <Paper 
              p="xl" 
              withBorder
              sx={{
                background: dark ? '#25262b' : '#fff',
                borderColor: dark ? '#373a40' : '#dee2e6',
              }}
            >
              <Text c={dark ? 'dimmed' : 'gray.6'} align="center">
                No downloads yet. Paste a YouTube URL above to get started.
              </Text>
            </Paper>
          ) : (
            downloads.map((download) => (
              <Paper 
                key={download.id} 
                p="sm" 
                withBorder
                sx={{
                  background: dark ? '#25262b' : '#fff',
                  borderColor: dark ? '#373a40' : '#dee2e6',
                }}
              >
                <Group position="apart" align="flex-start">
                  <Group spacing="sm" align="flex-start" noWrap>
                    {download.thumbnail ? (
                      <img
                        src={download.thumbnail}
                        alt={download.title || 'Video'}
                        style={{ 
                          width: 80, 
                          height: 45, 
                          objectFit: 'cover', 
                          borderRadius: 4,
                          background: dark ? '#2c2e33' : '#e9ecef',
                        }}
                      />
                    ) : (
                      <Paper 
                        w={80} 
                        h={45} 
                        sx={{ 
                          background: dark ? '#2c2e33' : '#e9ecef',
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          borderRadius: 4,
                        }}
                      >
                        <IconVideo size={20} color={dark ? '#5c5f66' : '#adb5bd'} />
                      </Paper>
                    )}
                    <Stack spacing={4}>
                      <Text size="sm" fw={500} lineClamp={1} sx={{ maxWidth: 300 }} c={dark ? '#fff' : '#000'}>
                        {download.title || 'Loading...'}
                      </Text>
                      <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
                        {download.channel || download.url}
                      </Text>
                      <Group spacing="xs">
                        <Badge 
                          size="sm" 
                          color={getStatusColor(download.status)} 
                          leftSection={getStatusIcon(download.status)}
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
                  
                  <Stack spacing="xs" align="flex-end">
                    <Group spacing={4}>
                      {download.status === 'downloading' && (
                        <Tooltip label="Pause">
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
                        <Tooltip label="Resume">
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
                        <Tooltip label="Retry">
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
                      <Tooltip label="Remove">
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
                      <Progress
                        value={download.progress}
                        size="sm"
                        w={100}
                        color="red"
                        radius="xs"
                      />
                    )}
                    <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
                      {Math.round(download.progress)}%
                    </Text>
                  </Stack>
                </Group>
                
                {download.errorMessage && (
                  <Alert color="red" mt="sm" py="xs" sx={{ background: dark ? '#2c1b1b' : '#fff5f5' }}>
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

// Icon placeholder for when no thumbnail
function IconVideo({ size, color }: { size: number; color?: string }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke={color || "currentColor"} strokeWidth="2">
      <rect x="2" y="2" width="20" height="20" rx="2" />
      <polygon points="10 8 16 12 10 16" fill="currentColor" />
    </svg>
  );
}
