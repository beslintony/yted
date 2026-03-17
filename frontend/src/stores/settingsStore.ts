import { create } from 'zustand';

import { GetSettings, SaveSettings } from '../../wailsjs/go/app/App';
import { config } from '../../wailsjs/go/models';
import { DEFAULT_SETTINGS, DownloadPreset, QualityOption, ThemeMode, UserSettings } from '../types';

interface SettingsState extends UserSettings {
  isLoading: boolean;
  error: string | null;

  // Actions
  setDownloadPath: (path: string) => void;
  setMaxConcurrentDownloads: (count: number) => void;
  setDefaultQuality: (quality: QualityOption) => void;
  setFilenameTemplate: (template: string) => void;
  setTheme: (theme: ThemeMode) => void;
  setAccentColor: (color: string) => void;
  toggleSidebar: () => void;
  setDefaultVolume: (volume: number) => void;
  setRememberPosition: (remember: boolean) => void;
  setSpeedLimit: (limit: number | null) => void;
  setProxyUrl: (url: string | null) => void;
  setLogPath: (path: string) => void;
  setLogExportPath: (path: string) => void;
  setMaxLogSessions: (count: number) => void;
  addDownloadPreset: (preset: Omit<DownloadPreset, 'id'>) => void;
  removeDownloadPreset: (id: string) => void;
  updateDownloadPreset: (id: string, preset: Partial<DownloadPreset>) => void;
  loadSettings: () => Promise<void>;
  saveSettings: () => Promise<void>;
  resetToDefaults: () => void;
}

export const useSettingsStore = create<SettingsState>((set, get) => ({
  ...DEFAULT_SETTINGS,
  isLoading: false,
  error: null,

  setDownloadPath: path => set({ downloadPath: path }),

  setMaxConcurrentDownloads: count =>
    set({ maxConcurrentDownloads: Math.max(1, Math.min(10, count)) }),

  setDefaultQuality: quality => set({ defaultQuality: quality }),

  setFilenameTemplate: template => set({ filenameTemplate: template }),

  setTheme: theme => set({ theme }),

  setAccentColor: color => set({ accentColor: color }),

  toggleSidebar: () => set(state => ({ sidebarCollapsed: !state.sidebarCollapsed })),

  setDefaultVolume: volume => set({ defaultVolume: Math.max(0, Math.min(100, volume)) }),

  setRememberPosition: remember => set({ rememberPosition: remember }),

  setSpeedLimit: limit => set({ speedLimitKbps: limit }),

  setProxyUrl: url => set({ proxyUrl: url }),

  setLogPath: path => set({ logPath: path }),

  setLogExportPath: path => set({ logExportPath: path }),

  setMaxLogSessions: count => set({ maxLogSessions: Math.max(1, Math.min(100, count)) }),

  addDownloadPreset: preset => {
    const id = crypto.randomUUID();
    set(state => ({
      downloadPresets: [...state.downloadPresets, { ...preset, id }],
    }));
  },

  removeDownloadPreset: id => {
    set(state => ({
      downloadPresets: state.downloadPresets.filter(p => p.id !== id),
    }));
  },

  updateDownloadPreset: (id, updates) => {
    set(state => ({
      downloadPresets: state.downloadPresets.map(p => (p.id === id ? { ...p, ...updates } : p)),
    }));
  },

  loadSettings: async () => {
    set({ isLoading: true, error: null });
    try {
      const backendSettings = await GetSettings();
      if (backendSettings) {
        // Map snake_case backend fields to camelCase frontend fields
        const mappedSettings: Partial<UserSettings> = {
          downloadPath: backendSettings.download_path,
          maxConcurrentDownloads: backendSettings.max_concurrent_downloads,
          defaultQuality: backendSettings.default_quality as QualityOption,
          filenameTemplate: backendSettings.filename_template,
          theme: backendSettings.theme as ThemeMode,
          accentColor: backendSettings.accent_color,
          sidebarCollapsed: backendSettings.sidebar_collapsed,
          defaultVolume: backendSettings.default_volume,
          rememberPosition: backendSettings.remember_position,
          speedLimitKbps: backendSettings.speed_limit_kbps ?? null,
          proxyUrl: backendSettings.proxy_url ?? null,
          logPath: backendSettings.log_path,
          logExportPath: backendSettings.log_export_path,
          maxLogSessions: backendSettings.max_log_sessions,
          downloadPresets: (backendSettings.download_presets || []).map(
            (p: config.DownloadPreset) => ({
              id: p.id,
              name: p.name,
              format: p.format,
              quality: p.quality as QualityOption,
              extension: p.extension,
            })
          ),
        };
        set({ ...mappedSettings, isLoading: false });
      } else {
        set({ isLoading: false });
      }
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load settings',
        isLoading: false,
      });
    }
  },

  saveSettings: async () => {
    set({ isLoading: true, error: null });
    try {
      const state = get();
      // Create backend config object (snake_case)
      const settings = config.Config.createFrom({
        download_path: state.downloadPath,
        max_concurrent_downloads: state.maxConcurrentDownloads,
        default_quality: state.defaultQuality,
        filename_template: state.filenameTemplate,
        theme: state.theme,
        accent_color: state.accentColor,
        sidebar_collapsed: state.sidebarCollapsed,
        default_volume: state.defaultVolume,
        remember_position: state.rememberPosition,
        speed_limit_kbps: state.speedLimitKbps ?? undefined,
        proxy_url: state.proxyUrl ?? undefined,
        log_path: state.logPath,
        log_export_path: state.logExportPath,
        max_log_sessions: state.maxLogSessions,
        download_presets: state.downloadPresets.map(p => ({
          id: p.id,
          name: p.name,
          format: p.format,
          quality: p.quality,
          extension: p.extension,
        })),
      });
      await SaveSettings(settings);
      set({ isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to save settings',
        isLoading: false,
      });
    }
  },

  resetToDefaults: () => set({ ...DEFAULT_SETTINGS }),
}));
