import { create } from 'zustand';
import { UserSettings, DEFAULT_SETTINGS, ThemeMode, QualityOption, DownloadPreset } from '../types';

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

  setDownloadPath: (path) => set({ downloadPath: path }),

  setMaxConcurrentDownloads: (count) => set({ maxConcurrentDownloads: Math.max(1, Math.min(10, count)) }),

  setDefaultQuality: (quality) => set({ defaultQuality: quality }),

  setFilenameTemplate: (template) => set({ filenameTemplate: template }),

  setTheme: (theme) => set({ theme }),

  setAccentColor: (color) => set({ accentColor: color }),

  toggleSidebar: () => set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed })),

  setDefaultVolume: (volume) => set({ defaultVolume: Math.max(0, Math.min(100, volume)) }),

  setRememberPosition: (remember) => set({ rememberPosition: remember }),

  setSpeedLimit: (limit) => set({ speedLimitKbps: limit }),

  setProxyUrl: (url) => set({ proxyUrl: url }),

  setLogPath: (path) => set({ logPath: path }),

  setLogExportPath: (path) => set({ logExportPath: path }),

  setMaxLogSessions: (count) => set({ maxLogSessions: Math.max(1, Math.min(100, count)) }),

  addDownloadPreset: (preset) => {
    const id = crypto.randomUUID();
    set((state) => ({
      downloadPresets: [...state.downloadPresets, { ...preset, id }],
    }));
  },

  removeDownloadPreset: (id) => {
    set((state) => ({
      downloadPresets: state.downloadPresets.filter((p) => p.id !== id),
    }));
  },

  updateDownloadPreset: (id, updates) => {
    set((state) => ({
      downloadPresets: state.downloadPresets.map((p) =>
        p.id === id ? { ...p, ...updates } : p
      ),
    }));
  },

  loadSettings: async () => {
    set({ isLoading: true, error: null });
    try {
      // TODO: Load from Go backend
      // const settings = await GetSettings();
      // set({ ...settings, isLoading: false });
      set({ isLoading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to load settings', isLoading: false });
    }
  },

  saveSettings: async () => {
    set({ isLoading: true, error: null });
    try {
      const state = get();
      const settings: UserSettings = {
        downloadPath: state.downloadPath,
        maxConcurrentDownloads: state.maxConcurrentDownloads,
        defaultQuality: state.defaultQuality,
        filenameTemplate: state.filenameTemplate,
        theme: state.theme,
        accentColor: state.accentColor,
        sidebarCollapsed: state.sidebarCollapsed,
        defaultVolume: state.defaultVolume,
        rememberPosition: state.rememberPosition,
        speedLimitKbps: state.speedLimitKbps,
        proxyUrl: state.proxyUrl,
        logPath: state.logPath,
        logExportPath: state.logExportPath,
        maxLogSessions: state.maxLogSessions,
        downloadPresets: state.downloadPresets,
      };
      // TODO: Save to Go backend
      // await SaveSettings(settings);
      set({ isLoading: false });
    } catch (err) {
      set({ error: err instanceof Error ? err.message : 'Failed to save settings', isLoading: false });
    }
  },

  resetToDefaults: () => set({ ...DEFAULT_SETTINGS }),
}));
