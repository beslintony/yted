import { beforeEach, describe, expect, it } from 'vitest';

import { DEFAULT_SETTINGS } from '../types';
import { useSettingsStore } from './settingsStore';

describe('settingsStore', () => {
  beforeEach(() => {
    // Reset store to defaults before each test
    useSettingsStore.setState({ ...DEFAULT_SETTINGS, isLoading: false, error: null });
  });

  it('should initialize with default settings', () => {
    const state = useSettingsStore.getState();
    expect(state.downloadPath).toBe(DEFAULT_SETTINGS.downloadPath);
    expect(state.maxConcurrentDownloads).toBe(DEFAULT_SETTINGS.maxConcurrentDownloads);
    expect(state.theme).toBe(DEFAULT_SETTINGS.theme);
  });

  it('should update download path', () => {
    const { setDownloadPath } = useSettingsStore.getState();
    setDownloadPath('/custom/path');
    expect(useSettingsStore.getState().downloadPath).toBe('/custom/path');
  });

  it('should clamp max concurrent downloads between 1-10', () => {
    const { setMaxConcurrentDownloads } = useSettingsStore.getState();

    setMaxConcurrentDownloads(15);
    expect(useSettingsStore.getState().maxConcurrentDownloads).toBe(10);

    setMaxConcurrentDownloads(0);
    expect(useSettingsStore.getState().maxConcurrentDownloads).toBe(1);

    setMaxConcurrentDownloads(5);
    expect(useSettingsStore.getState().maxConcurrentDownloads).toBe(5);
  });

  it('should toggle sidebar', () => {
    const { toggleSidebar } = useSettingsStore.getState();
    const initial = useSettingsStore.getState().sidebarCollapsed;

    toggleSidebar();
    expect(useSettingsStore.getState().sidebarCollapsed).toBe(!initial);

    toggleSidebar();
    expect(useSettingsStore.getState().sidebarCollapsed).toBe(initial);
  });

  it('should add download preset', () => {
    const { addDownloadPreset } = useSettingsStore.getState();
    const initialCount = useSettingsStore.getState().downloadPresets.length;

    addDownloadPreset({
      name: 'Test Preset',
      format: 'best',
      quality: '720p',
      extension: 'mp4',
    });

    expect(useSettingsStore.getState().downloadPresets).toHaveLength(initialCount + 1);
  });

  it('should remove download preset', () => {
    const { removeDownloadPreset } = useSettingsStore.getState();
    const presets = useSettingsStore.getState().downloadPresets;
    const initialCount = presets.length;

    if (presets.length > 0) {
      removeDownloadPreset(presets[0].id);
      expect(useSettingsStore.getState().downloadPresets).toHaveLength(initialCount - 1);
    }
  });

  it('should reset to defaults', () => {
    const { setTheme, setDownloadPath, resetToDefaults } = useSettingsStore.getState();

    setTheme('light');
    setDownloadPath('/custom');

    resetToDefaults();

    expect(useSettingsStore.getState().theme).toBe(DEFAULT_SETTINGS.theme);
    expect(useSettingsStore.getState().downloadPath).toBe(DEFAULT_SETTINGS.downloadPath);
  });
});
