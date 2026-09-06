import { create } from 'zustand';

import { Download, DownloadStatus, VideoFormat, VideoInfo } from '../types';

interface DownloadState {
  downloads: Download[];
  isLoading: boolean;
  error: string | null;

  // Actions
  addDownload: (
    url: string,
    info?: VideoInfo,
    format?: VideoFormat,
    existingId?: string
  ) => string | null;
  addDownloads: (items: Download[]) => void;
  removeDownload: (id: string) => void;
  startDownload: (id: string) => void;
  pauseDownload: (id: string) => void;
  resumeDownload: (id: string) => void;
  retryDownload: (id: string) => void;
  updateProgress: (id: string, progress: number) => void;
  updateDownloadInfo: (
    id: string,
    info: {
      speed?: string;
      eta?: string;
      size?: string;
      isThrottled?: boolean;
      speedLimit?: string;
    }
  ) => void;
  // Coalesced progress + info update in a single set() (perf: one notify per tick)
  updateProgressInfo: (
    id: string,
    info: {
      progress?: number;
      speed?: string;
      eta?: string;
      size?: string;
      isThrottled?: boolean;
      speedLimit?: string;
    }
  ) => void;
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

    set(state => ({
      downloads: [newDownload, ...state.downloads],
    }));
    return id;
  },

  // Batch-add downloads (e.g. playlist entries) in a single state update
  // to avoid one re-render per item
  addDownloads: items => {
    set(state => {
      const existing = new Set(state.downloads.map(d => d.id));
      const fresh = items.filter(d => !existing.has(d.id));
      if (fresh.length === 0) {
        return state;
      }
      return { downloads: [...fresh.reverse(), ...state.downloads] };
    });
  },

  removeDownload: id => {
    set(state => {
      if (!state.downloads.some(d => d.id === id)) {
        return state;
      }
      return {
        downloads: state.downloads.filter(d => d.id !== id),
      };
    });
  },

  startDownload: id => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || target.status === 'downloading') {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, status: 'downloading' as DownloadStatus, startedAt: Date.now() } : d
        ),
      };
    });
  },

  pauseDownload: id => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || target.status === 'paused') {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, status: 'paused' as DownloadStatus } : d
        ),
      };
    });
  },

  resumeDownload: id => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || target.status === 'downloading') {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, status: 'downloading' as DownloadStatus } : d
        ),
      };
    });
  },

  retryDownload: id => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || (target.status === 'pending' && target.errorMessage === undefined)) {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, status: 'pending' as DownloadStatus, errorMessage: undefined } : d
        ),
      };
    });
  },

  updateProgress: (id, progress) => {
    const clampedProgress = Math.min(100, Math.max(0, progress));
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || target.progress === clampedProgress) {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, progress: clampedProgress } : d
        ),
      };
    });
  },

  updateDownloadInfo: (
    id: string,
    info: {
      speed?: string;
      eta?: string;
      size?: string;
      isThrottled?: boolean;
      speedLimit?: string;
    }
  ) => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target) {
        return state;
      }
      if (
        (info.speed === undefined || target.speed === info.speed) &&
        (info.eta === undefined || target.eta === info.eta) &&
        (info.size === undefined || target.size === info.size) &&
        (info.isThrottled === undefined || target.isThrottled === info.isThrottled) &&
        (info.speedLimit === undefined || target.speedLimit === info.speedLimit)
      ) {
        return state;
      }
      return {
        downloads: state.downloads.map(d => (d.id === id ? { ...d, ...info } : d)),
      };
    });
  },

  // Single-set coalesced progress + info update for the high-frequency
  // download:progress event. Bails out (same state ref) when nothing changes.
  updateProgressInfo: (
    id: string,
    info: {
      progress?: number;
      speed?: string;
      eta?: string;
      size?: string;
      isThrottled?: boolean;
      speedLimit?: string;
    }
  ) => {
    const clampedProgress =
      info.progress === undefined ? undefined : Math.min(100, Math.max(0, info.progress));
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target) {
        return state;
      }
      if (
        (clampedProgress === undefined || target.progress === clampedProgress) &&
        (info.speed === undefined || target.speed === info.speed) &&
        (info.eta === undefined || target.eta === info.eta) &&
        (info.size === undefined || target.size === info.size) &&
        (info.isThrottled === undefined || target.isThrottled === info.isThrottled) &&
        (info.speedLimit === undefined || target.speedLimit === info.speedLimit)
      ) {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id
            ? {
                ...d,
                ...(clampedProgress !== undefined ? { progress: clampedProgress } : null),
                ...(info.speed !== undefined ? { speed: info.speed } : null),
                ...(info.eta !== undefined ? { eta: info.eta } : null),
                ...(info.size !== undefined ? { size: info.size } : null),
                ...(info.isThrottled !== undefined ? { isThrottled: info.isThrottled } : null),
                ...(info.speedLimit !== undefined ? { speedLimit: info.speedLimit } : null),
              }
            : d
        ),
      };
    });
  },

  completeDownload: id => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || (target.status === 'completed' && target.progress === 100)) {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id
            ? {
                ...d,
                status: 'completed' as DownloadStatus,
                progress: 100,
                completedAt: Date.now(),
              }
            : d
        ),
      };
    });
  },

  failDownload: (id, error) => {
    set(state => {
      const target = state.downloads.find(d => d.id === id);
      if (!target || (target.status === 'error' && target.errorMessage === error)) {
        return state;
      }
      return {
        downloads: state.downloads.map(d =>
          d.id === id ? { ...d, status: 'error' as DownloadStatus, errorMessage: error } : d
        ),
      };
    });
  },

  clearCompleted: () => {
    set(state => {
      if (!state.downloads.some(d => d.status === 'completed')) {
        return state;
      }
      return {
        downloads: state.downloads.filter(d => d.status !== 'completed'),
      };
    });
  },

  clearAll: () => {
    set(state => {
      if (state.downloads.length === 0) {
        return state;
      }
      return { downloads: [] };
    });
  },

  setDownloads: downloads => {
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
    return get().downloads.some(d => d.id === id);
  },
}));
