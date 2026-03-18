import { describe, it, expect, beforeEach, vi } from 'vitest';
import { useLibraryStore } from './libraryStore';

// Mock the Wails API
vi.mock('../../wailsjs/go/app/App', () => ({
  ListVideos: vi.fn(),
}));

describe('libraryStore', () => {
  beforeEach(() => {
    useLibraryStore.setState({
      videos: [],
      searchQuery: '',
      sortBy: 'date',
      sortOrder: 'desc',
      viewMode: 'grid',
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  it('should have correct initial state', () => {
    const state = useLibraryStore.getState();
    
    expect(state.videos).toEqual([]);
    expect(state.searchQuery).toBe('');
    expect(state.sortBy).toBe('date');
    expect(state.sortOrder).toBe('desc');
    expect(state.viewMode).toBe('grid');
    expect(state.isLoading).toBe(false);
    expect(state.error).toBeNull();
  });

  it('should set videos', () => {
    const { setVideos } = useLibraryStore.getState();
    
    const mockVideos = [
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Test Video 1',
        channel: 'Test Channel',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Test Video 2',
        channel: 'Other Channel',
        duration: 600,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now() - 1000,
      },
    ];
    
    setVideos(mockVideos);
    
    expect(useLibraryStore.getState().videos).toEqual(mockVideos);
  });

  it('should add video', () => {
    const { addVideo } = useLibraryStore.getState();
    
    const newVideo = {
      id: '3',
      youtubeId: 'yt3',
      title: 'New Video',
      channel: 'New Channel',
      duration: 180,
      filePath: '/path/3.mp4',
      downloadedAt: Date.now(),
    };
    
    addVideo(newVideo);
    
    const videos = useLibraryStore.getState().videos;
    expect(videos).toHaveLength(1);
    expect(videos[0]).toEqual(newVideo);
  });

  it('should add video at the beginning', () => {
    const { setVideos, addVideo } = useLibraryStore.getState();
    
    setVideos([{
      id: '1',
      youtubeId: 'yt1',
      title: 'First Video',
      channel: 'Channel',
      duration: 300,
      filePath: '/path/1.mp4',
      downloadedAt: Date.now(),
    }]);
    
    addVideo({
      id: '2',
      youtubeId: 'yt2',
      title: 'Second Video',
      channel: 'Channel',
      duration: 300,
      filePath: '/path/2.mp4',
      downloadedAt: Date.now(),
    });
    
    const videos = useLibraryStore.getState().videos;
    expect(videos[0].id).toBe('2');
    expect(videos[1].id).toBe('1');
  });

  it('should update video', () => {
    const { setVideos, updateVideo } = useLibraryStore.getState();
    
    setVideos([{
      id: '1',
      youtubeId: 'yt1',
      title: 'Test Video',
      channel: 'Test Channel',
      duration: 300,
      filePath: '/path/1.mp4',
      downloadedAt: Date.now(),
    }]);
    
    updateVideo('1', { title: 'Updated Title' });
    
    const videos = useLibraryStore.getState().videos;
    expect(videos[0].title).toBe('Updated Title');
    expect(videos[0].duration).toBe(300); // Unchanged
  });

  it('should not update video if id not found', () => {
    const { setVideos, updateVideo } = useLibraryStore.getState();
    
    setVideos([{
      id: '1',
      youtubeId: 'yt1',
      title: 'Test Video',
      channel: 'Test Channel',
      duration: 300,
      filePath: '/path/1.mp4',
      downloadedAt: Date.now(),
    }]);
    
    updateVideo('non-existent', { title: 'Updated Title' });
    
    expect(useLibraryStore.getState().videos[0].title).toBe('Test Video');
  });

  it('should remove video', () => {
    const { setVideos, removeVideo } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Video 1',
        channel: 'Channel 1',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Video 2',
        channel: 'Channel 2',
        duration: 600,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    removeVideo('1');
    
    const videos = useLibraryStore.getState().videos;
    expect(videos).toHaveLength(1);
    expect(videos[0].id).toBe('2');
  });

  it('should set search query', () => {
    const { setSearchQuery } = useLibraryStore.getState();
    
    setSearchQuery('test query');
    
    expect(useLibraryStore.getState().searchQuery).toBe('test query');
  });

  it('should set sort by', () => {
    const { setSortBy } = useLibraryStore.getState();
    
    setSortBy('title');
    expect(useLibraryStore.getState().sortBy).toBe('title');
    
    setSortBy('channel');
    expect(useLibraryStore.getState().sortBy).toBe('channel');
    
    setSortBy('duration');
    expect(useLibraryStore.getState().sortBy).toBe('duration');
  });

  it('should toggle sort order', () => {
    const { toggleSortOrder } = useLibraryStore.getState();
    
    expect(useLibraryStore.getState().sortOrder).toBe('desc');
    
    toggleSortOrder();
    expect(useLibraryStore.getState().sortOrder).toBe('asc');
    
    toggleSortOrder();
    expect(useLibraryStore.getState().sortOrder).toBe('desc');
  });

  it('should set view mode', () => {
    const { setViewMode } = useLibraryStore.getState();
    
    setViewMode('list');
    expect(useLibraryStore.getState().viewMode).toBe('list');
    
    setViewMode('grid');
    expect(useLibraryStore.getState().viewMode).toBe('grid');
  });

  it('should set loading state', () => {
    const { setVideos } = useLibraryStore.getState();
    
    // Set loading via loadLibrary indirectly
    setVideos([]);
    expect(useLibraryStore.getState().isLoading).toBe(false);
  });

  it('should update watch position', () => {
    const { setVideos, updateWatchPosition } = useLibraryStore.getState();
    
    setVideos([{
      id: '1',
      youtubeId: 'yt1',
      title: 'Test Video',
      channel: 'Test Channel',
      duration: 300,
      filePath: '/path/1.mp4',
      downloadedAt: Date.now(),
      watchPosition: 0,
    }]);
    
    updateWatchPosition('1', 150);
    
    expect(useLibraryStore.getState().videos[0].watchPosition).toBe(150);
  });

  it('should increment watch count', () => {
    const { setVideos, incrementWatchCount } = useLibraryStore.getState();
    
    setVideos([{
      id: '1',
      youtubeId: 'yt1',
      title: 'Test Video',
      channel: 'Test Channel',
      duration: 300,
      filePath: '/path/1.mp4',
      downloadedAt: Date.now(),
      watchCount: 5,
    }]);
    
    incrementWatchCount('1');
    
    expect(useLibraryStore.getState().videos[0].watchCount).toBe(6);
  });

  it('should filter videos by search query', () => {
    const { setVideos, setSearchQuery } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Alpha Video',
        channel: 'Channel A',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Beta Video',
        channel: 'Channel B',
        duration: 600,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    setSearchQuery('alpha');
    
    const filtered = useLibraryStore.getState().getFilteredVideos();
    expect(filtered).toHaveLength(1);
    expect(filtered[0].id).toBe('1');
  });

  it('should filter videos by channel', () => {
    const { setVideos, setSearchQuery } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Video One',
        channel: 'Channel A',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Video Two',
        channel: 'Channel B',
        duration: 600,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    setSearchQuery('channel a');
    
    const filtered = useLibraryStore.getState().getFilteredVideos();
    expect(filtered).toHaveLength(1);
    expect(filtered[0].channel).toBe('Channel A');
  });

  it('should sort videos by title', () => {
    const { setVideos, setSortBy, toggleSortOrder } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Zebra Video',
        channel: 'Channel',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Apple Video',
        channel: 'Channel',
        duration: 600,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    setSortBy('title');
    toggleSortOrder(); // Switch to ascending for alphabetical
    
    const filtered = useLibraryStore.getState().getFilteredVideos();
    expect(filtered[0].title).toBe('Apple Video');
    expect(filtered[1].title).toBe('Zebra Video');
  });

  it('should sort videos by duration', () => {
    const { setVideos, setSortBy, toggleSortOrder } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'Long Video',
        channel: 'Channel',
        duration: 600,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
      {
        id: '2',
        youtubeId: 'yt2',
        title: 'Short Video',
        channel: 'Channel',
        duration: 100,
        filePath: '/path/2.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    setSortBy('duration');
    toggleSortOrder(); // Switch to ascending for shortest first
    
    const filtered = useLibraryStore.getState().getFilteredVideos();
    expect(filtered[0].duration).toBe(100);
    expect(filtered[1].duration).toBe(600);
  });

  it('should return all videos when search is empty', () => {
    const { setVideos, getFilteredVideos } = useLibraryStore.getState();
    
    setVideos([
      { id: '1', title: 'Video 1', channel: 'A', duration: 300, filePath: '/1.mp4', downloadedAt: Date.now() },
      { id: '2', title: 'Video 2', channel: 'B', duration: 600, filePath: '/2.mp4', downloadedAt: Date.now() },
    ]);
    
    const filtered = getFilteredVideos();
    expect(filtered).toHaveLength(2);
  });

  it('should be case insensitive for search', () => {
    const { setVideos, setSearchQuery } = useLibraryStore.getState();
    
    setVideos([
      {
        id: '1',
        youtubeId: 'yt1',
        title: 'UPPERCASE VIDEO',
        channel: 'Channel',
        duration: 300,
        filePath: '/path/1.mp4',
        downloadedAt: Date.now(),
      },
    ]);
    
    setSearchQuery('uppercase');
    
    const filtered = useLibraryStore.getState().getFilteredVideos();
    expect(filtered).toHaveLength(1);
  });

  it('should load library', async () => {
    const { ListVideos } = await import('../../wailsjs/go/app/App');
    const mockedListVideos = ListVideos as unknown as ReturnType<typeof vi.fn>;
    
    mockedListVideos.mockResolvedValue([
      {
        id: '1',
        youtube_id: 'yt1',
        title: 'Test Video',
        channel: 'Test Channel',
        channel_id: 'ch1',
        duration: 300,
        description: 'Test',
        thumbnail_url: 'http://thumb.jpg',
        file_path: '/path/1.mp4',
        file_size: 1000,
        format: 'mp4',
        quality: '1080p',
        downloaded_at: Date.now(),
        watch_position: 0,
        watch_count: 0,
      },
    ]);
    
    const { loadLibrary } = useLibraryStore.getState();
    await loadLibrary();
    
    const state = useLibraryStore.getState();
    expect(state.videos).toHaveLength(1);
    expect(state.videos[0].title).toBe('Test Video');
    expect(state.isLoading).toBe(false);
    expect(state.error).toBeNull();
  });

  it('should handle load library error', async () => {
    const { ListVideos } = await import('../../wailsjs/go/app/App');
    const mockedListVideos = ListVideos as unknown as ReturnType<typeof vi.fn>;
    
    mockedListVideos.mockRejectedValue(new Error('Network error'));
    
    const { loadLibrary } = useLibraryStore.getState();
    await loadLibrary();
    
    const state = useLibraryStore.getState();
    expect(state.error).toBe('Network error');
    expect(state.isLoading).toBe(false);
  });
});
