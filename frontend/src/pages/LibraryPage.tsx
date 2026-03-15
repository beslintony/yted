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
  SimpleGrid,
  Select,
  Button,
  Image,
  useMantineColorScheme,
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
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

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
    <Stack gap="lg">
      <Group justify="space-between">
        <Text size="xl" fw={700} c={dark ? '#fff' : '#000'}>Library</Text>
        <Text size="sm" c={dark ? 'dimmed' : 'gray.6'}>
          {stats.total_videos} videos • {formatFileSize(stats.total_size)}
        </Text>
      </Group>

      {/* Filters */}
      <Paper 
        p="sm" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Group gap="md">
          <TextInput
            placeholder="Search videos..."
            leftSection={<IconSearch size={16} />}
            value={search}
            onChange={(e) => setSearch(e.currentTarget.value)}
            style={{ flex: 1 }}
            styles={{
              input: {
                background: dark ? '#141517' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
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
            styles={{
              input: {
                background: dark ? '#141517' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />
          <Tooltip label={sortDesc ? "Descending order" : "Ascending order"}>
            <ActionIcon
              variant={sortDesc ? 'filled' : 'light'}
              onClick={() => setSortDesc(!sortDesc)}
            >
              {sortDesc ? <IconSortDescending size={18} /> : <IconSortAscending size={18} />}
            </ActionIcon>
          </Tooltip>
          <Group gap={4}>
            <Tooltip label="Grid view">
              <ActionIcon
                variant={viewMode === 'grid' ? 'filled' : 'light'}
                onClick={() => setViewMode('grid')}
              >
                <IconGridDots size={18} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="List view">
              <ActionIcon
                variant={viewMode === 'list' ? 'filled' : 'light'}
                onClick={() => setViewMode('list')}
              >
                <IconList size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </Paper>

      {/* Videos */}
      {videos.length === 0 ? (
        <Paper 
          p="xl" 
          withBorder
          bg={dark ? '#25262b' : '#fff'}
          style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
        >
          <Text c={dark ? 'dimmed' : 'gray.6'} ta="center">No videos in library yet.</Text>
        </Paper>
      ) : viewMode === 'grid' ? (
        <SimpleGrid cols={{ base: 1, sm: 2, lg: 3, xl: 4 }} spacing="md">
          {videos.map((video) => (
            <VideoCard
              key={video.id}
              video={video}
              onDelete={() => handleDelete(video.id)}
              dark={dark}
            />
          ))}
        </SimpleGrid>
      ) : (
        <Stack gap="sm">
          {videos.map((video) => (
            <VideoListItem
              key={video.id}
              video={video}
              onDelete={() => handleDelete(video.id)}
              dark={dark}
            />
          ))}
        </Stack>
      )}
    </Stack>
  );

  function VideoCard({ video, onDelete, dark }: { video: app.VideoResult; onDelete: () => void; dark: boolean }) {
    return (
      <Paper 
        withBorder 
        radius="md"
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6', overflow: 'hidden' }}
      >
        <div style={{ position: 'relative' }}>
          <Image
            src={video.thumbnail_url || '/logo.svg'}
            height={140}
            alt={video.title}
            fallbackSrc="/logo.svg"
          />
          <Badge
            style={{ position: 'absolute', bottom: 8, right: 8 }}
            color="dark"
            variant="filled"
          >
            {formatDuration(video.duration)}
          </Badge>
        </div>
        <Stack p="sm" gap="xs">
          <Text size="sm" fw={500} lineClamp={2} c={dark ? '#fff' : '#000'}>{video.title}</Text>
          <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>{video.channel}</Text>
          <Group justify="space-between">
            <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>{formatFileSize(video.file_size)}</Text>
            <Group gap={4}>
              <Tooltip label="Play video">
                <ActionIcon size="sm" variant="subtle">
                  <IconPlayerPlay size={14} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Open containing folder">
                <ActionIcon size="sm" variant="subtle" onClick={() => OpenFolder(video.file_path)}>
                  <IconFolder size={14} />
                </ActionIcon>
              </Tooltip>
              <Tooltip label="Delete from library">
                <ActionIcon size="sm" color="red" variant="subtle" onClick={onDelete}>
                  <IconTrash size={14} />
                </ActionIcon>
              </Tooltip>
            </Group>
          </Group>
        </Stack>
      </Paper>
    );
  }

  function VideoListItem({ video, onDelete, dark }: { video: app.VideoResult; onDelete: () => void; dark: boolean }) {
    return (
      <Paper 
        p="sm" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Group justify="space-between" align="flex-start">
          <Group gap="sm" align="flex-start" wrap="nowrap">
            <Image
              src={video.thumbnail_url || '/logo.svg'}
              w={120}
              h={68}
              radius="sm"
              alt={video.title}
              fallbackSrc="/logo.svg"
            />
            <Stack gap={4}>
              <Text fw={500} lineClamp={1} c={dark ? '#fff' : '#000'}>{video.title}</Text>
              <Text size="sm" c={dark ? 'dimmed' : 'gray.7'}>{video.channel}</Text>
              <Group gap="xs">
                <Badge size="sm">{formatDuration(video.duration)}</Badge>
                <Badge size="sm" variant="outline">{formatFileSize(video.file_size)}</Badge>
                <Badge size="sm" variant="outline">{video.quality}</Badge>
              </Group>
            </Stack>
          </Group>
          <Group gap={4}>
            <Tooltip label="Play video">
              <ActionIcon variant="subtle">
                <IconPlayerPlay size={18} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Open containing folder">
              <ActionIcon variant="subtle" onClick={() => OpenFolder(video.file_path)}>
                <IconFolder size={18} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label="Delete from library">
              <ActionIcon color="red" variant="subtle" onClick={onDelete}>
                <IconTrash size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </Paper>
    );
  }
}
