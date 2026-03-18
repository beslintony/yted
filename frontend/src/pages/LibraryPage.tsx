import {
  ActionIcon,
  Badge,
  Group,
  Image,
  Paper,
  Select,
  SimpleGrid,
  Stack,
  Text,
  TextInput,
  Tooltip,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconFolder,
  IconGridDots,
  IconList,
  IconPlayerPlay,
  IconRefresh,
  IconSearch,
  IconSortAscending,
  IconSortDescending,
  IconTrash,
} from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';

import {
  DeleteVideo,
  GetLibraryStats,
  ListVideos,
  OpenFile,
  OpenFolder,
} from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime';
import { useLibraryStore, useNotifications } from '../stores';
import { Video } from '../types';

// Wrapper functions with error handling
const handleOpenFile = async (
  filePath: string,
  error: (title: string, message: string) => void
) => {
  if (!filePath) {
    error('Cannot Open File', 'File path is empty');
    return;
  }
  try {
    await OpenFile(filePath);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    error('Cannot Open File', err?.message || 'Failed to open file');
  }
};

const handleOpenFolder = async (
  filePath: string,
  error: (title: string, message: string) => void
) => {
  if (!filePath) {
    error('Cannot Open Folder', 'File path is empty');
    return;
  }
  try {
    await OpenFolder(filePath);
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    error('Cannot Open Folder', err?.message || 'Failed to open folder');
  }
};

// Helper function to map backend VideoResult to frontend Video type
function mapVideoResultToVideo(v: app.VideoResult): Video {
  return {
    id: v.id,
    youtubeId: v.youtube_id,
    title: v.title,
    channel: v.channel,
    channelId: v.channel_id,
    duration: v.duration,
    description: v.description,
    thumbnailUrl: v.thumbnail_url,
    filePath: v.file_path,
    fileSize: v.file_size,
    format: v.format,
    quality: v.quality,
    downloadedAt: v.downloaded_at,
    watchPosition: v.watch_position,
    watchCount: v.watch_count,
  };
}

