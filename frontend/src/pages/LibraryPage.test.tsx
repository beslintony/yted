import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { act, fireEvent, render, screen, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  DeleteVideo,
  GenerateEditPreview,
  GetEditOptions,
  GetEditVideoMetadata,
  GetLibraryStats,
  ListVideos,
} from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
import { useEditorStore } from '../stores/editorStore';
import { useLibraryStore } from '../stores/libraryStore';
import { theme } from '../theme';

import { LibraryPage } from './LibraryPage';

// Mock the Wails API (includes everything imported via the stores barrel)
vi.mock('../../wailsjs/go/app/App', () => ({
  ListVideos: vi.fn(),
  GetLibraryStats: vi.fn(),
  DeleteVideo: vi.fn(),
  OpenFile: vi.fn(),
  OpenFolder: vi.fn(),
  GetEditOptions: vi.fn(),
  GetEditVideoMetadata: vi.fn(),
  GenerateEditPreview: vi.fn(),
  SubmitEditJob: vi.fn(),
  CancelEditJob: vi.fn(),
  GetSettings: vi.fn(),
  SaveSettings: vi.fn(),
  GetVersion: vi.fn(),
  ExportLogs: vi.fn(),
}));

// Mock the wails runtime; EventsOn returns an unsubscribe function
vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const mockedListVideos = ListVideos as unknown as ReturnType<typeof vi.fn>;
const mockedGetLibraryStats = GetLibraryStats as unknown as ReturnType<typeof vi.fn>;
const mockedDeleteVideo = DeleteVideo as unknown as ReturnType<typeof vi.fn>;
const mockedGetEditOptions = GetEditOptions as unknown as ReturnType<typeof vi.fn>;
const mockedGetEditVideoMetadata = GetEditVideoMetadata as unknown as ReturnType<typeof vi.fn>;
const mockedGenerateEditPreview = GenerateEditPreview as unknown as ReturnType<typeof vi.fn>;
const mockedEventsOn = vi.mocked(EventsOn);

const mockVideos = [
  {
    id: 'video-1',
    youtube_id: 'yt-1',
    title: 'Alpha Video',
    channel: 'Channel A',
    channel_id: 'ch-a',
    duration: 300,
    description: '',
    thumbnail_url: '',
    file_path: '/videos/alpha.mp4',
    file_size: 1000000,
    format: 'mp4',
    quality: '1080p',
    downloaded_at: 1000,
    watch_position: 0,
    watch_count: 0,
  },
  {
    id: 'video-2',
    youtube_id: 'yt-2',
    title: 'Beta Video',
    channel: 'Channel B',
    channel_id: 'ch-b',
    duration: 600,
    description: '',
    thumbnail_url: '',
    file_path: '/videos/beta.mp4',
    file_size: 2000000,
    format: 'mp4',
    quality: '720p',
    downloaded_at: 2000,
    watch_position: 0,
    watch_count: 0,
  },
];

const mockEditOptions = {
  formats: [{ id: 'mp4', name: 'MP4', extension: 'mp4', description: '', codecs: ['h264'] }],
  codecs: [{ id: 'h264', name: 'H.264', description: '', quality: '', speed: '' }],
  crop_presets: [],
  effect_ranges: [
    { id: 'brightness', min: 0, max: 2, default: 1, step: 0.1, description: 'Brightness' },
  ],
  rotations: [{ value: 0, label: 'None', description: '' }],
  watermark_positions: { bottom_right: 'Bottom right' },
};

const mockMetadata = {
  duration: 300,
  width: 1920,
  height: 1080,
  fps: 30,
  bitrate: 5000000,
  codec: 'h264',
  has_audio: true,
};

const renderLibraryPage = () =>
  render(
    <MantineProvider defaultColorScheme="dark" theme={theme}>
      <ModalsProvider>
        <Notifications position="top-right" />
        <LibraryPage />
      </ModalsProvider>
    </MantineProvider>
  );

// Find an ActionIcon button by the tabler icon class it contains
const findIconButton = (iconClass: string) =>
  screen.getAllByRole('button').find(button => button.querySelector(`.${iconClass}`));

