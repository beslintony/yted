import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import {
  AddPlaylistDownload,
  GetPlaylistInfo,
  GetSettings,
  ValidateURL,
} from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
import { useDownloadStore } from '../stores/downloadStore';
import { theme } from '../theme';

import { DownloadPage } from './DownloadPage';

// Mock the Wails API (includes everything imported via the stores barrel)
vi.mock('../../wailsjs/go/app/App', () => ({
  AddDownload: vi.fn(),
  AddPlaylistDownload: vi.fn(),
  CheckDownloadStatus: vi.fn(),
  GetDownloadQueue: vi.fn(),
  GetPlaylistInfo: vi.fn(),
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

const mockedGetPlaylistInfo = GetPlaylistInfo as unknown as ReturnType<typeof vi.fn>;
const mockedAddPlaylistDownload = AddPlaylistDownload as unknown as ReturnType<typeof vi.fn>;
const mockedGetSettings = GetSettings as unknown as ReturnType<typeof vi.fn>;
const mockedValidateURL = ValidateURL as unknown as ReturnType<typeof vi.fn>;
const mockedEventsOn = vi.mocked(EventsOn);

const PLAYLIST_URL = 'https://www.youtube.com/playlist?list=PLtest123';

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

describe('DownloadPage playlist flow', () => {
  beforeAll(() => {
    class ResizeObserverMock {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, 'ResizeObserver', {
      writable: true,
      value: ResizeObserverMock,
    });
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
    mockedValidateURL.mockResolvedValue(true);
    mockedGetPlaylistInfo.mockResolvedValue({
      id: 'PLtest123',
      title: 'Test Playlist',
      channel: 'Test Channel',
      count: 2,
      entries: [
        { id: 'v1', title: 'Video One', duration: 60 },
        { id: 'v2', title: 'Video Two', duration: 120 },
      ],
    });
    mockedAddPlaylistDownload.mockResolvedValue(2);
  });

  it('fetches playlist info and queues all videos', async () => {
    const { getByPlaceholderText, getByText } = renderPage();

    fireEvent.change(getByPlaceholderText('Paste YouTube URL here...'), {
      target: { value: PLAYLIST_URL },
    });
    fireEvent.click(getByText('Get Info'));

    // Playlist card appears with title and entry count
    await waitFor(() => {
      expect(getByText('Test Playlist')).toBeTruthy();
      expect(getByText('2 videos')).toBeTruthy();
    });

    fireEvent.click(getByText('Download 2 videos'));

    await waitFor(() => {
      expect(mockedAddPlaylistDownload).toHaveBeenCalledWith(PLAYLIST_URL, 'best', 'best');
    });
  });

  it('syncs backend-added downloads into the store via download:added', async () => {
    renderPage();

    // Grab the registered download:added handler and fire it
    const addedHandler = mockedEventsOn.mock.calls.find(
      call => call[0] === 'download:added'
    )?.[1] as ((data: unknown) => void) | undefined;
    expect(addedHandler).toBeDefined();

    addedHandler!({
      id: 'dl-1',
      url: 'https://www.youtube.com/watch?v=v1',
      status: 'pending',
      progress: 0,
      title: 'Video One',
      channel: 'Test Channel',
      thumbnail_url: '',
      format_id: 'bestvideo+bestaudio/best',
      quality: 'best',
      duration: 60,
    });

    await waitFor(() => {
      const dl = useDownloadStore.getState().downloads.find(d => d.id === 'dl-1');
      expect(dl?.title).toBe('Video One');
      expect(dl?.status).toBe('pending');
    });
  });
});
