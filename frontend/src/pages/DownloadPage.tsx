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
  Select,
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
  IconVideo,
  IconGauge,
} from '@tabler/icons-react';
import { useCallback, useEffect, useRef, useState } from 'react';

import { GetVideoInfo, AddDownload, ValidateURL, GetSettings, GetDownloadQueue, StartProcessingDownloads, RetryDownload } from '../../wailsjs/go/app/App';
import { app, config } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime';
import { useDownloadStore, useSettingsStore, useNotifications } from '../stores';
import { VideoFormat } from '../types';

export function DownloadPage() {
  const [url, setUrl] = useState('');
  const [loading, setLoading] = useState(false);
  const [videoInfo, setVideoInfo] = useState<app.VideoInfoResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [adding, setAdding] = useState(false);
  const [presets, setPresets] = useState<config.DownloadPreset[]>([]);
  const [selectedPreset, setSelectedPreset] = useState<string>('');
  const [settingsLoaded, setSettingsLoaded] = useState(false);
  const [queueRestored, setQueueRestored] = useState(false);

  const {
    downloads,
    addDownload,
    removeDownload,
    pauseDownload,
    resumeDownload,
    retryDownload,
    updateProgress,
    updateDownloadInfo,
    startDownload,
    completeDownload,
    failDownload,
    hasDownload,
  } = useDownloadStore();

  const { defaultQuality } = useSettingsStore();
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';
  const { success, error: showError, warning } = useNotifications();
  
  // Use refs to track processed events to prevent duplicates
  const processedEvents = useRef<Set<string>>(new Set());

  // Wrapper to remove download and clear its processed events
  const removeDownloadWithCleanup = useCallback((id: string) => {
    processedEvents.current.delete(`completed_${id}`);
    processedEvents.current.delete(`error_${id}`);
    removeDownload(id);
  }, [removeDownload]);

  // Load presets from settings
  useEffect(() => {
    const loadPresets = async () => {
      try {
        const settings = await GetSettings();
        if (settings && settings.download_presets) {
          setPresets(settings.download_presets);
          // Set default preset based on defaultQuality
          const defaultPreset = settings.download_presets.find(
            (p: config.DownloadPreset) => p.quality === defaultQuality
          );
          if (defaultPreset) {
            setSelectedPreset(defaultPreset.id);
          } else if (settings.download_presets.length > 0) {
            setSelectedPreset(settings.download_presets[0].id);
          }
        }
        setSettingsLoaded(true);
      } catch (err) {
        console.error('Failed to load presets:', err);
        setSettingsLoaded(true);
      }
    };
    loadPresets();
  }, [defaultQuality]);

  // Restore download queue from previous session
  useEffect(() => {
    if (queueRestored) return;
    
    const restoreQueue = async () => {
      try {
        const queue = await GetDownloadQueue();
        
        if (queue && queue.length > 0) {
          for (const data of queue) {
            if (!data?.id || !data?.url) continue;

            const backendStatus = data.status;
            
            if (!hasDownload(data.id)) {
              // Add restored download to the store with correct metadata
              addDownload(data.url, {
                id: data.youtube_id || data.id || '',
                title: data.title || 'Restored Download',
                channel: data.channel || '',
                channelId: '',
                duration: 0,
                description: '',
                thumbnail: data.thumbnail_url || '',
                formats: [],
              }, {
                formatId: data.format_id || 'best',
                quality: data.quality || 'best',
                ext: 'mp4',
                resolution: '',
                fps: 0,
                vcodec: '',
                acodec: '',
                filesize: 0,
              }, data.id);
            }

            // Always sync status and progress from backend (fixes missed events)
            updateProgress(data.id, data.progress || 0);

            if (backendStatus === 'downloading') {
              startDownload(data.id);
            } else if (backendStatus === 'completed') {
              completeDownload(data.id);
            } else if (backendStatus === 'error') {
              failDownload(data.id, data.error_message || 'Unknown error');
            }
          }
          // Tell backend to start processing pending downloads
          await StartProcessingDownloads();
        }
        setQueueRestored(true);
      } catch (err) {
        console.error('Failed to restore download queue:', err);
        setQueueRestored(true);
      }
    };
    restoreQueue();
  }, [queueRestored, addDownload, updateProgress, startDownload, completeDownload, failDownload, hasDownload]);

  // Listen for download progress events from backend
  useEffect(() => {
    const cancelProgress = EventsOn('download:progress', (data: { id?: string; progress?: number; speed?: string; eta?: string; size?: string; is_throttled?: boolean; speed_limit?: string }) => {
      if (data?.id && typeof data.progress === 'number') {
        updateProgress(data.id, data.progress);
        if (data.speed || data.eta || data.size || data.is_throttled !== undefined) {
          updateDownloadInfo(data.id, {
            speed: data.speed,
            eta: data.eta,
            size: data.size,
            isThrottled: data.is_throttled,
            speedLimit: data.speed_limit,
          });
        }
      }
    });

    const cancelCompleted = EventsOn('download:completed', (data: unknown) => {
      const completedId = typeof data === 'string' ? data : (data as { id?: string })?.id;
      if (completedId && !processedEvents.current.has(`completed_${completedId}`)) {
        processedEvents.current.add(`completed_${completedId}`);
        completeDownload(completedId);
        const completedDownload = downloads.find(d => d.id === completedId);
        if (completedDownload) {
          success('Download Complete', `"${completedDownload.title}" has finished downloading`);
        }
      }
    });

    const cancelError = EventsOn('download:error', (data: { id?: string; error?: string }) => {
      if (data?.id && data?.error && !processedEvents.current.has(`error_${data.id}`)) {
        processedEvents.current.add(`error_${data.id}`);
        failDownload(data.id, data.error);
        const failedDownload = downloads.find(d => d.id === data.id);
        if (failedDownload) {
          showError('Download Failed', `"${failedDownload.title}" failed: ${data.error}`);
        }
      }
    });

    const cancelStarted = EventsOn('download:started', (data: { id?: string }) => {
      if (data?.id) {
        startDownload(data.id);
      }
    });

    const cancelRetried = EventsOn('download:retried', (retryId: string) => {
      if (retryId) {
        // Clear processed events for this download so new events can be handled
        processedEvents.current.delete(`completed_${retryId}`);
        processedEvents.current.delete(`error_${retryId}`);
        retryDownload(retryId);
      }
    });

    // Cleanup old processed events periodically
    const cleanupInterval = setInterval(() => {
      if (processedEvents.current.size > 1000) {
        processedEvents.current.clear();
      }
    }, 60000);

    return () => {
      cancelProgress();
      cancelCompleted();
      cancelError();
      cancelStarted();
      cancelRetried();
      clearInterval(cleanupInterval);
    };
  }, [updateProgress, completeDownload, failDownload, startDownload, retryDownload, downloads, success, showError, updateDownloadInfo]);

  const handleFetchInfo = async () => {
    if (!url.trim()) {
      setError('Please enter a YouTube URL');
      warning('Missing URL', 'Please enter a YouTube URL');
      return;
    }

    setLoading(true);
    setError(null);
    setVideoInfo(null);

    try {
      const isValid = await ValidateURL(url);
      if (!isValid) {
        setError('Invalid YouTube URL. Please enter a valid YouTube video URL.');
        showError('Invalid URL', 'Please enter a valid YouTube video URL');
        return;
      }

      const info = await GetVideoInfo(url);
      if (!info || !info.id) {
        setError('Could not fetch video information. Please check the URL and try again.');
        showError('Fetch Failed', 'Could not fetch video information. Please check the URL and try again.');
        return;
      }
      setVideoInfo(info);
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to fetch video info';
      setError(errorMessage);
      showError('Fetch Failed', errorMessage);
    } finally {
      setLoading(false);
    }
  };

  const handleDownload = async () => {
    if (!videoInfo || !url) {
      setError('No video info available');
      showError('No Video Info', 'Please fetch video information first');
      return;
    }

    setAdding(true);
    setError(null);

    try {
      const preset = presets.find(p => p.id === selectedPreset);
      const formatId = preset?.format || 'best';
      const quality = preset?.quality || 'best';

      const id = await AddDownload(url, formatId, quality);

      if (id) {
        addDownload(url, {
          id: videoInfo.id,
          title: videoInfo.title,
          channel: videoInfo.channel,
          channelId: videoInfo.channel_id,
          duration: videoInfo.duration,
          description: videoInfo.description,
          thumbnail: videoInfo.thumbnail,
          formats: videoInfo.formats as unknown as VideoFormat[],
        }, {
          formatId: preset?.format || 'best',
          ext: preset?.extension || 'mp4',
          resolution: '',
          fps: 0,
          vcodec: '',
          acodec: '',
          filesize: 0,
          quality: preset?.quality || 'best',
        }, id);

        success('Download Added', `"${videoInfo.title}" has been added to the queue`);
        setUrl('');
        setVideoInfo(null);
      }
    } catch (err: unknown) {
      const errorMessage = err instanceof Error ? err.message : 'Failed to add download';
      setError(errorMessage);
      showError('Download Failed', errorMessage);
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

  const formatETA = (eta: string) => eta || '';

  return (
    <Stack gap="lg">
      <Text c={dark ? '#fff' : '#000'} fw={700} size="xl">Downloads</Text>

      {/* URL Input */}
      <Paper
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        p="md"
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <TextInput
            disabled={loading}
            leftSection={<IconLink size={16} />}
            placeholder="Paste YouTube URL here..."
            rightSection={
              url ? (
                <Tooltip label="Clear">
                  <ActionIcon color="gray" variant="subtle" onClick={handleClear}>
                    <IconX size={16} />
                  </ActionIcon>
                </Tooltip>
              ) : undefined
            }
            size="md"
            styles={{
              input: {
                background: dark ? '#141517' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={url}
            onChange={(e) => setUrl(e.currentTarget.value)}
            onKeyDown={(e) => e.key === 'Enter' && handleFetchInfo()}
          />

          <Group justify="flex-end">
            <Tooltip label="Get video information">
              <Button
                color="yted"
                disabled={!url.trim() || loading}
                leftSection={<IconSearch size={16} />}
                loading={loading}
                onClick={handleFetchInfo}
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
              withBorder
              bg={dark ? '#1a1b1e' : '#f8f9fa'}
              p="md"
              style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
            >
              <Group align="flex-start" wrap="nowrap">
                {videoInfo.thumbnail ? (
                  <img
                    alt={videoInfo.title}
                    src={videoInfo.thumbnail}
                    style={{
                      width: 160,
                      height: 90,
                      objectFit: 'cover',
                      borderRadius: 8,
                    }}
                  />
                ) : (
                  <Paper
                    bg={dark ? '#2c2e33' : '#e9ecef'}
                    h={90}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      borderRadius: 8,
                    }}
                    w={160}
                  >
                    <IconVideo color={dark ? '#5c5f66' : '#adb5bd'} size={32} />
                  </Paper>
                )}
                <Stack gap="xs" style={{ flex: 1 }}>
                  <Text c={dark ? '#fff' : '#000'} fw={600} lineClamp={2}>
                    {videoInfo.title || 'Unknown Title'}
                  </Text>
                  <Text c={dark ? 'dimmed' : 'gray.7'} size="sm">
                    {videoInfo.channel || 'Unknown Channel'}
                  </Text>
                  <Text c={dark ? 'dimmed' : 'gray.7'} size="sm">
                    Duration: {formatDuration(videoInfo.duration)}
                  </Text>

                  {settingsLoaded && presets.length > 0 && (
                    <Select
                      data={presets.map(p => ({
                        value: p.id,
                        label: `${p.name} (${p.quality}, .${p.extension})`,
                      }))}
                      description="Choose quality and format"
                      label="Download Preset"
                      size="sm"
                      styles={{
                        input: {
                          background: dark ? '#141517' : '#f8f9fa',
                          color: dark ? '#c1c2c5' : '#212529',
                        },
                      }}
                      value={selectedPreset}
                      w={250}
                      onChange={(value) => value && setSelectedPreset(value)}
                    />
                  )}

                  <Group justify="flex-end" mt="xs">
                    <Tooltip label="Add to download queue">
                      <Button
                        color="yted"
                        disabled={!selectedPreset && presets.length > 0}
                        leftSection={<IconDownload size={16} />}
                        loading={adding}
                        size="sm"
                        onClick={handleDownload}
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
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          Queue ({downloads.length})
        </Text>
        {downloads.length > 0 && (
          <Tooltip label="Clear all downloads">
            <Button
              color="gray"
              size="xs"
              variant="subtle"
              onClick={() => downloads.forEach(d => removeDownloadWithCleanup(d.id))}
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
              withBorder
              bg={dark ? '#25262b' : '#fff'}
              p="xl"
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
                withBorder
                bg={dark ? '#25262b' : '#fff'}
                p="sm"
                style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
              >
                <Group align="flex-start" justify="space-between">
                  <Group align="flex-start" gap="sm" wrap="nowrap">
                    {download.thumbnail ? (
                      <img
                        alt={download.title || 'Video'}
                        src={download.thumbnail}
                        style={{
                          width: 80,
                          height: 45,
                          objectFit: 'cover',
                          borderRadius: 4,
                        }}
                      />
                    ) : (
                      <Paper
                        bg={dark ? '#2c2e33' : '#e9ecef'}
                        h={45}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          borderRadius: 4,
                        }}
                        w={80}
                      >
                        <IconVideo color={dark ? '#5c5f66' : '#adb5bd'} size={20} />
                      </Paper>
                    )}
                    <Stack gap={4}>
                      <Text c={dark ? '#fff' : '#000'} fw={500} lineClamp={1} size="sm" style={{ maxWidth: 300 }}>
                        {download.title || 'Loading...'}
                      </Text>
                      <Text c={dark ? 'dimmed' : 'gray.6'} size="xs">
                        {download.channel || download.url}
                      </Text>
                      <Group gap="xs">
                        <Badge
                          color={getStatusColor(download.status)}
                          leftSection={
                            download.status === 'completed' ? <IconCheck size={12} /> :
                            download.status === 'downloading' ? <Loader size={12} /> :
                            download.status === 'error' ? <IconAlertCircle size={12} /> :
                            download.status === 'paused' ? <IconPlayerPause size={12} /> :
                            <IconDownload size={12} />
                          }
                          size="sm"
                        >
                          {download.status}
                        </Badge>
                        {download.quality && (
                          <Badge color={dark ? 'gray' : 'dark'} size="sm" variant="outline">
                            {download.quality}
                          </Badge>
                        )}
                      </Group>
                    </Stack>
                  </Group>

                  <Stack align="flex-end" gap="xs">
                    <Group gap={4}>
                      {download.status === 'downloading' && (
                        <Tooltip label="Pause download">
                          <ActionIcon
                            color="yellow"
                            size="sm"
                            variant="light"
                            onClick={() => pauseDownload(download.id)}
                          >
                            <IconPlayerPause size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'paused' && (
                        <Tooltip label="Resume download">
                          <ActionIcon
                            color="green"
                            size="sm"
                            variant="light"
                            onClick={() => resumeDownload(download.id)}
                          >
                            <IconPlayerPlay size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      {download.status === 'error' && (
                        <Tooltip label="Retry download">
                          <ActionIcon
                            color="blue"
                            size="sm"
                            variant="light"
                            onClick={async () => {
                              try {
                                await RetryDownload(download.id);
                                retryDownload(download.id);
                              } catch (err) {
                                console.error('Failed to retry download:', err);
                                showError('Retry Failed', 'Failed to restart download');
                              }
                            }}
                          >
                            <IconRefresh size={14} />
                          </ActionIcon>
                        </Tooltip>
                      )}
                      <Tooltip label="Remove from queue">
                        <ActionIcon
                          color="red"
                          size="sm"
                          variant="light"
                          onClick={() => removeDownloadWithCleanup(download.id)}
                        >
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Tooltip>
                    </Group>

                    {download.status === 'downloading' && (
                      <Tooltip label={`${Math.round(download.progress)}% complete${download.speed ? ` • ${download.speed}` : ''}${download.eta ? ` • ${formatETA(download.eta)} left` : ''}`}>
                        <Progress
                          color="yted"
                          radius="xs"
                          size="sm"
                          value={download.progress}
                          w={100}
                        />
                      </Tooltip>
                    )}
                    <Group gap={8}>
                      <Text c={dark ? 'dimmed' : 'gray.6'} size="xs">
                        {Math.round(download.progress)}%
                      </Text>
                      {download.status === 'downloading' && download.speed && (
                        <Group gap={4}>
                          <Text c={dark ? 'dimmed' : 'gray.6'} size="xs">
                            {download.speed}
                          </Text>
                          {download.isThrottled && download.speedLimit && (
                            <Tooltip label={`Speed limited to ${download.speedLimit}`}>
                              <IconGauge size={14} color={dark ? '#909296' : '#868e96'} />
                            </Tooltip>
                          )}
                        </Group>
                      )}
                      {download.status === 'downloading' && download.eta && formatETA(download.eta) && (
                        <Text c={dark ? 'dimmed' : 'gray.6'} size="xs">
                          {formatETA(download.eta)} left
                        </Text>
                      )}
                    </Group>
                  </Stack>
                </Group>

                {download.errorMessage && (
                  <Alert bg={dark ? '#2c1b1b' : '#fff5f5'} color="red" mt="sm" p="xs">
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
