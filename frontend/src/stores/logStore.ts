import { create } from 'zustand';
import { ExportLogs } from '../../wailsjs/go/app/App';

export type LogLevel = 'DEBUG' | 'INFO' | 'WARN' | 'ERROR';

export interface LogEntry {
  timestamp: string;
  level: LogLevel;
  component: string;
  message: string;
  error?: string;
}

interface LogState {
  entries: LogEntry[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  addEntry: (entry: LogEntry) => void;
  setEntries: (entries: LogEntry[]) => void;
  clearLogs: () => void;
  loadLogs: () => Promise<void>;
  exportLogs: () => Promise<void>;
  getLogsByLevel: (level: LogLevel) => LogEntry[];
  getRecentLogs: (count: number) => LogEntry[];
}

export const useLogStore = create<LogState>((set, get) => ({
  entries: [],
  isLoading: false,
  error: null,

  addEntry: (entry) => {
    set((state) => ({
      entries: [...state.entries, entry].slice(-1000), // Keep last 1000 entries
    }));
  },

  setEntries: (entries) => set({ entries }),

  clearLogs: () => set({ entries: [] }),

  loadLogs: async () => {
    set({ isLoading: true, error: null });
    try {
      // This will be populated via Wails events
      set({ isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load logs',
        isLoading: false,
      });
    }
  },

  exportLogs: async () => {
    set({ isLoading: true, error: null });
    try {
      // ExportLogs uses the configured log export path
      await ExportLogs('');
      set({ isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to export logs',
        isLoading: false,
      });
    }
  },

  getLogsByLevel: (level) => {
    return get().entries.filter((e) => e.level === level);
  },

  getRecentLogs: (count) => {
    const { entries } = get();
    return entries.slice(-count);
  },
}));
