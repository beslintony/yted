import { create } from 'zustand';

import { Download, DownloadStatus, VideoInfo, VideoFormat } from '../types';

interface DownloadState {
  downloads: Download[];
  isLoading: boolean;
  error: string | null;

  // Actions
  addDownload: (url: string, info?: VideoInfo, format?: VideoFormat, existingId?: string) => string | null;
  removeDownload: (id: string) => void;
  startDownload: (id: string) => void;
  pauseDownload: (id: string) => void;
  resumeDownload: (id: string) => void;
  retryDownload: (id: string) => void;
  updateProgress: (id: string, progress: number) => void;
  updateDownloadInfo: (id: string, info: { speed?: string; eta?: string; size?: string; isThrottled?: boolean; speedLimit?: string }) => void;
  completeDownload: (id: string) => void;
  failDownload: (id: string, error: string) => void;
  clearCompleted: () => void;
  clearAll: () => void;
  setDownloads: (downloads: Download[]) => void;
  getActiveDownloads: () => Download[];
  getPendingDownloads: () => Download[];
  getCompletedDownloads: () => Download[];
  hasDownload: (id: string) => boolean;
}

export const useDownloadStore = create<DownloadState>((set, get) => ({
  downloads: [],
  isLoading: false,
  error: null,

  addDownload: (url: string, info?: VideoInfo, format?: VideoFormat, existingId?: string) => {
    const id = existingId || crypto.randomUUID();
    
    // Check if download already exists
    if (get().hasDownload(id)) {
      console.warn(`Download with id ${id} already exists, skipping`);
      return null;
    }

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
        d.id === id ? { ...d, status: 'downloading' as DownloadStatus, startedAt: Date.now() } : d
      ),
    }));
  },

  pauseDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'paused' as DownloadStatus } : d
      ),
    }));
  },

  resumeDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'downloading' as DownloadStatus } : d
      ),
    }));
  },

  retryDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id
          ? { ...d, status: 'pending' as DownloadStatus, errorMessage: undefined }
          : d
      ),
    }));
  },

  updateProgress: (id, progress) => {
    const clampedProgress = Math.min(100, Math.max(0, progress));
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, progress: clampedProgress } : d
      ),
    }));
  },

  updateDownloadInfo: (id: string, info: { speed?: string; eta?: string; size?: string }) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, ...info } : d
      ),
    }));
  },

  completeDownload: (id) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id
          ? { ...d, status: 'completed' as DownloadStatus, progress: 100, completedAt: Date.now() }
          : d
      ),
    }));
  },

  failDownload: (id, error) => {
    set((state) => ({
      downloads: state.downloads.map((d) =>
        d.id === id ? { ...d, status: 'error' as DownloadStatus, errorMessage: error } : d
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

  setDownloads: (downloads) => {
    set({ downloads });
  },

  getActiveDownloads: (): Download[] => {
    return get().downloads.filter((d: Download) => d.status === 'downloading');
  },

  getPendingDownloads: (): Download[] => {
    return get().downloads.filter((d: Download) => d.status === 'pending');
  },

  getCompletedDownloads: (): Download[] => {
    return get().downloads.filter((d: Download) => d.status === 'completed');
  },

  hasDownload: (id: string): boolean => {
    return get().downloads.some((d) => d.id === id);
  },
}));
