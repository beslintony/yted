import { create } from 'zustand';
import { Video } from '../types';

interface LibraryState {
  videos: Video[];
  searchQuery: string;
  sortBy: 'date' | 'title' | 'channel' | 'duration';
  sortOrder: 'asc' | 'desc';
  viewMode: 'grid' | 'list';
  isLoading: boolean;
  error: string | null;
  
  // Actions
  setVideos: (videos: Video[]) => void;
  addVideo: (video: Video) => void;
  removeVideo: (id: string) => void;
  updateVideo: (id: string, updates: Partial<Video>) => void;
  updateWatchPosition: (id: string, position: number) => void;
  incrementWatchCount: (id: string) => void;
  setSearchQuery: (query: string) => void;
  setSortBy: (sortBy: 'date' | 'title' | 'channel' | 'duration') => void;
  toggleSortOrder: () => void;
  setViewMode: (mode: 'grid' | 'list') => void;
  getFilteredVideos: () => Video[];
  loadLibrary: () => Promise<void>;
}

export const useLibraryStore = create<LibraryState>((set, get) => ({
  videos: [],
  searchQuery: '',
  sortBy: 'date',
  sortOrder: 'desc',
  viewMode: 'grid',
  isLoading: false,
  error: null,

  setVideos: (videos) => set({ videos }),

  addVideo: (video) => {
    set((state) => ({
      videos: [video, ...state.videos],
    }));
  },

  removeVideo: (id) => {
    set((state) => ({
      videos: state.videos.filter((v) => v.id !== id),
    }));
  },

  updateVideo: (id, updates) => {
    set((state) => ({
      videos: state.videos.map((v) =>
        v.id === id ? { ...v, ...updates } : v
      ),
    }));
  },

  updateWatchPosition: (id, position) => {
    set((state) => ({
      videos: state.videos.map((v) =>
        v.id === id ? { ...v, watchPosition: position } : v
      ),
    }));
  },

  incrementWatchCount: (id) => {
    set((state) => ({
      videos: state.videos.map((v) =>
        v.id === id ? { ...v, watchCount: v.watchCount + 1 } : v
      ),
    }));
  },

  setSearchQuery: (query) => set({ searchQuery: query }),

  setSortBy: (sortBy) => set({ sortBy }),

  toggleSortOrder: () => {
    set((state) => ({
      sortOrder: state.sortOrder === 'asc' ? 'desc' : 'asc',
    }));
  },

  setViewMode: (mode) => set({ viewMode: mode }),

  getFilteredVideos: () => {
    const { videos, searchQuery, sortBy, sortOrder } = get();
    
    let filtered = videos;
    
    // Apply search filter
    if (searchQuery) {
      const query = searchQuery.toLowerCase();
      filtered = videos.filter(
        (v) =>
          v.title.toLowerCase().includes(query) ||
          v.channel.toLowerCase().includes(query)
      );
    }
    
    // Apply sorting
    filtered = [...filtered].sort((a, b) => {
      let comparison = 0;
      switch (sortBy) {
        case 'title':
          comparison = a.title.localeCompare(b.title);
          break;
        case 'channel':
          comparison = a.channel.localeCompare(b.channel);
          break;
        case 'duration':
          comparison = a.duration - b.duration;
          break;
        case 'date':
        default:
          comparison = a.downloadedAt - b.downloadedAt;
          break;
      }
      return sortOrder === 'asc' ? comparison : -comparison;
    });
    
    return filtered;
  },

  loadLibrary: async () => {
    set({ isLoading: true, error: null });
    try {
      // TODO: Load from Go backend
      // const videos = await GetVideos();
      // set({ videos, isLoading: false });
      set({ isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load library',
        isLoading: false,
      });
    }
  },
}));
