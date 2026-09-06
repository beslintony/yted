import { MantineProvider } from '@mantine/core';
import { ModalsProvider } from '@mantine/modals';
import { Notifications } from '@mantine/notifications';
import { act, render } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { GetDownloadQueue, GetSettings } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
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

const mockedEventsOn = vi.mocked(EventsOn);
const mockedGetDownloadQueue = GetDownloadQueue as unknown as ReturnType<typeof vi.fn>;
const mockedGetSettings = GetSettings as unknown as ReturnType<typeof vi.fn>;

const EXPECTED_EVENTS = [
  'download:progress',
  'download:completed',
  'download:error',
  'download:started',
  'download:added',
  'download:cancelled',
  'download:retried',
];

function makeDownload(id: string): Download {
  return {
    id,
    url: `https://www.youtube.com/watch?v=${id.padEnd(11, '0')}`,
    status: 'downloading',
    progress: 10,
    title: 'Active Video',
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

function callsFor(event: string) {
  return mockedEventsOn.mock.calls.filter(([name]) => name === event);
}

describe('DownloadPage - stable event subscriptions', () => {
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
  });

  it('subscribes to all 7 backend events exactly once on mount', () => {
    useDownloadStore.setState({
      downloads: [makeDownload('dl-stable-1')],
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(mockedEventsOn).toHaveBeenCalledTimes(7);
    for (const event of EXPECTED_EVENTS) {
      expect(callsFor(event)).toHaveLength(1);
    }
  });

  it('does not re-subscribe when a progress update re-renders the page', () => {
    useDownloadStore.setState({
      downloads: [makeDownload('dl-stable-2')],
      isLoading: false,
      error: null,
    });

    renderPage();

    expect(mockedEventsOn).toHaveBeenCalledTimes(7);

    // High-frequency progress tick: re-renders via the downloads slice but
    // must not tear down / re-create the backend subscriptions
    act(() => {
      useDownloadStore.getState().updateProgressInfo('dl-stable-2', {
        progress: 55,
        speed: '2.0 MiB/s',
        eta: '00:10',
      });
    });

    expect(useDownloadStore.getState().downloads[0].progress).toBe(55);
    expect(mockedEventsOn).toHaveBeenCalledTimes(7);
    for (const event of EXPECTED_EVENTS) {
      expect(callsFor(event)).toHaveLength(1);
    }

    // A second tick keeps subscriptions stable too
    act(() => {
      useDownloadStore.getState().updateProgressInfo('dl-stable-2', { progress: 56 });
    });

    expect(mockedEventsOn).toHaveBeenCalledTimes(7);
    for (const event of EXPECTED_EVENTS) {
      expect(callsFor(event)).toHaveLength(1);
    }
  });
});
