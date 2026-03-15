import { create } from 'zustand';
import { Video } from '../types';
import { ListVideos } from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';

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
      const options = new app.ListVideosOptions({
        search: '',
        channel: '',
        sort_by: 'date',
        sort_desc: true,
        limit: 1000,
        offset: 0,
      });
      const backendVideos = await ListVideos(options);
      
      // Map backend VideoResult to frontend Video type
      const videos: Video[] = (backendVideos || []).map((v: app.VideoResult) => ({
        id: v.id,
        youtubeId: v.youtube_id,
        title: v.title,
        channel: v.channel,
        channelId: v.channel_id,
        duration: v.duration,
        description: v.description,
        thumbnailUrl: v.thumbnail_url,
        filePath: v.file_path,
        fileSize: v.file_size,
        format: v.format,
        quality: v.quality,
        downloadedAt: v.downloaded_at,
        watchPosition: v.watch_position,
        watchCount: v.watch_count,
      }));
      
      set({ videos, isLoading: false });
    } catch (err) {
      set({
        error: err instanceof Error ? err.message : 'Failed to load library',
        isLoading: false,
      });
    }
  },
}));
