import { useState } from 'react';
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
} from '@tabler/icons-react';
import { useDownloadStore } from '../stores';
import { GetVideoInfo, AddDownload, ValidateURL } from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';

export function DownloadPage() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [videoInfo, setVideoInfo] = useState<app.VideoInfoResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  
  const { downloads, removeDownload, pauseDownload, resumeDownload, retryDownload } = useDownloadStore();

  const handleFetchInfo = async () => {
    if (!url.trim()) return;
    
    setLoading(true);
    setError(null);
    setVideoInfo(null);
    
    try {
      const isValid = await ValidateURL(url);
      if (!isValid) {
        setError('Invalid YouTube URL');
        return;
      }
      
      const info = await GetVideoInfo(url);
      setVideoInfo(info);
    } catch (err: any) {
      setError(err?.message || 'Failed to fetch video info');
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    if (!videoInfo) return;
    
    try {
      // Use best quality format by default
      const formatId = videoInfo.formats?.[0]?.format_id || 'best';
      const quality = 'best';
      
      await AddDownload(url, formatId, quality);
      
      // Clear the form
      setUrl('');
      setVideoInfo(null);
    } catch (err: any) {
      setError(err?.message || 'Failed to add download');
    }
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

  return (
    <Stack spacing="lg">
      <Text size="xl" fw={700}>Downloads</Text>
      
      {/* URL Input */}
      <Paper p="md" withBorder>
        <Stack spacing="md">
          <TextInput
            placeholder="Paste YouTube URL here..."
            icon={<IconLink size={16} />}
            value={url}
            onChange={(e) => setUrl(e.currentTarget.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleFetchInfo()}
            disabled={loading}
            size="md"
          />
          
          <Group position="right">
            <Button
              onClick={handleFetchInfo}
              loading={loading}
              leftIcon={<IconDownload size={16} />}
              disabled={!url.trim()}
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
            <Paper p="md" withBorder bg="dark.6">
              <Group align="flex-start" noWrap>
                {videoInfo.thumbnail && (
                  <img
                    src={videoInfo.thumbnail}
                    alt={videoInfo.title}
                    style={{ width: 160, height: 90, objectFit: 'cover', borderRadius: 4 }}
                  />
                )}
                <Stack spacing="xs" style={{ flex: 1 }}>
                  <Text fw={500} lineClamp={2}>{videoInfo.title}</Text>
                  <Text size="sm" c="dimmed">{videoInfo.channel}</Text>
                  <Text size="sm" c="dimmed">
                    {videoInfo.duration ? `${Math.floor(videoInfo.duration / 60)}:${(videoInfo.duration % 60).toString().padStart(2, '0')}` : 'Unknown duration'}
                  </Text>
                  <Group position="right" mt="xs">
                    <Button size="sm" onClick={handleDownload} color="yted">
                      Download
                    </Button>
                  </Group>
                </Stack>
              </Group>
            </Paper>
          )}
        </Stack>
      </Paper>

      {/* Download Queue */}
      <Text size="lg" fw={600}>Queue ({downloads.length})</Text>
      
      <ScrollArea style={{ height: 'calc(100vh - 400px)' }}>
        <Stack spacing="sm">
          {downloads.length === 0 ? (
            <Paper p="xl" withBorder>
              <Text c="dimmed" align="center">No downloads yet. Paste a YouTube URL above to get started.</Text>
            </Paper>
          ) : (
            downloads.map((download) => (
              <Paper key={download.id} p="sm" withBorder>
                <Group position="apart" align="flex-start">
                  <Group spacing="sm" align="flex-start" noWrap>
                    {download.thumbnail ? (
                      <img
                        src={download.thumbnail}
                        alt={download.title || 'Video'}
                        style={{ width: 80, height: 45, objectFit: 'cover', borderRadius: 4 }}
                      />
                    ) : (
                      <Paper w={80} h={45} bg="dark.6" style={{ display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
                        <IconVideo size={20} />
                      </Paper>
                    )}
                    <Stack spacing={4}>
                      <Text size="sm" fw={500} lineClamp={1} style={{ maxWidth: 300 }}>
                        {download.title || 'Loading...'}
                      </Text>
                      <Text size="xs" c="dimmed">{download.channel || download.url}</Text>
                      <Group spacing="xs">
                        <Badge size="sm" color={getStatusColor(download.status)} leftSection={getStatusIcon(download.status)}>
                          {download.status}
                        </Badge>
                        {download.quality && (
                          <Badge size="sm" variant="outline">{download.quality}</Badge>
                        )}
                      </Group>
                    </Stack>
                  </Group>
                  
                  <Stack spacing="xs" align="flex-end">
                    <Group spacing={4}>
                      {download.status === 'downloading' && (
                        <Tooltip label="Pause">
                          <ActionIcon size="sm" onClick={() => pauseDownload(download.id)}>
                            <IconPlayerPause size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'paused' && (
                        <Tooltip label="Resume">
                          <ActionIcon size="sm" onClick={() => resumeDownload(download.id)}>
                            <IconPlayerPlay size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'error' && (
                        <Tooltip label="Retry">
                          <ActionIcon size="sm" onClick={() => retryDownload(download.id)}>
                            <IconRefresh size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      <Tooltip label="Remove">
                        <ActionIcon size="sm" color="red" onClick={() => removeDownload(download.id)}>
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Tooltip>
                    </Group>
                    
                    {download.status === 'downloading' && (
                      <Progress
                        value={download.progress}
                        size="sm"
                        w={100}
                        color="yted"
                      />
                    )}
                    <Text size="xs" c="dimmed">{Math.round(download.progress)}%</Text>
                  </Stack>
                </Group>
                
                {download.error_message && (
                  <Alert color="red" mt="sm" py="xs">
                    <Text size="xs">{download.error_message}</Text>
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
function IconVideo({ size }: { size: number }) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
      <rect x="2" y="2" width="20" height="20" rx="2" />
      <polygon points="10 8 16 12 10 16" fill="currentColor" />
    </svg>
  );
}
