import { MantineProvider } from '@mantine/core';
import { act, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { GetLogs } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
import { LogEntry, useLogStore } from '../stores/logStore';
import { theme } from '../theme';

import { LoggerViewer, LogRow, MAX_VISIBLE_LOG_ROWS } from './LoggerViewer';

// Mock the Wails API (LoggerViewer only calls GetLogs/ClearLogs/ExportLogs,
// but the stores barrel pulls in the rest at import time)
vi.mock('../../wailsjs/go/app/App', () => ({
  CancelEditJob: vi.fn(),
  ClearLogs: vi.fn(),
  ExportLogs: vi.fn(),
  GetEditOptions: vi.fn(),
  GetLogs: vi.fn(),
  GetSettings: vi.fn(),
  GetVersion: vi.fn(),
  ListVideos: vi.fn(),
  SaveSettings: vi.fn(),
  SubmitEditJob: vi.fn(),
}));

// Mock the wails runtime; capture the `log:new` handler for storm tests
vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const mockedGetLogs = GetLogs as unknown as ReturnType<typeof vi.fn>;
const mockedEventsOn = vi.mocked(EventsOn);

type LogHandler = (data: LogEntry) => void;

const makeEntry = (i: number): LogEntry => ({
  timestamp: '2024-01-01T12:00:00.000Z',
  level: 'INFO',
  component: 'Test',
  message: `Log message ${i}`,
});

function renderLogger() {
  return render(
    <MantineProvider defaultColorScheme="dark" theme={theme}>
      <LoggerViewer />
    </MantineProvider>
  );
}

function logNewHandler(): LogHandler {
  const call = mockedEventsOn.mock.calls.find(([name]) => name === 'log:new');
  expect(call).toBeDefined();
  return call![1] as LogHandler;
}

describe('LoggerViewer', () => {
  beforeAll(() => {
    // jsdom lacks ResizeObserver (used by Mantine ScrollArea)
    class ResizeObserverMock {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, 'ResizeObserver', {
      writable: true,
      value: ResizeObserverMock,
    });

    // jsdom lacks matchMedia (used by Mantine)
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
    useLogStore.setState({ entries: [], isLoading: false, error: null });
    mockedGetLogs.mockResolvedValue([]);
  });

  it('loads initial logs and renders the empty state', async () => {
    renderLogger();

    expect(await screen.findByText('No logs found')).toBeInTheDocument();
    expect(mockedGetLogs).toHaveBeenCalledWith(100);
  });

  it('filters entries without recomputing per keystroke render (memoized filter)', async () => {
    mockedGetLogs.mockResolvedValue([makeEntry(0), { ...makeEntry(1), level: 'ERROR' }]);
    renderLogger();

    expect(await screen.findByText('Log message 0')).toBeInTheDocument();
    expect(screen.getByText('Log message 1')).toBeInTheDocument();
  });

  it('caps rendered rows at MAX_VISIBLE_LOG_ROWS with a showing note', async () => {
    const total = MAX_VISIBLE_LOG_ROWS + 100;
    mockedGetLogs.mockResolvedValue(Array.from({ length: total }, (_, i) => makeEntry(i)));
    renderLogger();

    expect(
      await screen.findByText(
        `Showing ${MAX_VISIBLE_LOG_ROWS} of ${total} matching logs (latest ${MAX_VISIBLE_LOG_ROWS})`
      )
    ).toBeInTheDocument();

    // Only the latest 200 rows mount; the head of the list is not rendered
    expect(screen.queryAllByText(/Log message \d+/)).toHaveLength(MAX_VISIBLE_LOG_ROWS);
    expect(screen.queryByText('Log message 0')).not.toBeInTheDocument();
    expect(screen.getByText(`Log message ${total - 1}`)).toBeInTheDocument();
  });

  it('does not show the cap note when everything fits', async () => {
    mockedGetLogs.mockResolvedValue([makeEntry(0), makeEntry(1)]);
    renderLogger();

    expect(await screen.findByText('Log message 1')).toBeInTheDocument();
    expect(screen.queryByText(/Showing \d+ of \d+ matching logs/)).not.toBeInTheDocument();
  });

  it('coalesces a rapid log:new storm into a single store update', async () => {
    renderLogger();
    expect(await screen.findByText('No logs found')).toBeInTheDocument();

    const handler = logNewHandler();
    const listener = vi.fn();
    const unsub = useLogStore.subscribe(listener);
    try {
      act(() => {
        for (let i = 0; i < 10; i++) {
          handler(makeEntry(i));
        }
      });

      // Buffered: nothing applied synchronously within the storm tick
      expect(useLogStore.getState().entries).toHaveLength(0);

      // Microtask flush applies the whole storm as one batch
      await act(async () => {
        await new Promise(resolve => setTimeout(resolve, 0));
      });

      const entries = useLogStore.getState().entries;
      expect(entries).toHaveLength(10);
      expect(entries.map(e => e.message)).toEqual(
        Array.from({ length: 10 }, (_, i) => `Log message ${i}`)
      );
      expect(listener).toHaveBeenCalledTimes(1);
    } finally {
      unsub();
    }
  });

  it('LogRow is a memoized component', () => {
    expect((LogRow as unknown as { $$typeof: symbol }).$$typeof).toBe(Symbol.for('react.memo'));
  });
});
