/**
 * Development Mock for Wails Runtime
 * 
 * This provides mock implementations of Wails bindings for development
 * when running outside the Wails environment (e.g., `npm run dev`)
 */

const isDev = import.meta.env.DEV;

// Function to check if Wails is available
function hasWails(): boolean {
  return typeof (window as any).go !== 'undefined' && 
         (window as any).go?.app?.App !== undefined;
}

// Function to initialize mocks
function initMocks() {
  if (!isDev) {
    console.log('[DevMock] Not in dev mode, skipping');
    return;
  }
  
  if (hasWails()) {
    console.log('[DevMock] Wails detected, skipping mocks');
    return;
  }
  
  console.log('[DevMock] Wails NOT detected, initializing mocks...');
  
  // Safety check - never overwrite existing go
  if ((window as any).go) {
    console.warn('[DevMock] window.go already exists but looks incomplete, skipping to avoid conflicts');
    return;
  }

  // Mock window.go
  (window as any).go = {
    app: {
      App: {
        // Settings
        GetSettings: async () => ({
          download_path: '/home/user/YTed',
          max_concurrent_downloads: 3,
          default_quality: '1080p',
          theme: 'dark',
        }),
        SaveSettings: async () => {},
        UpdateSetting: async () => {},

        // Library
        ListVideos: async () => [
          {
            id: 'test-video-1',
            youtube_id: 'dQw4w9WgXcQ',
            title: 'Test Video 1',
            channel: 'Test Channel',
            channel_id: 'UC_test',
            description: 'A test video',
            thumbnail_url: 'https://i.ytimg.com/vi/dQw4w9WgXcQ/default.jpg',
            file_path: '/home/user/YTed/test_video_1.mp4',
            file_size: 1024 * 1024 * 50,
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
            file_size: 1024 * 1024 * 30,
            format: 'mp4',
            quality: '720p',
            duration: 180,
            downloaded_at: Date.now() / 1000 - 86400,
            watch_position: 60,
            watch_count: 2,
          },
        ],
        GetVideo: async (id: string) => ({
          id,
          youtube_id: 'test123',
          title: 'Test Video',
          channel: 'Test Channel',
          channel_id: 'UC_test',
          description: 'A test video',
          thumbnail_url: 'https://i.ytimg.com/vi/test123/default.jpg',
          file_path: `/home/user/YTed/test_video_${id}.mp4`,
          file_size: 1024 * 1024 * 50,
          format: 'mp4',
          quality: '1080p',
          duration: 300,
          downloaded_at: Date.now() / 1000,
          watch_position: 0,
          watch_count: 0,
        }),
        DeleteVideo: async () => {},

        // Downloads
        GetDownloads: async () => [],
        GetDownloadQueue: async () => [],
        AddDownload: async () => 'new-download-id',
        CancelDownload: async () => {},
        PauseDownload: async () => {},
        ResumeDownload: async () => {},
        RetryDownload: async () => {},
        ClearCompletedDownloads: async () => {},

        // Version
        GetVersion: async () => 'dev',
        GetAppName: async () => 'YTed',

        // FFmpeg
        CheckFFmpegWithGuidance: async () => ({
          installed: true,
          version: '4.4.2',
          path: '/usr/bin/ffmpeg',
          can_auto_install: false,
          install_method: 'package_manager',
          install_command: '',
          install_guide: '',
          download_url: '',
          requires_admin: false,
        }),

        // Editor
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
        GetEditJobStatus: async () => null,
        CancelEditJob: async () => {},
        GetVideoFile: async () => {
          throw new Error('Video file loading not available in dev mode');
        },
        GetPreviewImage: async () => '',
        PreviewEdit: async () => '',

        // Dialogs
        ShowOpenDirectoryDialog: async () => '',
        ShowSaveDialog: async () => '',
        ShowError: async () => {},
      },
    },
  };

  // Mock window.runtime with all common functions
  (window as any).runtime = {
    EventsOn: () => () => {},
    EventsOnMultiple: () => () => {},
    EventsOff: () => {},
    EventsOffAll: () => {},
    EventsEmit: () => {},
    BrowserOpenURL: (url: string) => window.open(url, '_blank'),
    Environment: async () => ({
      OS: 'linux',
      Platform: 'desktop',
      Arch: 'amd64',
    }),
  };

  console.log('[DevMock] Wails mocks initialized');
}

// Initialize mocks immediately
initMocks();
