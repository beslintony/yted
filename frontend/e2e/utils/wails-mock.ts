/**
 * Wails Runtime Mock Utilities
 * 
 * These utilities help mock the Wails runtime for E2E testing.
 * In a real Wails app, these would be injected by the Wails runtime.
 */

export interface MockVideo {
  id: string;
  youtube_id: string;
  title: string;
  channel: string;
  channel_id: string;
  description: string;
  thumbnail_url: string;
  file_path: string;
  file_size: number;
  format: string;
  quality: string;
  duration: number;
  downloaded_at: number;
  watch_position: number;
  watch_count: number;
}

export interface MockDownload {
  id: string;
  url: string;
  video_id: string;
  title: string;
  status: 'pending' | 'downloading' | 'paused' | 'completed' | 'error';
  progress: number;
  error_message?: string;
}

// Default mock data
export const defaultMocks = {
  videos: [
    {
      id: 'test-video-1',
      youtube_id: 'dQw4w9WgXcQ',
      title: 'Test Video 1',
      channel: 'Test Channel',
      channel_id: 'UC_test',
      description: 'A test video for E2E testing',
      thumbnail_url: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg',
      file_path: '/home/user/YTed/test_video_1.mp4',
      file_size: 1024 * 1024 * 50, // 50MB
      format: 'mp4',
      quality: '1080p',
      duration: 300,
      downloaded_at: Date.now() / 1000,
      watch_position: 0,
      watch_count: 0,
    },
    {
      id: 'test-video-2',
      youtube_id: 'abc123',
      title: 'Test Video 2',
      channel: 'Another Channel',
      channel_id: 'UC_another',
      description: 'Another test video',
      thumbnail_url: 'https://i.ytimg.com/vi/abc123/default.jpg',
      file_path: '/home/user/YTed/test_video_2.mp4',
      file_size: 1024 * 1024 * 30, // 30MB
      format: 'mp4',
      quality: '720p',
      duration: 180,
      downloaded_at: Date.now() / 1000 - 86400, // 1 day ago
      watch_position: 60,
      watch_count: 2,
    },
  ] as MockVideo[],
  
  downloads: [
    {
      id: 'download-1',
      url: 'https://youtube.com/watch?v=test123',
      video_id: 'test-video-1',
      title: 'Downloading Test Video',
      status: 'completed',
      progress: 100,
    },
  ] as MockDownload[],
};

// Mock Wails Go functions
export function createWailsMock(page: any, mocks: typeof defaultMocks = defaultMocks) {
  return page.addInitScript((mocksData: typeof defaultMocks) => {
    // @ts-ignore
    window.go = {
      app: {
        App: {
          // Library functions
          ListVideos: async () => mocksData.videos,
          GetVideo: async (id: string) => mocksData.videos.find(v => v.id === id),
          DeleteVideo: async () => {},
          
          // Download functions
          GetDownloads: async () => mocksData.downloads,
          AddDownload: async () => 'new-download-id',
          CancelDownload: async () => {},
          PauseDownload: async () => {},
          ResumeDownload: async () => {},
          
          // Settings functions
          GetSettings: async () => ({
            download_path: '/home/user/YTed',
            max_concurrent_downloads: 3,
            default_quality: '1080p',
            theme: 'dark',
          }),
          SaveSettings: async () => {},
          
          // FFmpeg functions
          CheckFFmpegWithGuidance: async () => ({
            installed: true,
            version: '4.4.2',
            path: '/usr/bin/ffmpeg',
          }),
          
          // Editor functions
          GetVideoMetadata: async () => ({
            duration: 300,
            width: 1920,
            height: 1080,
            fps: 30,
            bitrate: 5000000,
            codec: 'h264',
            has_audio: true,
            audio_codec: 'aac',
          }),
          ListEditJobs: async () => [],
          SubmitEditJob: async () => 'job-123',
        },
      },
    };
    
    // @ts-ignore
    window.runtime = {
      EventsOn: () => () => {}, // Returns unsubscribe function
      EventsOff: () => {},
      EventsEmit: () => {},
    };
  }, mocks);
}
