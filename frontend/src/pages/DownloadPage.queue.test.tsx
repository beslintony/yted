import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { fireEvent, render, waitFor } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { CancelDownload, GetDownloadQueue, GetSettings } from '../../wailsjs/go/app/App';
import { useDownloadStore } from '../stores/downloadStore';
import { theme } from '../theme';
import { Download } from '../types';

import { DownloadPage } from './DownloadPage';

vi.mock('../../wailsjs/go/app/App', () => ({
  AddDownload: vi.fn(),
  AddPlaylistDownload: vi.fn(),
  CancelDownload: vi.fn(),
  CheckDownloadStatus: vi.fn(),
  ClearDownloadCache: vi.fn(),
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

const mockedCancelDownload = CancelDownload as unknown as ReturnType<typeof vi.fn>;
const mockedGetDownloadQueue = GetDownloadQueue as unknown as ReturnType<typeof vi.fn>;
const mockedGetSettings = GetSettings as unknown as ReturnType<typeof vi.fn>;

function makeDownload(id: string, status: Download['status'], title: string): Download {
  return {
    id,
    url: `https://www.youtube.com/watch?v=${id.padEnd(11, '0')}`,
    status,
    progress: status === 'completed' ? 100 : 10,
    title,
    channel: 'Channel',
    createdAt: Date.now(),
  };
}

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

describe('DownloadPage queue display', () => {
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
    mockedGetDownloadQueue.mockResolvedValue(null);
    mockedCancelDownload.mockResolvedValue(undefined);
  });

  it('groups the queue by status and collapses completed by default', async () => {
    useDownloadStore.setState({
      downloads: [
        makeDownload('dl-1', 'downloading', 'Active Video'),
        makeDownload('dl-2', 'pending', 'Waiting Video'),
        makeDownload('dl-3', 'error', 'Broken Video'),
        makeDownload('dl-4', 'completed', 'Done Video One'),
        makeDownload('dl-5', 'completed', 'Done Video Two'),
      ],
      isLoading: false,
      error: null,
    });

    const { getByText, queryByText } = renderPage();

    // Live per-status count badges in the header
    expect(getByText('1 downloading')).toBeTruthy();
    expect(getByText('1 pending')).toBeTruthy();
    expect(getByText('1 failed')).toBeTruthy();
    expect(getByText('2 completed')).toBeTruthy();

    // Section headers with counts
    expect(getByText('Downloading (1)')).toBeTruthy();
    expect(getByText('Pending (1)')).toBeTruthy();
    expect(getByText('Failed (1)')).toBeTruthy();
    expect(getByText('Completed (2)')).toBeTruthy();

    // Active items visible, completed hidden while collapsed
    expect(getByText('Active Video')).toBeTruthy();
    expect(getByText('Waiting Video')).toBeTruthy();
    expect(getByText('Broken Video')).toBeTruthy();
    expect(queryByText('Done Video One')).toBeNull();

    // Expand completed
    fireEvent.click(getByText('Show'));
    await waitFor(() => {
      expect(getByText('Done Video One')).toBeTruthy();
      expect(getByText('Done Video Two')).toBeTruthy();
    });
  });

  it('batch-adds downloads with dedupe in one update', () => {
    const { addDownloads } = useDownloadStore.getState();
    useDownloadStore.setState({
      downloads: [makeDownload('dl-1', 'pending', 'Existing')],
      isLoading: false,
      error: null,
    });

    addDownloads([
      makeDownload('dl-1', 'pending', 'Duplicate'),
      makeDownload('dl-2', 'pending', 'New One'),
      makeDownload('dl-3', 'pending', 'New Two'),
    ]);

    const { downloads } = useDownloadStore.getState();
    expect(downloads).toHaveLength(3);
    expect(downloads.filter(d => d.id === 'dl-1')).toHaveLength(1);
    // Fresh items are prepended, existing order preserved
    expect(downloads[0].id).toBe('dl-3');
    expect(downloads[2].id).toBe('dl-1');
  });

  it('trash button cancels on the backend and removes from the queue', async () => {
    useDownloadStore.setState({
      downloads: [makeDownload('dl-9', 'pending', 'Doomed Video')],
      isLoading: false,
      error: null,
    });

    const { container } = renderPage();
    const trash = container.querySelector('.tabler-icon-trash');
    expect(trash).not.toBeNull();
    fireEvent.click(trash!.closest('button')!);

    await waitFor(() => {
      expect(mockedCancelDownload).toHaveBeenCalledWith('dl-9');
    });
    await waitFor(() => {
      expect(useDownloadStore.getState().downloads).toHaveLength(0);
    });
  });
});
