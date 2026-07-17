import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { render, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { CheckDownloadStatus, GetDownloadQueue, GetSettings } from '../../wailsjs/go/app/App';
import { useDownloadStore } from '../stores/downloadStore';
import { theme } from '../theme';

import { DownloadPage } from './DownloadPage';

// Mock the Wails API (includes everything imported via the stores barrel)
vi.mock('../../wailsjs/go/app/App', () => ({
  AddDownload: vi.fn(),
  CheckDownloadStatus: vi.fn(),
  GetDownloadQueue: vi.fn(),
  GetSettings: vi.fn(),
  GetVideoInfo: vi.fn(),
  PauseDownload: vi.fn(),
  ResumeDownload: vi.fn(),
  RetryDownload: vi.fn(),
  StartProcessingDownloads: vi.fn(),
  ValidateURL: vi.fn(),
  ListVideos: vi.fn(),
  GetLibraryStats: vi.fn(),
  DeleteVideo: vi.fn(),
  GetEditOptions: vi.fn(),
  GetEditVideoMetadata: vi.fn(),
  GenerateEditPreview: vi.fn(),
  SubmitEditJob: vi.fn(),
  CancelEditJob: vi.fn(),
  SaveSettings: vi.fn(),
  GetVersion: vi.fn(),
  ExportLogs: vi.fn(),
}));

vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const mockedGetDownloadQueue = GetDownloadQueue as unknown as ReturnType<typeof vi.fn>;
const mockedCheckDownloadStatus = CheckDownloadStatus as unknown as ReturnType<typeof vi.fn>;
const mockedGetSettings = GetSettings as unknown as ReturnType<typeof vi.fn>;

function renderPage() {
  return render(
    <MantineProvider defaultColorScheme="dark" theme={theme}>
      <ModalsProvider>
        <Notifications />
        <DownloadPage />
      </ModalsProvider>
    </MantineProvider>
  );
}

describe('DownloadPage queue restore', () => {
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
    useDownloadStore.setState({ downloads: [], isLoading: false, error: null });
    mockedGetSettings.mockResolvedValue({ download_presets: [] });
    mockedGetDownloadQueue.mockResolvedValue(null);
    mockedCheckDownloadStatus.mockResolvedValue({ exists: true, status: 'completed' });
  });

  it('marks a stale "downloading" item as completed when the backend no longer reports it', async () => {
    // Ghost state: completion event was missed while the page was unmounted
    useDownloadStore.setState({
      downloads: [
        {
          id: 'ghost-1',
          url: 'https://www.youtube.com/watch?v=abc123xyz00',
          status: 'downloading',
          progress: 100,
          title: 'Ghost Video',
          channel: 'Channel',
          createdAt: Date.now(),
        },
      ],
      isLoading: false,
      error: null,
    });

    renderPage();

    await waitFor(() => {
      expect(mockedCheckDownloadStatus).toHaveBeenCalledWith('ghost-1');
    });
    await waitFor(() => {
      expect(useDownloadStore.getState().downloads[0].status).toBe('completed');
    });
  });

  it('restores a paused download with paused status', async () => {
    mockedGetDownloadQueue.mockResolvedValue([
      {
        id: 'paused-1',
        url: 'https://www.youtube.com/watch?v=def456uvw01',
        status: 'paused',
        progress: 40,
        title: 'Paused Video',
        channel: 'Channel',
        thumbnail_url: '',
        format_id: 'best',
        quality: 'best',
        error_message: '',
        youtube_id: 'def456uvw01',
      },
    ]);

    renderPage();

    await waitFor(() => {
      const dl = useDownloadStore.getState().downloads.find(d => d.id === 'paused-1');
      expect(dl?.status).toBe('paused');
      expect(dl?.progress).toBe(40);
    });
  });

  it('restores an errored download with its error message', async () => {
    mockedGetDownloadQueue.mockResolvedValue([
      {
        id: 'err-1',
        url: 'https://www.youtube.com/watch?v=ghi789rst02',
        status: 'error',
        progress: 12,
        title: 'Failed Video',
        channel: 'Channel',
        thumbnail_url: '',
        format_id: 'best',
        quality: 'best',
        error_message: 'HTTP Error 403: Forbidden',
        youtube_id: 'ghi789rst02',
      },
    ]);

    renderPage();

    await waitFor(() => {
      const dl = useDownloadStore.getState().downloads.find(d => d.id === 'err-1');
      expect(dl?.status).toBe('error');
      expect(dl?.errorMessage).toBe('HTTP Error 403: Forbidden');
    });
  });
});
