import {
  Alert,
  Button,
  Card,
  Grid,
  Group,
  Loader,
  Paper,
  SegmentedControl,
  Select,
  Stack,
  Tabs,
  Text,
  Title,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconAlertCircle,
  IconCrop,
  IconDownload,
  IconHistory,
  IconPhoto,
  IconPlayerPlay,
  IconScissors,
  IconSettings,
  IconTool,
  IconVideo,
  IconWand,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import { EventsOn } from '../../wailsjs/runtime';
import { FFmpegInstallerModal } from '../components/FFmpegInstallerModal';
import { ConvertTool } from '../components/editor/ConvertTool';
import { CropTool } from '../components/editor/CropTool';
import { EffectsTool } from '../components/editor/EffectsTool';
import { WatermarkTool } from '../components/editor/WatermarkTool';
import { VideoPlayer } from '../components/VideoPlayer';
import { useEditorStore, useLibraryStore } from '../stores';
import { EditOperation } from '../types/editor';

export function EditorPage() {
  const { colorScheme } = useMantineColorScheme();
  const isDark = colorScheme === 'dark';
  const [showFFmpegModal, setShowFFmpegModal] = useState(false);
  const [ffmpegReady, setFfmpegReady] = useState(false);

  const { videos } = useLibraryStore();
  const {
    ffmpegStatus,
    checkFFmpeg,
    selectedVideoId,
    selectVideo,
    videoMetadata,
    isLoadingMetadata,
    currentOperation,
    setOperation,
    settings,
    updateSettings,
    previewFrame,
    isGeneratingPreview,
    jobs,
    isSubmitting,
    activeTab,
    setActiveTab,
    submitJob,
    reset,
    loadJobs,
  } = useEditorStore();

  useEffect(() => {
    checkFFmpeg();
    // Don't reset on unmount - we want to preserve state when switching tabs
  }, []);

  // Listen for editor events - separate effect to avoid recreating listeners unnecessarily
  useEffect(() => {
    if (!selectedVideoId) return;

    // Listen for editor events
    const cancelJobProgress = EventsOn('editor:job_progress', () => {
      loadJobs(selectedVideoId);
    });

    const cancelJobCompleted = EventsOn('editor:job_completed', () => {
      loadJobs(selectedVideoId);
    });

    const cancelJobFailed = EventsOn('editor:job_failed', () => {
      loadJobs(selectedVideoId);
    });

    return () => {
      cancelJobProgress();
      cancelJobCompleted();
      cancelJobFailed();
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [selectedVideoId]);

  useEffect(() => {
    if (ffmpegStatus && !ffmpegStatus.installed) {
      setShowFFmpegModal(true);
      setFfmpegReady(false);
    } else if (ffmpegStatus?.installed) {
      setFfmpegReady(true);
    }
  }, [ffmpegStatus]);

  const handleVideoSelect = (videoId: string | null) => {
    selectVideo(videoId);
  };

  const handleOperationSelect = (operation: EditOperation) => {
    setOperation(operation);
  };

  const handleSubmit = async () => {
    if (!selectedVideoId || !currentOperation) return;
    await submitJob(settings.replaceOriginal);
  };

  const selectedVideo = videos.find(v => v.id === selectedVideoId);

  const operationTools = [
    { value: 'crop', label: 'Crop & Trim', icon: IconScissors },
    { value: 'watermark', label: 'Watermark', icon: IconPhoto },
    { value: 'convert', label: 'Convert', icon: IconDownload },
    { value: 'effects', label: 'Effects', icon: IconWand },
  ];

  const renderToolPanel = () => {
    switch (currentOperation) {
      case 'crop':
        return (
          <CropTool
            settings={settings}
            onChange={updateSettings}
            videoDuration={videoMetadata?.duration || 0}
            videoWidth={videoMetadata?.width || 0}
            videoHeight={videoMetadata?.height || 0}
          />
        );
      case 'watermark':
        return <WatermarkTool settings={settings} onChange={updateSettings} />;
      case 'convert':
        return <ConvertTool settings={settings} onChange={updateSettings} />;
      case 'effects':
        return <EffectsTool settings={settings} onChange={updateSettings} />;
      default:
        return (
          <Paper p="xl" withBorder>
            <Stack align="center" gap="md">
              <IconTool size={48} color="var(--mantine-color-gray-5)" />
              <Text c="dimmed" ta="center">
                Select an operation from the toolbar to start editing
              </Text>
            </Stack>
          </Paper>
        );
    }
  };

  return (
    <Stack gap="lg" h="100%">
      <Group justify="space-between">
        <Title fw={700} size="xl">
          Video Editor
        </Title>
        {!ffmpegReady && (
          <Alert color="red" icon={<IconAlertCircle size={16} />}>
            <Group gap="xs">
              <Text size="sm">FFmpeg not installed</Text>
              <Button size="xs" variant="light" onClick={() => setShowFFmpegModal(true)}>
                Install Now
              </Button>
            </Group>
          </Alert>
        )}
      </Group>

      <Grid gutter="md" style={{ flex: 1 }}>
        <Grid.Col span={4}>
          <Stack gap="md">
            <Paper withBorder p="md">
              <Stack gap="sm">
                <Text fw={600}>Select Video</Text>
                <Select
                  placeholder="Choose a video from library"
                  value={selectedVideoId}
                  onChange={handleVideoSelect}
                  data={videos.map(v => ({
                    value: v.id,
                    label: v.title,
                  }))}
                  leftSection={<IconVideo size={16} />}
                />
              </Stack>
            </Paper>

            {selectedVideo && (
              <>
                <Paper withBorder p="md">
                  <Stack gap="sm">
                    <Text fw={600}>Operation</Text>
                    <SegmentedControl
                      value={currentOperation || ''}
                      onChange={val => handleOperationSelect(val as EditOperation)}
                      data={operationTools.map(tool => ({
                        value: tool.value,
                        label: (
                          <Group gap={6} wrap="nowrap">
                            <tool.icon size={14} />
                            <span>{tool.label}</span>
                          </Group>
                        ),
                      }))}
                      orientation="vertical"
                      fullWidth
                    />
                  </Stack>
                </Paper>

                <Paper withBorder p="md" style={{ flex: 1 }}>
                  {renderToolPanel()}
                </Paper>

                {currentOperation && (
                  <Button
                    fullWidth
                    size="lg"
                    leftSection={<IconPlayerPlay size={20} />}
                    loading={isSubmitting}
                    disabled={!ffmpegReady}
                    onClick={handleSubmit}
                  >
                    Process Video
                  </Button>
                )}
              </>
            )}
          </Stack>
        </Grid.Col>

        <Grid.Col span={8}>
          <Paper withBorder p="md" h="100%">
            {selectedVideo ? (
              <Stack gap="md" h="100%">
                <Tabs value={activeTab} onChange={val => setActiveTab(val as typeof activeTab)}>
                  <Tabs.List>
                    <Tabs.Tab value="preview" leftSection={<IconPlayerPlay size={14} />}>
                      Preview
                    </Tabs.Tab>
                    <Tabs.Tab value="history" leftSection={<IconHistory size={14} />}>
                      History ({jobs.length})
                    </Tabs.Tab>
                  </Tabs.List>

                  <Tabs.Panel value="preview" pt="md" style={{ flex: 1 }}>
                    {isLoadingMetadata ? (
                      <Stack align="center" justify="center" h={400}>
                        <Loader />
                        <Text c="dimmed">Loading video metadata...</Text>
                      </Stack>
                    ) : (
                      <VideoPlayer
                        videoId={selectedVideo.id}
                        previewFrame={previewFrame}
                        isGeneratingPreview={isGeneratingPreview}
                        metadata={videoMetadata}
                      />
                    )}
                  </Tabs.Panel>

                  <Tabs.Panel value="history" pt="md">
                    <Stack gap="sm">
                      {jobs.length === 0 ? (
                        <Text c="dimmed" ta="center" py="xl">
                          No edit jobs yet
                        </Text>
                      ) : (
                        jobs.map(job => (
                          <Card key={job.id} withBorder padding="sm">
                            <Group justify="space-between">
                              <Stack gap={4}>
                                <Text size="sm" fw={500} tt="capitalize">
                                  {job.operation}
                                </Text>
                                <Text size="xs" c="dimmed">
                                  {new Date(job.createdAt * 1000).toLocaleString()}
                                </Text>
                              </Stack>
                              <Text
                                size="sm"
                                c={
                                  job.status === 'completed'
                                    ? 'green'
                                    : job.status === 'error'
                                      ? 'red'
                                      : job.status === 'processing'
                                        ? 'blue'
                                        : 'gray'
                                }
                                tt="capitalize"
                              >
                                {job.status}
                              </Text>
                            </Group>
                            {job.status === 'processing' && (
                              <Text size="xs" c="dimmed" mt={4}>
                                {Math.round(job.progress)}%
                              </Text>
                            )}
                          </Card>
                        ))
                      )}
                    </Stack>
                  </Tabs.Panel>
                </Tabs>
              </Stack>
            ) : (
              <Stack align="center" justify="center" h="100%">
                <IconVideo size={64} color="var(--mantine-color-gray-5)" />
                <Text c="dimmed" ta="center" size="lg">
                  Select a video from your library to start editing
                </Text>
              </Stack>
            )}
          </Paper>
        </Grid.Col>
      </Grid>

      <FFmpegInstallerModal
        opened={showFFmpegModal}
        onClose={() => setShowFFmpegModal(false)}
        onInstalled={() => {
          setFfmpegReady(true);
          setShowFFmpegModal(false);
        }}
      />
    </Stack>
  );
}
