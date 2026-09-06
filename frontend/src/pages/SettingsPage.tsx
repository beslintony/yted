import { Badge, Group, Paper, Stack, Text, useMantineColorScheme } from '@mantine/core';
import { useCallback, useEffect, useRef, useState } from 'react';

import {
  GetFFmpegLocations,
  GetSettings,
  RefreshFFmpegStatus,
  SaveSettings,
  ShowFFmpegDialog,
  ShowOpenDirectoryDialog,
} from '../../wailsjs/go/app/App';
import { app, config } from '../../wailsjs/go/models';
import {
  ActionsSection,
  CacheSection,
  DownloadsSection,
  FfmpegSection,
  LoggingSection,
  PlayerSection,
  PresetsSection,
  UiSection,
} from '../components/settings';
import { useNotifications, useSettingsStore } from '../stores';
import { QualityOption, ThemeMode } from '../types';

export function SettingsPage() {
  const { colorScheme, setColorScheme: setMantineColorScheme } = useMantineColorScheme();
  const [settings, setSettings] = useState<config.Config | null>(null);
  const [originalSettings, setOriginalSettings] = useState<config.Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [ffmpegStatus, setFfmpegStatus] = useState<app.FFmpegCheckResult | null>(null);
  const [loadingFfmpeg, setLoadingFfmpeg] = useState(false);

  const {
    setTheme,
    toggleSidebar,
    setDownloadPath,
    setMaxConcurrentDownloads,
    setDefaultQuality,
    saveSettings: saveSettingsToStore,
  } = useSettingsStore();

  const { success, error, confirm } = useNotifications();

  const dark = colorScheme === 'dark';

  // Latest-ref pattern: handleSave/handleReset read through refs so their
  // identities stay stable across keystrokes. Combined with the memoized
  // section components below, this keeps sections that don't depend on the
  // edited values (Cache, Actions) from re-rendering per keystroke.
  const settingsRef = useRef(settings);
  settingsRef.current = settings;
  const originalSettingsRef = useRef(originalSettings);
  originalSettingsRef.current = originalSettings;

  // useMantineColorScheme() returns new setColorScheme/clearColorScheme
  // closures on every render (plain function declarations, not useCallback),
  // so stabilize it here — otherwise every useCallback depending on it (and
  // every memoized child receiving those callbacks) would invalidate per
  // render. Behavior is identical: it always calls the latest setter.
  const setColorSchemeRef = useRef(setMantineColorScheme);
  setColorSchemeRef.current = setMantineColorScheme;
  const setColorScheme = useCallback(
    (value: 'auto' | 'dark' | 'light') => setColorSchemeRef.current(value),
    []
  );

  useEffect(() => {
    loadSettings();
    loadFfmpegStatus();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (originalSettings && settings) {
      const changed = JSON.stringify(originalSettings) !== JSON.stringify(settings);
      setHasChanges(changed);
    }
  }, [settings, originalSettings]);

  const loadSettings = async () => {
    try {
      const result = await GetSettings();
      if (result) {
        setSettings(result);
        setOriginalSettings(JSON.parse(JSON.stringify(result)));

        setDownloadPath(result.download_path);
        setMaxConcurrentDownloads(result.max_concurrent_downloads);
        setDefaultQuality(result.default_quality as QualityOption);
        setTheme(result.theme as ThemeMode);
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = useCallback(async () => {
    const current = settingsRef.current;
    if (!current) return;

    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);

    try {
      await SaveSettings(current);
      setOriginalSettings(JSON.parse(JSON.stringify(current)));
      setHasChanges(false);
      setSaveSuccess(true);

      await saveSettingsToStore();

      if (current.theme === 'dark' && colorScheme !== 'dark') {
        setColorScheme('dark');
      } else if (current.theme === 'light' && colorScheme !== 'light') {
        setColorScheme('light');
      }

      success('Settings Saved', 'Your settings have been saved successfully');
      setTimeout(() => setSaveSuccess(false), 3000);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      console.error('Failed to save settings:', err);
      const msg = err?.message || 'Failed to save settings';
      setSaveError(msg);
      error('Save Failed', msg);
    } finally {
      setSaving(false);
    }
  }, [colorScheme, setColorScheme, saveSettingsToStore, success, error]);

  const handleBrowseDownloadPath = useCallback(async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, download_path: path } as any) : null));
        setDownloadPath(path);
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  }, [setDownloadPath]);

  const handleBrowseLogExportPath = useCallback(async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, log_export_path: path } as any) : null));
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  }, []);

  const handleBrowseLogPath = useCallback(async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, log_path: path } as any) : null));
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  }, []);

  const refreshFfmpegStatus = useCallback(async () => {
    setLoadingFfmpeg(true);
    try {
      const result = await RefreshFFmpegStatus();
      setFfmpegStatus(result);
    } catch (err) {
      console.error('Failed to refresh FFmpeg status:', err);
    } finally {
      setLoadingFfmpeg(false);
    }
  }, []);

  const handleBrowseFFmpeg = useCallback(async () => {
    try {
      const path = await ShowFFmpegDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, ffmpeg_path: path } as any) : null));
        // Clear cache and refresh FFmpeg status immediately
        await refreshFfmpegStatus();
      }
    } catch (err) {
      console.error('Failed to browse for ffmpeg:', err);
    }
  }, [refreshFfmpegStatus]);

  const loadFfmpegStatus = async () => {
    setLoadingFfmpeg(true);
    try {
      const result = await GetFFmpegLocations();
      setFfmpegStatus(result);
    } catch (err) {
      console.error('Failed to load FFmpeg status:', err);
    } finally {
      setLoadingFfmpeg(false);
    }
  };

  const getSelectedFfmpegInfo = useCallback(() => {
    if (!ffmpegStatus || !ffmpegStatus.installed) return null;
    if (ffmpegStatus.selectedIndex >= 0 && ffmpegStatus.allLocations) {
      return ffmpegStatus.allLocations[ffmpegStatus.selectedIndex];
    }
    return null;
  }, [ffmpegStatus]);

  const handleReset = useCallback(() => {
    if (originalSettingsRef.current) {
      setSettings(JSON.parse(JSON.stringify(originalSettingsRef.current)));
    }
  }, []);

  const handleThemeChange = useCallback(
    (value: string) => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setSettings(s => (s ? ({ ...s, theme: value } as any) : null));
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      setTheme(value as any);
      if (value === 'dark') {
        setColorScheme('dark');
      } else if (value === 'light') {
        setColorScheme('light');
      }
    },
    [setTheme, setColorScheme]
  );

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const updateSetting = useCallback((key: string, value: any) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    setSettings(s => (s ? ({ ...s, [key]: value } as any) : null));
  }, []);

  if (loading) {
    return <Text>Loading settings...</Text>;
  }

  if (!settings) {
    return <Text c="red">Failed to load settings</Text>;
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Text c={dark ? '#fff' : '#000'} fw={700} size="xl">
          Settings
        </Text>
        {hasChanges && (
          <Badge color="yellow" variant="filled">
            Unsaved Changes
          </Badge>
        )}
      </Group>

      {saveError && (
        <Paper withBorder bg="#2c1b1b" p="sm" style={{ borderColor: '#c92a2a' }}>
          <Text c="red">{saveError}</Text>
        </Paper>
      )}

      {saveSuccess && (
        <Paper withBorder bg="#1b2c1b" p="sm" style={{ borderColor: '#2ac92a' }}>
          <Text c="green">Settings saved successfully!</Text>
        </Paper>
      )}

      {/* Downloads */}
      <DownloadsSection
        dark={dark}
        handleBrowseDownloadPath={handleBrowseDownloadPath}
        settings={settings}
        updateSetting={updateSetting}
      />

      {/* Download Presets */}
      <PresetsSection dark={dark} settings={settings} updateSetting={updateSetting} />

      {/* UI */}
      <UiSection
        dark={dark}
        handleThemeChange={handleThemeChange}
        settings={settings}
        toggleSidebar={toggleSidebar}
        updateSetting={updateSetting}
      />

      {/* Player */}
      <PlayerSection dark={dark} settings={settings} updateSetting={updateSetting} />

      {/* Logging */}
      <LoggingSection
        dark={dark}
        handleBrowseLogExportPath={handleBrowseLogExportPath}
        handleBrowseLogPath={handleBrowseLogPath}
        settings={settings}
        updateSetting={updateSetting}
      />

      {/* FFmpeg */}
      <FfmpegSection
        dark={dark}
        ffmpegStatus={ffmpegStatus}
        getSelectedFfmpegInfo={getSelectedFfmpegInfo}
        handleBrowseFFmpeg={handleBrowseFFmpeg}
        loadingFfmpeg={loadingFfmpeg}
        refreshFfmpegStatus={refreshFfmpegStatus}
        settings={settings}
        updateSetting={updateSetting}
      />

      {/* Cache Management */}
      <CacheSection confirm={confirm} dark={dark} error={error} success={success} />

      {/* Actions */}
      <ActionsSection
        handleReset={handleReset}
        handleSave={handleSave}
        hasChanges={hasChanges}
        saving={saving}
      />
    </Stack>
  );
}
