import { beforeEach, describe, expect, it } from 'vitest';

import { useDownloadStore } from './downloadStore';

describe('downloadStore', () => {
  beforeEach(() => {
    useDownloadStore.setState({ downloads: [], isLoading: false, error: null });
  });

  it('should add a download', () => {
    const { addDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    const downloads = useDownloadStore.getState().downloads;
    expect(downloads).toHaveLength(1);
    expect(downloads[0].url).toBe('https://youtube.com/watch?v=test');
    expect(downloads[0].status).toBe('pending');
    expect(downloads[0].id).toBe(id);
  });

  it('should remove a download', () => {
    const { addDownload, removeDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    removeDownload(id);
    expect(useDownloadStore.getState().downloads).toHaveLength(0);
  });

  it('should start a download', () => {
    const { addDownload, startDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    startDownload(id);
    const download = useDownloadStore.getState().downloads[0];
    expect(download.status).toBe('downloading');
    expect(download.startedAt).toBeDefined();
  });

  it('should pause and resume a download', () => {
    const { addDownload, startDownload, pauseDownload, resumeDownload } =
      useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    startDownload(id);
    pauseDownload(id);
    expect(useDownloadStore.getState().downloads[0].status).toBe('paused');

    resumeDownload(id);
    expect(useDownloadStore.getState().downloads[0].status).toBe('downloading');
  });

  it('should update progress', () => {
    const { addDownload, updateProgress } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    updateProgress(id, 50);
    expect(useDownloadStore.getState().downloads[0].progress).toBe(50);
  });

  it('should clamp progress between 0-100', () => {
    const { addDownload, updateProgress } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    updateProgress(id, 150);
    expect(useDownloadStore.getState().downloads[0].progress).toBe(100);

    updateProgress(id, -20);
    expect(useDownloadStore.getState().downloads[0].progress).toBe(0);
  });

  it('should complete a download', () => {
    const { addDownload, startDownload, completeDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    startDownload(id);
    completeDownload(id);

    const download = useDownloadStore.getState().downloads[0];
    expect(download.status).toBe('completed');
    expect(download.progress).toBe(100);
    expect(download.completedAt).toBeDefined();
  });

  it('should fail a download', () => {
    const { addDownload, startDownload, failDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    startDownload(id);
    failDownload(id, 'Network error');

    const download = useDownloadStore.getState().downloads[0];
    expect(download.status).toBe('error');
    expect(download.errorMessage).toBe('Network error');
  });

  it('should retry a failed download', () => {
    const { addDownload, startDownload, failDownload, retryDownload } = useDownloadStore.getState();
    const id = addDownload('https://youtube.com/watch?v=test');

    startDownload(id);
    failDownload(id, 'Network error');
    retryDownload(id);

    const download = useDownloadStore.getState().downloads[0];
    expect(download.status).toBe('pending');
    expect(download.progress).toBe(0);
    expect(download.errorMessage).toBeUndefined();
  });

  it('should clear completed downloads', () => {
    const { addDownload, completeDownload, clearCompleted } = useDownloadStore.getState();

    const id1 = addDownload('https://youtube.com/watch?v=test1');
    const id2 = addDownload('https://youtube.com/watch?v=test2');

    completeDownload(id1);

    clearCompleted();

    const downloads = useDownloadStore.getState().downloads;
    expect(downloads).toHaveLength(1);
    expect(downloads[0].id).toBe(id2);
  });

  it('should filter downloads by status', () => {
    const { addDownload, startDownload, completeDownload } = useDownloadStore.getState();

    addDownload('https://youtube.com/watch?v=test1');
    addDownload('https://youtube.com/watch?v=test2');
    addDownload('https://youtube.com/watch?v=test3');

    const downloads = useDownloadStore.getState().downloads;
    startDownload(downloads[0].id);
    completeDownload(downloads[2].id);

    const updatedDownloads = useDownloadStore.getState().downloads;
    expect(updatedDownloads.filter(d => d.status === 'downloading')).toHaveLength(1);
    expect(updatedDownloads.filter(d => d.status === 'pending')).toHaveLength(1);
    expect(updatedDownloads.filter(d => d.status === 'completed')).toHaveLength(1);
  });
});
