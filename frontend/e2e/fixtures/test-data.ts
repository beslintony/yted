/**
 * Test Data Fixtures
 * 
 * Reusable test data for E2E tests
 */

export const testVideos = [
  {
    id: 'video-1',
    youtube_id: 'test123',
    title: 'Introduction to Testing',
    channel: 'Test Academy',
    duration: 600,
    format: 'mp4',
    quality: '1080p',
  },
  {
    id: 'video-2', 
    youtube_id: 'test456',
    title: 'Advanced E2E Testing',
    channel: 'Test Academy',
    duration: 1200,
    format: 'mp4',
    quality: '720p',
  },
];

export const testDownloads = [
  {
    id: 'dl-1',
    url: 'https://youtube.com/watch?v=test123',
    title: 'Introduction to Testing',
    status: 'completed',
    progress: 100,
  },
  {
    id: 'dl-2',
    url: 'https://youtube.com/watch?v=test456',
    title: 'Advanced E2E Testing', 
    status: 'downloading',
    progress: 45,
  },
];

export const testSettings = {
  downloadPath: '/home/user/YTed',
  maxConcurrentDownloads: 3,
  defaultQuality: '1080p',
  theme: 'dark' as const,
};
