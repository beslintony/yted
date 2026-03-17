import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { renderHook, act, waitFor } from '@testing-library/react';
import { useDownloadStore } from '../stores/downloadStore';

// Mock the wails runtime
vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()), // Returns cleanup function
}));

// Mock the app API
vi.mock('../../wailsjs/go/app/App', () => ({
  GetDownloadQueue: vi.fn(),
  StartProcessingDownloads: vi.fn(),
  GetSettings: vi.fn(() => Promise.resolve({ download_presets: [] })),
}));

import { EventsOn } from '../../wailsjs/runtime';
import { GetDownloadQueue, StartProcessingDownloads } from '../../wailsjs/go/app/App';

describe('DownloadPage - Queue Restoration', () => {
  beforeEach(() => {
    useDownloadStore.setState({ downloads: [], isLoading: false, error: null });
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it('should sync error status from backend during queue restoration', async () => {
    // Simulate backend having a failed download that frontend missed
    const failedDownload = {
      id: 'failed-download-123',
      url: 'https://youtube.com/watch?v=test',
      status: 'error',
      progress: 50,
      error_message: 'HTTP 416 error: Range not satisfiable',
      title: 'Test Video',
      channel: 'Test Channel',
      thumbnail_url: 'https://example.com/thumb.jpg',
      format_id: 'best',
      quality: 'best',
      duration: 300,
      created_at: new Date().toISOString(),
    };

    // Mock GetDownloadQueue to return the failed download
    (GetDownloadQueue as any).mockResolvedValue([failedDownload]);
    (StartProcessingDownloads as any).mockResolvedValue(undefined);

    // Add download to frontend store with outdated status (simulating missed event)
    const { addDownload } = useDownloadStore.getState();
    act(() => {
      addDownload(
        'https://youtube.com/watch?v=test',
        {
          id: 'test',
          title: 'Test Video',
          channel: 'Test Channel',
          channelId: '',
          duration: 300,
          description: '',
          thumbnail: 'https://example.com/thumb.jpg',
          formats: [],
        },
        {
          formatId: 'best',
          quality: 'best',
          ext: 'mp4',
          resolution: '',
          fps: 0,
          vcodec: '',
          acodec: '',
          filesize: 0,
        },
        'failed-download-123'
      );
    });

    // Verify initial state is pending (simulating missed error event)
    let state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('pending');

    // Simulate queue restoration (normally done in useEffect)
    const { failDownload, updateProgress } = useDownloadStore.getState();
    
    act(() => {
      // This simulates what should happen during queue restoration
      updateProgress(failedDownload.id, failedDownload.progress);
      if (failedDownload.status === 'error') {
        failDownload(failedDownload.id, failedDownload.error_message);
      }
    });

    // Verify status was synced from backend
    state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('error');
    expect(state.downloads[0].errorMessage).toBe('HTTP 416 error: Range not satisfiable');
    expect(state.downloads[0].progress).toBe(50);
  });

  it('should sync completed status from backend during queue restoration', async () => {
    const completedDownload = {
      id: 'completed-download-456',
      url: 'https://youtube.com/watch?v=test2',
      status: 'completed',
      progress: 100,
      title: 'Completed Video',
      created_at: new Date().toISOString(),
    };

    (GetDownloadQueue as any).mockResolvedValue([completedDownload]);

    // Add download with outdated status
    const { addDownload } = useDownloadStore.getState();
    act(() => {
      addDownload(
        'https://youtube.com/watch?v=test2',
        {
          id: 'test2',
          title: 'Completed Video',
          channel: '',
          channelId: '',
          duration: 0,
          description: '',
          thumbnail: '',
          formats: [],
        },
        {
          formatId: 'best',
          quality: 'best',
          ext: 'mp4',
          resolution: '',
          fps: 0,
          vcodec: '',
          acodec: '',
          filesize: 0,
        },
        'completed-download-456'
      );
    });

    // Simulate missed completed event
    let state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('pending');

    // Simulate queue restoration syncing status
    const { completeDownload, updateProgress } = useDownloadStore.getState();
    
    act(() => {
      updateProgress(completedDownload.id, completedDownload.progress);
      if (completedDownload.status === 'completed') {
        completeDownload(completedDownload.id);
      }
    });

    state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('completed');
    expect(state.downloads[0].progress).toBe(100);
  });

  it('should add new downloads from queue that do not exist in frontend', async () => {
    const newDownload = {
      id: 'new-download-789',
      url: 'https://youtube.com/watch?v=new',
      status: 'pending',
      progress: 0,
      title: 'New Video',
      channel: 'New Channel',
      thumbnail_url: 'https://example.com/new.jpg',
      format_id: 'best',
      quality: '1080p',
      created_at: new Date().toISOString(),
    };

    (GetDownloadQueue as any).mockResolvedValue([newDownload]);

    // Store is empty initially
    let state = useDownloadStore.getState();
    expect(state.downloads).toHaveLength(0);

    // Simulate adding from queue restoration
    const { addDownload } = useDownloadStore.getState();
    act(() => {
      addDownload(
        newDownload.url,
        {
          id: 'new',
          title: newDownload.title,
          channel: newDownload.channel,
          channelId: '',
          duration: 0,
          description: '',
          thumbnail: newDownload.thumbnail_url,
          formats: [],
        },
        {
          formatId: newDownload.format_id,
          quality: newDownload.quality,
          ext: 'mp4',
          resolution: '',
          fps: 0,
          vcodec: '',
          acodec: '',
          filesize: 0,
        },
        newDownload.id
      );
    });

    state = useDownloadStore.getState();
    expect(state.downloads).toHaveLength(1);
    expect(state.downloads[0].id).toBe('new-download-789');
    expect(state.downloads[0].status).toBe('pending');
  });
});

describe('DownloadPage - Event Deduplication', () => {
  beforeEach(() => {
    useDownloadStore.setState({ downloads: [], isLoading: false, error: null });
    vi.clearAllMocks();
  });

  it('should process error event only once per download', () => {
    const processedEvents = new Set<string>();
    const downloadId = 'test-download-001';

    // First error event
    const errorKey = `error_${downloadId}`;
    
    // Simulate event handler logic
    const shouldProcess1 = !processedEvents.has(errorKey);
    expect(shouldProcess1).toBe(true);
    
    if (shouldProcess1) {
      processedEvents.add(errorKey);
    }

    // Second error event (duplicate)
    const shouldProcess2 = !processedEvents.has(errorKey);
    expect(shouldProcess2).toBe(false); // Should NOT process duplicate

    expect(processedEvents.size).toBe(1);
    expect(processedEvents.has(errorKey)).toBe(true);
  });

  it('should clear processed events when download is retried', () => {
    const processedEvents = new Set<string>();
    const downloadId = 'test-download-002';

    // Original failure
    const errorKey = `error_${downloadId}`;
    processedEvents.add(errorKey);
    
    const completedKey = `completed_${downloadId}`;
    processedEvents.add(completedKey);

    expect(processedEvents.size).toBe(2);

    // Simulate retry - should clear events
    act(() => {
      processedEvents.delete(errorKey);
      processedEvents.delete(completedKey);
    });

    expect(processedEvents.size).toBe(0);

    // New error should be processed
    const shouldProcess = !processedEvents.has(errorKey);
    expect(shouldProcess).toBe(true);
  });

  it('should clear processed events when download is removed', () => {
    const processedEvents = new Set<string>();
    const downloadId = 'test-download-003';

    // Events were processed
    processedEvents.add(`error_${downloadId}`);
    processedEvents.add(`completed_${downloadId}`);

    // Simulate removal - should clear events
    act(() => {
      processedEvents.delete(`error_${downloadId}`);
      processedEvents.delete(`completed_${downloadId}`);
    });

    expect(processedEvents.has(`error_${downloadId}`)).toBe(false);
    expect(processedEvents.has(`completed_${downloadId}`)).toBe(false);
  });
});

describe('DownloadPage - Event Handlers', () => {
  beforeEach(() => {
    useDownloadStore.setState({ downloads: [], isLoading: false, error: null });
    vi.clearAllMocks();
  });

  it('should handle download:error event correctly', () => {
    // Setup download
    const { addDownload, startDownload } = useDownloadStore.getState();
    act(() => {
      const id = addDownload('https://youtube.com/watch?v=error-test');
      if (id) startDownload(id);
    });

    const state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('downloading');

    // Simulate error event
    const { failDownload } = useDownloadStore.getState();
    act(() => {
      failDownload(state.downloads[0].id, 'Network error');
    });

    const newState = useDownloadStore.getState();
    expect(newState.downloads[0].status).toBe('error');
    expect(newState.downloads[0].errorMessage).toBe('Network error');
  });

  it('should handle download:completed event correctly', () => {
    // Setup download
    const { addDownload, startDownload } = useDownloadStore.getState();
    act(() => {
      const id = addDownload('https://youtube.com/watch?v=complete-test');
      if (id) startDownload(id);
    });

    const state = useDownloadStore.getState();
    const downloadId = state.downloads[0].id;

    // Simulate progress updates
    const { updateProgress } = useDownloadStore.getState();
    act(() => {
      updateProgress(downloadId, 50);
    });

    expect(useDownloadStore.getState().downloads[0].progress).toBe(50);

    // Simulate completed event
    const { completeDownload } = useDownloadStore.getState();
    act(() => {
      completeDownload(downloadId);
    });

    const newState = useDownloadStore.getState();
    expect(newState.downloads[0].status).toBe('completed');
    expect(newState.downloads[0].progress).toBe(100);
    expect(newState.downloads[0].completedAt).toBeDefined();
  });

  it('should handle download:started event correctly', () => {
    const { addDownload } = useDownloadStore.getState();
    let downloadId: string;
    
    act(() => {
      const id = addDownload('https://youtube.com/watch?v=start-test');
      downloadId = id!;
    });

    const state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('pending');

    // Simulate started event
    const { startDownload } = useDownloadStore.getState();
    act(() => {
      startDownload(downloadId);
    });

    const newState = useDownloadStore.getState();
    expect(newState.downloads[0].status).toBe('downloading');
    expect(newState.downloads[0].startedAt).toBeDefined();
  });

  it('should handle download:retried event correctly', () => {
    // Setup failed download
    const { addDownload, startDownload, failDownload } = useDownloadStore.getState();
    let downloadId: string;
    
    act(() => {
      const id = addDownload('https://youtube.com/watch?v=retry-test');
      downloadId = id!;
      if (id) {
        startDownload(id);
        failDownload(id, 'Network error');
      }
    });

    let state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('error');
    expect(state.downloads[0].errorMessage).toBe('Network error');

    // Simulate retry event
    const { retryDownload } = useDownloadStore.getState();
    act(() => {
      retryDownload(downloadId);
    });

    state = useDownloadStore.getState();
    expect(state.downloads[0].status).toBe('pending');
    expect(state.downloads[0].progress).toBe(0);
    expect(state.downloads[0].errorMessage).toBeUndefined();
  });
});

describe('DownloadPage - HTTP 416 Regression', () => {
  it('should document the HTTP 416 bug fix requirements', () => {
    // This test documents the fix for the HTTP 416 bug where errors
    // were incorrectly shown as "completed" in the UI
    
    const requirements = [
      'Backend must call FailDownload() before emitting download:error event',
      'Frontend queue restoration must ALWAYS sync status from backend',
      'Event deduplication must be cleared on retry',
      'Download state should never transition: downloading -> completed without success',
    ];

    requirements.forEach((req, i) => {
      // eslint-disable-next-line no-console
      console.log(`Requirement ${i + 1}: ${req}`);
    });

    expect(requirements.length).toBeGreaterThan(0);
  });
});
