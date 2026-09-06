import { MantineProvider } from '@mantine/core';
import { fireEvent, render, screen } from '@testing-library/react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { GetVersion } from '../wailsjs/go/app/App';

import App from './App';
import { useSettingsStore } from './stores/settingsStore';
import { theme } from './theme';

const mounts = vi.hoisted(() => ({ downloads: 0, library: 0, settings: 0 }));

// Mock the wails runtime; EventsOn returns an unsubscribe function
vi.mock('../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

// Mock the Wails API surface pulled in via the stores barrel
vi.mock('../wailsjs/go/app/App', () => ({
  CancelEditJob: vi.fn(),
  ClearLogs: vi.fn(),
  ExportLogs: vi.fn(),
  GetDownloadQueue: vi.fn(),
  GetEditOptions: vi.fn(),
  GetLibraryStats: vi.fn(),
  GetLogs: vi.fn(),
  GetSettings: vi.fn(),
  GetVersion: vi.fn(),
  ListVideos: vi.fn(),
  SaveSettings: vi.fn(),
  StartProcessingDownloads: vi.fn(),
  SubmitEditJob: vi.fn(),
}));

// Mock the logger drawer content (unrelated to tab keep-alive)
vi.mock('./components/LoggerViewer', () => ({
  LoggerViewer: () => null,
}));

// Stateful mock pages: a text input proves state survives tab switches,
// and the mount counters prove mount-on-first-visit (no remounts).
vi.mock('./pages/DownloadPage', async () => {
  const React = await import('react');
  return {
    DownloadPage: () => {
      const [value, setValue] = React.useState('');
      React.useEffect(() => {
        mounts.downloads += 1;
      }, []);
      return (
        <input
          aria-label="downloads-probe"
          value={value}
          onChange={e => setValue(e.target.value)}
        />
      );
    },
  };
});

vi.mock('./pages/LibraryPage', async () => {
  const React = await import('react');
  return {
    LibraryPage: () => {
      const [value, setValue] = React.useState('');
      React.useEffect(() => {
        mounts.library += 1;
      }, []);
      return (
        <input aria-label="library-probe" value={value} onChange={e => setValue(e.target.value)} />
      );
    },
  };
});

vi.mock('./pages/SettingsPage', async () => {
  const React = await import('react');
  return {
    SettingsPage: () => {
      React.useEffect(() => {
        mounts.settings += 1;
      }, []);
      return <div>settings-probe</div>;
    },
  };
});

const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

function renderApp() {
  return render(
    <MantineProvider defaultColorScheme="dark" theme={theme}>
      <App />
    </MantineProvider>
  );
}

describe('App keep-alive tabs', () => {
  beforeAll(() => {
    // jsdom lacks ResizeObserver (used by Mantine AppShell)
    class ResizeObserverMock {
      observe() {}
      unobserve() {}
      disconnect() {}
    }
    Object.defineProperty(window, 'ResizeObserver', {
      writable: true,
      value: ResizeObserverMock,
    });

    // jsdom lacks a working matchMedia (used by Mantine hooks)
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
    mounts.downloads = 0;
    mounts.library = 0;
    mounts.settings = 0;
    useSettingsStore.setState({ sidebarCollapsed: false });
    mockedGetVersion.mockResolvedValue('1.0.0-test');
  });

  it('mounts only the downloads tab initially (lazy startup)', () => {
    renderApp();

    expect(screen.getByTestId('tab-panel-downloads')).toBeInTheDocument();
    expect(screen.queryByTestId('tab-panel-library')).not.toBeInTheDocument();
    expect(screen.queryByTestId('tab-panel-settings')).not.toBeInTheDocument();
    expect(mounts).toEqual({ downloads: 1, library: 0, settings: 0 });
  });

  it('hides inactive tabs instead of unmounting, preserving their state', () => {
    renderApp();

    fireEvent.change(screen.getByLabelText('downloads-probe'), {
      target: { value: 'my queue search' },
    });

    fireEvent.click(screen.getByText('Library'));

    // Library mounted on first visit; downloads hidden but still mounted
    expect(screen.getByTestId('tab-panel-library')).toBeVisible();
    expect(screen.getByTestId('tab-panel-downloads')).not.toBeVisible();
    expect(mounts.downloads).toBe(1);

    fireEvent.change(screen.getByLabelText('library-probe'), {
      target: { value: 'my library search' },
    });

    // Switching back restores both states with no remounts
    fireEvent.click(screen.getByText('Downloads'));

    expect(screen.getByTestId('tab-panel-downloads')).toBeVisible();
    expect(screen.getByLabelText('downloads-probe')).toHaveValue('my queue search');
    expect(screen.getByTestId('tab-panel-library')).not.toBeVisible();
    expect(mounts).toEqual({ downloads: 1, library: 1, settings: 0 });

    fireEvent.click(screen.getByText('Library'));
    expect(screen.getByLabelText('library-probe')).toHaveValue('my library search');
    expect(mounts.library).toBe(1);
  });

  it('mounts each tab exactly once across repeated switches', () => {
    renderApp();

    fireEvent.click(screen.getByText('Library'));
    fireEvent.click(screen.getByText('Settings'));
    fireEvent.click(screen.getByText('Downloads'));
    fireEvent.click(screen.getByText('Settings'));
    fireEvent.click(screen.getByText('Library'));

    expect(mounts).toEqual({ downloads: 1, library: 1, settings: 1 });

    // All visited tabs stay in the DOM; only the active one is visible
    expect(screen.getByTestId('tab-panel-downloads')).not.toBeVisible();
    expect(screen.getByTestId('tab-panel-library')).toBeVisible();
    expect(screen.getByTestId('tab-panel-settings')).not.toBeVisible();
    expect(screen.getByText('settings-probe')).toBeInTheDocument();
  });

  it('does not refetch on tab switches (existing mount effects run once)', () => {
    renderApp();

    fireEvent.click(screen.getByText('Library'));
    fireEvent.click(screen.getByText('Settings'));
    fireEvent.click(screen.getByText('Downloads'));

    // Version fetch (App mount effect) still runs exactly once
    expect(mockedGetVersion).toHaveBeenCalledTimes(1);
  });
});
