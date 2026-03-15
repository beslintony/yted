import { useState, useEffect } from 'react';
import {
  TextInput,
  Group,
  Stack,
  Paper,
  Text,
  Badge,
  ActionIcon,
  Tooltip,
  Grid,
  Select,
  Button,
  Image,
  SimpleGrid,
  Menu,
} from '@mantine/core';
import {
  IconSearch,
  IconPlayerPlay,
  IconTrash,
  IconFolder,
  IconGridDots,
  IconList,
  IconSortAscending,
  IconSortDescending,
} from '@tabler/icons-react';
import { useLibraryStore } from '../stores';
import { ListVideos, GetLibraryStats, OpenFolder, DeleteVideo } from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';

export function LibraryPage() {
  const [videos, setVideos] = useState<app.VideoResult[]>([]);
  const [search, setSearch] = useState('');
  const [sortBy, setSortBy] = useState('date');
  const [sortDesc, setSortDesc] = useState(true);
  const [viewMode, setViewMode] = useState<'grid' | 'list'>('grid');
  const [stats, setStats] = useState({ total_videos: 0, total_size: 0 });
  
  const { setVideos: setStoreVideos } = useLibraryStore();

  useEffect(() => {
    loadVideos();
    loadStats();
  }, [search, sortBy, sortDesc]);

  const loadVideos = async () => {
    try {
      const result = await ListVideos({
        search,
        channel: '',
        sort_by: sortBy,
        sort_desc: sortDesc,
        limit: 100,
        offset: 0,
      });
      setVideos(result || []);
      setStoreVideos(result as any || []);
    } catch (err) {
      console.error('Failed to load videos:', err);
    }
  };

  const loadStats = async () => {
    try {
      const result = await GetLibraryStats();
      setStats(result as any);
    } catch (err) {
      console.error('Failed to load stats:', err);
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await DeleteVideo(id);
      loadVideos();
    } catch (err) {
      console.error('Failed to delete video:', err);
    }
  };

  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  const formatFileSize = (bytes: number) => {
    const units = ['B', 'KB', 'MB', 'GB'];
    let size = bytes;
    let unitIndex = 0;
    while (size >= 1024 && unitIndex < units.length - 1) {
      size /= 1024;
      unitIndex++;
    }
    return `${size.toFixed(1)} ${units[unitIndex]}`;
  };

  return (
    <Stack spacing="lg">
      <Group position="apart">
        <Text size="xl" fw={700}>Library</Text>
        <Text size="sm" c="dimmed">
          {stats.total_videos} videos • {formatFileSize(stats.total_size)}
        </Text>
      </Group>

      {/* Filters */}
      <Paper p="sm" withBorder>
        <Group spacing="md">
          <TextInput
            placeholder="Search videos..."
            icon={<IconSearch size={16} />}
            value={search}
            onChange={(e) => setSearch(e.currentTarget.value)}
            style={{ flex: 1 }}
          />
          <Select
            value={sortBy}
            onChange={(v) => v && setSortBy(v)}
            data={[
              { value: 'date', label: 'Date' },
              { value: 'title', label: 'Title' },
              { value: 'channel', label: 'Channel' },
              { value: 'duration', label: 'Duration' },
            ]}
            w={120}
          />
          <ActionIcon
            variant={sortDesc ? 'filled' : 'light'}
            onClick={() => setSortDesc(!sortDesc)}
          >
            {sortDesc ? <IconSortDescending size={18} /> : <IconSortAscending size={18} />}
          </ActionIcon>
          <Group spacing={4}>
            <ActionIcon
              variant={viewMode === 'grid' ? 'filled' : 'light'}
              onClick={() => setViewMode('grid')}
            >
              <IconGridDots size={18} />
            </ActionIcon>
            <ActionIcon
              variant={viewMode === 'list' ? 'filled' : 'light'}
              onClick={() => setViewMode('list')}
            >
              <IconList size={18} />
            </ActionIcon>
          </Group>
        </Group>
      </Paper>

      {/* Videos */}
      {videos.length === 0 ? (
        <Paper p="xl" withBorder>
          <Text c="dimmed" align="center">No videos in library yet.</Text>
        </Paper>
      ) : viewMode === 'grid' ? (
        <SimpleGrid cols={4} spacing="md" breakpoints={[
          { maxWidth: 'lg', cols: 3 },
          { maxWidth: 'md', cols: 2 },
          { maxWidth: 'sm', cols: 1 },
        ]}>
          {videos.map((video) => (
            <VideoCard
              key={video.id}
              video={video}
              onDelete={() => handleDelete(video.id)}
            />
          ))}
        </SimpleGrid>
      ) : (
        <Stack spacing="sm">
          {videos.map((video) => (
            <VideoListItem
              key={video.id}
              video={video}
              onDelete={() => handleDelete(video.id)}
            />
          ))}
        </Stack>
      )}
    </Stack>
  );

  function VideoCard({ video, onDelete }: { video: app.VideoResult; onDelete: () => void }) {
    return (
      <Paper withBorder radius="md" sx={{ overflow: 'hidden' }}>
        <div style={{ position: 'relative' }}>
          <Image
            src={video.thumbnail_url || '/logo.svg'}
            height={140}
            alt={video.title}
            withPlaceholder
          />
          <Badge
            style={{ position: 'absolute', bottom: 8, right: 8 }}
            color="dark"
            variant="filled"
          >
            {formatDuration(video.duration)}
          </Badge>
        </div>
        <Stack p="sm" spacing="xs">
          <Text size="sm" fw={500} lineClamp={2}>{video.title}</Text>
          <Text size="xs" c="dimmed">{video.channel}</Text>
          <Group position="apart">
            <Text size="xs" c="dimmed">{formatFileSize(video.file_size)}</Text>
            <Group spacing={4}>
              <Tooltip label="Play">
                <ActionIcon size="sm">
                  <IconPlayerPlay size={14} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Open Folder">
                <ActionIcon size="sm" onClick={() => OpenFolder(video.file_path)}>
                  <IconFolder size={14} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Delete">
                <ActionIcon size="sm" color="red" onClick={onDelete}>
                  <IconTrash size={14} />
                </ActionIcon>
              </Tooltip>
            </Group>
          </Group>
        </Stack>
      </Paper>
    );
  }

  function VideoListItem({ video, onDelete }: { video: app.VideoResult; onDelete: () => void }) {
    return (
      <Paper p="sm" withBorder>
        <Group position="apart" align="flex-start">
          <Group spacing="sm" align="flex-start" noWrap>
            <Image
              src={video.thumbnail_url || '/logo.svg'}
              width={120}
              height={68}
              radius="sm"
              alt={video.title}
              withPlaceholder
            />
            <Stack spacing={4}>
              <Text fw={500} lineClamp={1}>{video.title}</Text>
              <Text size="sm" c="dimmed">{video.channel}</Text>
              <Group spacing="xs">
                <Badge size="sm">{formatDuration(video.duration)}</Badge>
                <Badge size="sm" variant="outline">{formatFileSize(video.file_size)}</Badge>
                <Badge size="sm" variant="outline">{video.quality}</Badge>
              </Group>
            </Stack>
          </Group>
          <Group spacing={4}>
            <Tooltip label="Play">
              <ActionIcon>
                <IconPlayerPlay size={18} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Open Folder">
              <ActionIcon onClick={() => OpenFolder(video.file_path)}>
                <IconFolder size={18} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Delete">
              <ActionIcon color="red" onClick={onDelete}>
                <IconTrash size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </Paper>
    );
  }
}
