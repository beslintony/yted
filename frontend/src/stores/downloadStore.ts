import { create } from 'zustand';
import { Download, DownloadStatus, VideoInfo, VideoFormat } from '../types';

interface DownloadState {
  downloads: Download[];
  isLoading: boolean;
  error: string | null;
  
  // Actions
  addDownload: (url: string, info?: VideoInfo, format?: VideoFormat) => string;
  removeDownload: (id: string) => void;
  startDownload: (id: string) => void;
  pauseDownload: (id: string) => void;
  resumeDownload: (id: string) => void;
  retryDownload: (id: string) => void;
  updateProgress: (id: string, progress: number) => void;
  completeDownload: (id: string) => void;
  failDownload: (id: string, error: string) => void;
  clearCompleted: () => void;
  clearAll: () => void;
  getActiveDownloads: () => Download[];
  getPendingDownloads: () => Download[];
  getCompletedDownloads: () => Download[];
}

export const useDownloadStore = create<DownloadState>((set, get) => ({
  downloads: [],
  isLoading: false,
  error: null,

  addDownload: (url, info, format) => {
    const id = crypto.randomUUID();
    const newDownload: Download = {
      id,
      url,
      status: 'pending',
      progress: 0,
      title: info?.title,
      channel: info?.channel,
      thumbnail: info?.thumbnail,
      formatId: format?.formatId,
      quality: format?.quality,
      createdAt: Date.now(),
    };
    set((state) => ({
      downloads: [newDownload, ...state.downloads],
    }));
    return id;
  },

  removeDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.filter((d) => d.id !== id),
    }));
  },

  startDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'downloading', startedAt: Date.now() } : d
      ),
    }));
  },

  pauseDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'paused' } : d
      ),
    }));
  },

  resumeDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'downloading' } : d
      ),
    }));
  },

  retryDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id
          ? { ...d, status: 'pending', progress: 0, errorMessage: undefined }
          : d
      ),
    }));
  },

  updateProgress: (id, progress) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, progress: Math.min(100, Math.max(0, progress)) } : d
      ),
    }));
  },

  completeDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id
          ? { ...d, status: 'completed', progress: 100, completedAt: Date.now() }
          : d
      ),
    }));
  },

  failDownload: (id, error) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'error', errorMessage: error } : d
      ),
    }));
  },

  clearCompleted: () => {
    set((state) => ({
      downloads: state.downloads.filter((d) => d.status !== 'completed'),
    }));
  },

  clearAll: () => {
    set({ downloads: [] });
  },


}));
