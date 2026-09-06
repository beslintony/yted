import { MantineProvider } from '@mantine/core';
import { act, render, screen, waitFor } from '@testing-library/react';
import { memo } from 'react';
import { beforeAll, beforeEach, describe, expect, it, vi } from 'vitest';

import { GetFFmpegLocations, GetSettings } from '../../wailsjs/go/app/App';
import { theme } from '../theme';

import { SettingsPage } from './SettingsPage';

const captured = vi.hoisted(() => ({}) as Record<string, Array<Record<string, unknown>>>);
const renderCounts = vi.hoisted(() => ({}) as Record<string, number>);

// Capture the props SettingsPage passes to each section on every render so
// the test can assert callback stability across settings changes. The stubs
// are themselves memoized — exactly like the real sections — so renderCounts
// only increase when a stub's props actually change, proving which sections
// skip re-render per keystroke.
vi.mock('../components/settings', () => {
  const make = (name: string) => {
    const Inner = (props: Record<string, unknown>) => {
      renderCounts[name] = (renderCounts[name] || 0) + 1;
      captured[name] = captured[name] || [];
      captured[name].push(props);
      return null;
    };
    Inner.displayName = `${name}Stub`;
    return memo(Inner);
  };
  return {
    ActionsSection: make('ActionsSection'),
    CacheSection: make('CacheSection'),
    DownloadsSection: make('DownloadsSection'),
    FfmpegSection: make('FfmpegSection'),
    LoggingSection: make('LoggingSection'),
    PlayerSection: make('PlayerSection'),
    PresetsSection: make('PresetsSection'),
    UiSection: make('UiSection'),
  };
});

vi.mock('../../wailsjs/go/app/App', () => ({
  GetFFmpegLocations: vi.fn(),
  GetSettings: vi.fn(),
  RefreshFFmpegStatus: vi.fn(),
  SaveSettings: vi.fn(),
  ShowFFmpegDialog: vi.fn(),
  ShowOpenDirectoryDialog: vi.fn(),
}));

vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const mockedGetSettings = GetSettings as unknown as ReturnType<typeof vi.fn>;
const mockedGetFFmpegLocations = GetFFmpegLocations as unknown as ReturnType<typeof vi.fn>;

const lastProps = (name: string): Record<string, unknown> => {
  const calls = captured[name] || [];
  expect(calls.length).toBeGreaterThan(0);
  return calls[calls.length - 1];
};

function renderSettings() {
  return render(
    <MantineProvider defaultColorScheme="dark" theme={theme}>
      <SettingsPage />
    </MantineProvider>
  );
}

describe('SettingsPage memo wiring', () => {
  beforeAll(() => {
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
    for (const key of Object.keys(captured)) {
      delete captured[key];
    }
    for (const key of Object.keys(renderCounts)) {
      delete renderCounts[key];
    }
    mockedGetSettings.mockResolvedValue({
      download_path: '/tmp/downloads',
      max_concurrent_downloads: 3,
      default_quality: 'best',
      theme: 'dark',
    });
    mockedGetFFmpegLocations.mockResolvedValue({ installed: false });
  });

  it('loads once and keeps section callbacks stable across a settings change', async () => {
    renderSettings();

    await waitFor(() => expect((captured.DownloadsSection || []).length).toBeGreaterThan(0));

    const beforeDownloads = lastProps('DownloadsSection');
    const beforeCache = lastProps('CacheSection');
    const beforeActions = lastProps('ActionsSection');
    const beforeFfmpeg = lastProps('FfmpegSection');
    const beforeUi = lastProps('UiSection');

    // Simulate a keystroke in the downloads section
    act(() => {
      (beforeDownloads.updateSetting as (key: string, value: unknown) => void)(
        'filename_template',
        'new-template'
      );
    });

    const afterDownloads = lastProps('DownloadsSection');

    // The settings object identity changes (sections that read settings
    // legitimately re-render)...
    expect(afterDownloads.settings).not.toBe(beforeDownloads.settings);

    // ...but every callback prop keeps its identity, so memoized sections
    // whose inputs didn't change bail out.
    expect(afterDownloads.updateSetting).toBe(beforeDownloads.updateSetting);
    expect(lastProps('CacheSection').confirm).toBe(beforeCache.confirm);
    expect(lastProps('CacheSection').success).toBe(beforeCache.success);
    expect(lastProps('CacheSection').error).toBe(beforeCache.error);
    expect(lastProps('ActionsSection').handleSave).toBe(beforeActions.handleSave);
    expect(lastProps('ActionsSection').handleReset).toBe(beforeActions.handleReset);
    expect(lastProps('FfmpegSection').getSelectedFfmpegInfo).toBe(
      beforeFfmpeg.getSelectedFfmpegInfo
    );
    expect(lastProps('UiSection').handleThemeChange).toBe(beforeUi.handleThemeChange);

    // Dirty-check behavior is unchanged: editing marks unsaved changes...
    expect(await screen.findByText('Unsaved Changes')).toBeInTheDocument();

    // ...and keystrokes never refetch settings from the backend.
    expect(mockedGetSettings).toHaveBeenCalledTimes(1);
  });

  it('skips memoized sections whose props did not change (render counts)', async () => {
    renderSettings();

    await waitFor(() => expect((captured.DownloadsSection || []).length).toBeGreaterThan(0));

    const baseline = { ...renderCounts };
    const downloadsRenders = baseline.DownloadsSection || 0;
    const cacheRenders = baseline.CacheSection || 0;
    const actionsRenders = baseline.ActionsSection || 0;

    const update = (value: string) => {
      const props = lastProps('DownloadsSection');
      act(() => {
        (props.updateSetting as (key: string, v: unknown) => void)('filename_template', value);
      });
    };

    // First keystroke flips hasChanges false -> true, so ActionsSection
    // (which reads hasChanges) legitimately re-renders once...
    update('template-one');
    expect(await screen.findByText('Unsaved Changes')).toBeInTheDocument();
    expect(renderCounts.DownloadsSection || 0).toBeGreaterThan(downloadsRenders);
    expect(renderCounts.ActionsSection).toBe(actionsRenders + 1);
    // ...but CacheSection takes no settings state and all its props are
    // stable, so it skips the keystroke entirely.
    expect(renderCounts.CacheSection).toBe(cacheRenders);

    // Second keystroke changes nothing ActionsSection reads (hasChanges stays
    // true, handlers are stable), so it skips too — this only holds because
    // handleSave/handleReset keep referential stability via refs.
    const actionsAfterFirst = renderCounts.ActionsSection || 0;
    update('template-two');
    expect(renderCounts.CacheSection).toBe(cacheRenders);
    expect(renderCounts.ActionsSection).toBe(actionsAfterFirst);

    expect(mockedGetSettings).toHaveBeenCalledTimes(1);
  });

  it('does not render sections before settings load', async () => {
    let resolveSettings: (value: unknown) => void = () => {};
    mockedGetSettings.mockReturnValue(
      new Promise(resolve => {
        resolveSettings = resolve;
      })
    );

    renderSettings();

    expect(screen.getByText('Loading settings...')).toBeInTheDocument();
    expect(captured.DownloadsSection || []).toHaveLength(0);

    await act(async () => {
      resolveSettings({
        download_path: '/tmp/downloads',
        max_concurrent_downloads: 3,
        default_quality: 'best',
        theme: 'dark',
      });
    });

    await waitFor(() => expect((captured.DownloadsSection || []).length).toBeGreaterThan(0));
  });
});