describe('LibraryPage', () => {
  let eventHandlers: Record<string, (data?: unknown) => void>;

  beforeAll(() => {
    // jsdom lacks ResizeObserver (used by Mantine FloatingIndicator/SegmentedControl)
    class ResizeObserverMock {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, 'ResizeObserver', {
      writable: true,
      value: ResizeObserverMock,
    });

    // jsdom lacks a working matchMedia (used by Mantine useColorScheme/useMediaQuery)
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: (query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: () => {},
        removeListener: () => {},
        addEventListener: () => {},
        removeEventListener: () => {},
        dispatchEvent: () => false,
      }),
    });
  });

  beforeEach(() => {
    vi.clearAllMocks();
    useLibraryStore.setState({ videos: [] });
    useEditorStore.setState({ options: null, ffmpegError: null, jobs: {} });

    eventHandlers = {};
    mockedEventsOn.mockImplementation((eventName, callback) => {
      eventHandlers[eventName] = callback;
      return vi.fn();
    });

    mockedListVideos.mockResolvedValue(mockVideos);
    mockedGetLibraryStats.mockResolvedValue({ total_videos: 2, total_size: 3000000 });
    mockedDeleteVideo.mockResolvedValue(undefined);
    mockedGetEditOptions.mockResolvedValue(mockEditOptions);
    mockedGetEditVideoMetadata.mockResolvedValue(mockMetadata);
    mockedGenerateEditPreview.mockResolvedValue('data:image/jpeg;base64,preview');
  });

  it('renders videos from ListVideos and library stats', async () => {
    renderLibraryPage();

    expect(await screen.findByText('Alpha Video')).toBeInTheDocument();
    expect(screen.getByText('Beta Video')).toBeInTheDocument();
    expect(screen.getByText('Channel A')).toBeInTheDocument();

    await waitFor(() => {
      expect(screen.getByText(/Library \(2 videos, 2\.9 MB\)/)).toBeInTheDocument();
    });

    expect(mockedListVideos).toHaveBeenCalledTimes(1);
    expect(mockedGetLibraryStats).toHaveBeenCalledTimes(1);
  });

  it('filters videos when typing in the search input', async () => {
    renderLibraryPage();

    expect(await screen.findByText('Alpha Video')).toBeInTheDocument();
    expect(screen.getByText('Beta Video')).toBeInTheDocument();

    fireEvent.change(screen.getByPlaceholderText('Search videos...'), {
      target: { value: 'alpha' },
    });

    await waitFor(() => {
      expect(screen.queryByText('Beta Video')).not.toBeInTheDocument();
    });
    expect(screen.getByText('Alpha Video')).toBeInTheDocument();
  });

  it('deletes a video through the confirm modal', async () => {
    renderLibraryPage();

    expect(await screen.findByText('Alpha Video')).toBeInTheDocument();

    // Click the delete ActionIcon on the first video card (Alpha Video)
    const deleteButton = findIconButton('tabler-icon-trash');
    expect(deleteButton).toBeDefined();
    fireEvent.click(deleteButton!);

    // The confirm modal from useNotifications().confirm appears
    expect(await screen.findByText('Delete Video')).toBeInTheDocument();
    expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument();

    fireEvent.click(screen.getByRole('button', { name: 'Delete' }));

    await waitFor(() => {
      expect(mockedDeleteVideo).toHaveBeenCalledWith('video-1', true);
    });
    await waitFor(() => {
      expect(screen.queryByText('Alpha Video')).not.toBeInTheDocument();
    });
    expect(useLibraryStore.getState().videos.some(v => v.id === 'video-1')).toBe(false);
  });

  it('opens the editor modal when clicking the edit action', async () => {
    renderLibraryPage();

    expect(await screen.findByText('Alpha Video')).toBeInTheDocument();

    // Click the edit ActionIcon on the first video card (Alpha Video)
    const editButton = findIconButton('tabler-icon-edit');
    expect(editButton).toBeDefined();
    fireEvent.click(editButton!);

    // EditorModal opens with the video title
    expect(await screen.findByText('Edit: Alpha Video')).toBeInTheDocument();

    await waitFor(() => {
      expect(mockedGetEditOptions).toHaveBeenCalled();
      expect(mockedGetEditVideoMetadata).toHaveBeenCalledWith('video-1');
    });

    // Trim tab shows the duration hint from the loaded metadata
    expect(await screen.findByText(/Total duration: 300\.0s/)).toBeInTheDocument();
  });

  it('reloads videos when a library:updated event is received', async () => {
    renderLibraryPage();

    expect(await screen.findByText('Alpha Video')).toBeInTheDocument();
    expect(mockedListVideos).toHaveBeenCalledTimes(1);
    expect(mockedGetLibraryStats).toHaveBeenCalledTimes(1);
    expect(eventHandlers['library:updated']).toBeDefined();

    act(() => {
      eventHandlers['library:updated']();
    });

    await waitFor(() => {
      expect(mockedListVideos).toHaveBeenCalledTimes(2);
      expect(mockedGetLibraryStats).toHaveBeenCalledTimes(2);
    });
  });
});