export function LibraryPage() {
  const [videos, setVideos] = useState<Video[]>([]);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState('date');
  const [sortDesc, setSortDesc] = useState(true);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [stats, setStats] = useState({ totalVideos: 0, totalSize: 0 });

  const { setVideos: setStoreVideos, removeVideo: removeStoreVideo } = useLibraryStore();
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';
  const { success, error, confirm } = useNotifications();

  const loadVideos = useCallback(async () => {
    try {
      const result = await ListVideos({
        search,
        channel: '',
        sort_by: sortBy,
        sort_desc: sortDesc,
        limit: 100,
        offset: 0,
      });

      // Map backend VideoResult to frontend Video type
      const mappedVideos = (result || []).map(mapVideoResultToVideo);
      setVideos(mappedVideos);
      setStoreVideos(mappedVideos);
    } catch (err) {
      error('Failed to load videos', err instanceof Error ? err.message : 'Unknown error');
    }
  }, [search, sortBy, sortDesc, setStoreVideos, error]);

  const loadStats = useCallback(async () => {
    try {
      const result = (await GetLibraryStats()) as { total_videos: number; total_size: number };
      setStats({ totalVideos: result.total_videos, totalSize: result.total_size });
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  }, []);

  useEffect(() => {
    loadVideos();
    loadStats();

    // Listen for library updates (when a new download completes)
    const cancelLibraryUpdate = EventsOn('library:updated', () => {
      loadVideos();
      loadStats();
    });

    return () => {
      cancelLibraryUpdate();
    };
  }, [loadVideos, loadStats]);

  const handleDelete = async (video: Video) => {
    confirm({
      title: 'Delete Video',
      message: `Are you sure you want to delete "${video.title}"?`,
      confirmLabel: 'Delete',
      cancelLabel: 'Cancel',
      confirmColor: 'red',
      onConfirm: async () => {
        try {
          await DeleteVideo(video.id, true);
          removeStoreVideo(video.id);
          setVideos(prev => prev.filter(v => v.id !== video.id));
          success('Video deleted', `"${video.title}" has been deleted`);
        } catch (err) {
          error('Failed to delete video', err instanceof Error ? err.message : 'Unknown error');
        }
      },
    });
  };

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${(bytes / k ** i).toFixed(1)} ${sizes[i]}`;
  };

  const formatDuration = (seconds: number) => {
    if (!seconds) return 'Unknown';
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const filteredVideos = videos.filter(
    video =>
      video.title.toLowerCase().includes(search.toLowerCase()) ||
      video.channel.toLowerCase().includes(search.toLowerCase())
  );

  return (
    <Stack gap="md">
      <Group justify="space-between">
        <Text fw={700} size="xl">
          Library ({stats.totalVideos} videos, {formatFileSize(stats.totalSize)})
        </Text>
        <Group gap="sm">
          <TextInput
            leftSection={<IconSearch size={16} />}
            placeholder="Search videos..."
            value={search}
            onChange={e => setSearch(e.currentTarget.value)}
          />
          <Select
            data={[
              { value: 'date', label: 'Date' },
              { value: 'title', label: 'Title' },
              { value: 'channel', label: 'Channel' },
              { value: 'duration', label: 'Duration' },
            ]}
            value={sortBy}
            onChange={value => value && setSortBy(value)}
          />
          <ActionIcon variant="light" onClick={() => setSortDesc(!sortDesc)}>
            {sortDesc ? <IconSortDescending size={18} /> : <IconSortAscending size={18} />}
          </ActionIcon>
          <ActionIcon
            variant="light"
            onClick={() => setViewMode(viewMode === 'grid' ? 'list' : 'grid')}
          >
            {viewMode === 'grid' ? <IconList size={18} /> : <IconGridDots size={18} />}
          </ActionIcon>
          <ActionIcon variant="light" onClick={loadVideos}>
            <IconRefresh size={18} />
          </ActionIcon>
        </Group>
      </Group>

      {filteredVideos.length === 0 ? (
        <Paper withBorder p="xl">
          <Text c="dimmed" ta="center">
            No videos found. Download some videos to see them here.
          </Text>
        </Paper>
      ) : viewMode === 'grid' ? (
        <SimpleGrid cols={3} spacing="md">
          {filteredVideos.map(video => (
            <Paper key={video.id} withBorder p="md">
              <Stack gap="sm">
                {video.thumbnailUrl ? (
                  <Image
                    alt={video.title}
                    height={120}
                    radius="sm"
                    src={video.thumbnailUrl}
                    style={{ cursor: 'pointer', objectFit: 'cover' }}
                    onClick={() => handleOpenFile(video.filePath, error)}
                  />
                ) : (
                  <Paper
                    bg={dark ? '#2c2e33' : '#e9ecef'}
                    h={120}
                    style={{
                      cursor: 'pointer',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                    }}
                    onClick={() => handleOpenFile(video.filePath, error)}
                  >
                    <IconPlayerPlay color={dark ? '#5c5f66' : '#adb5bd'} size={32} />
                  </Paper>
                )}
                <div>
                  <Text fw={600} lineClamp={2}>
                    {video.title}
                  </Text>
                  <Text c="dimmed" size="sm">
                    {video.channel}
                  </Text>
                  <Group gap="xs" mt="xs">
                    <Badge size="sm" variant="light">
                      {formatDuration(video.duration)}
                    </Badge>
                    <Badge size="sm" variant="light">
                      {video.quality}
                    </Badge>
                    <Badge size="sm" variant="light">
                      {formatFileSize(video.fileSize)}
                    </Badge>
                  </Group>
                </div>
                <Group gap="xs" justify="flex-end">
                  <Tooltip label="Open file">
                    <ActionIcon
                      color="blue"
                      variant="light"
                      onClick={() => handleOpenFile(video.filePath, error)}
                    >
                      <IconPlayerPlay size={16} />
                    </ActionIcon>
                  </Tooltip>
                  <Tooltip label="Open folder">
                    <ActionIcon
                      color="gray"
                      variant="light"
                      onClick={() => handleOpenFolder(video.filePath, error)}
                    >
                      <IconFolder size={16} />
                    </ActionIcon>
                  </Tooltip>
                  <Tooltip label="Delete">
                    <ActionIcon color="red" variant="light" onClick={() => handleDelete(video)}>
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Tooltip>
                </Group>
              </Stack>
            </Paper>
          ))}
        </SimpleGrid>
      ) : (
        <Stack gap="sm">
          {filteredVideos.map(video => (
            <Paper key={video.id} withBorder p="sm">
              <Group justify="space-between">
                <Group gap="sm">
                  {video.thumbnailUrl ? (
                    <Image
                      alt={video.title}
                      height={60}
                      radius="sm"
                      src={video.thumbnailUrl}
                      style={{ cursor: 'pointer', objectFit: 'cover', width: 80 }}
                      onClick={() => handleOpenFile(video.filePath, error)}
                    />
                  ) : (
                    <Paper
                      bg={dark ? '#2c2e33' : '#e9ecef'}
                      h={60}
                      style={{
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        width: 80,
                      }}
                      onClick={() => handleOpenFile(video.filePath, error)}
                    >
                      <IconPlayerPlay color={dark ? '#5c5f66' : '#adb5bd'} size={24} />
                    </Paper>
                  )}
                  <div>
                    <Text fw={600} lineClamp={1}>
                      {video.title}
                    </Text>
                    <Text c="dimmed" size="sm">
                      {video.channel}
                    </Text>
                    <Group gap="xs">
                      <Badge size="sm" variant="light">
                        {formatDuration(video.duration)}
                      </Badge>
                      <Badge size="sm" variant="light">
                        {video.quality}
                      </Badge>
                      <Badge size="sm" variant="light">
                        {formatFileSize(video.fileSize)}
                      </Badge>
                    </Group>
                  </div>
                </Group>
                <Group gap="xs">
                  <Tooltip label="Open file">
                    <ActionIcon
                      color="blue"
                      variant="light"
                      onClick={() => handleOpenFile(video.filePath, error)}
                    >
                      <IconPlayerPlay size={16} />
                    </ActionIcon>
                  </Tooltip>
                  <Tooltip label="Open folder">
                    <ActionIcon
                      color="gray"
                      variant="light"
                      onClick={() => handleOpenFolder(video.filePath, error)}
                    >
                      <IconFolder size={16} />
                    </ActionIcon>
                  </Tooltip>
                  <Tooltip label="Delete">
                    <ActionIcon color="red" variant="light" onClick={() => handleDelete(video)}>
                      <IconTrash size={16} />
                    </ActionIcon>
                  </Tooltip>
                </Group>
              </Group>
            </Paper>
          ))}
        </Stack>
      )}
    </Stack>
  );
}
