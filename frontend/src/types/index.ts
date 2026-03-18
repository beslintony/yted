// YTed TypeScript Types

export interface VideoInfo {
  id: string;
  title: string;
  channel: string;
  channelId: string;
  duration: number;
  description: string;
  thumbnail: string;
  formats: VideoFormat[];
}

export interface VideoFormat {
  formatId: string;
  ext: string;
  resolution: string;
  fps: number;
  vcodec: string;
  acodec: string;
  filesize: number;
  quality: string;
}

export type DownloadStatus = 'pending' | 'downloading' | 'paused' | 'completed' | 'error';

export interface Download {
  id: string;
  url: string;
  status: DownloadStatus;
  progress: number;
  title?: string;
  channel?: string;
  thumbnail?: string;
  formatId?: string;
  quality?: string;
  errorMessage?: string;
  speed?: string;
  eta?: string;
  size?: string;
  isThrottled?: boolean;  // true if speed limit is being enforced
  speedLimit?: string;    // formatted limit (e.g., "1.5 MiB/s")
  createdAt: number;
  startedAt?: number;
  completedAt?: number;
}

export interface Video {
  id: string;
  youtubeId: string;
  title: string;
  channel: string;
  channelId: string;
  duration: number;
  description: string;
  thumbnailUrl: string;
  filePath: string;
  fileSize: number;
  format: string;
  quality: string;
  downloadedAt: number;
  watchPosition: number;
  watchCount: number;
}

export type ThemeMode = 'dark' | 'light' | 'auto';
export type QualityOption =
  | 'best'
  | '2160p'
  | '1440p'
  | '1080p'
  | '720p'
  | '480p'
  | '360p'
  | 'audio';

export interface DownloadPreset {
  id: string;
  name: string;
  format: string;
  quality: QualityOption;
  extension: string;
}

export interface UserSettings {
  // Downloads
  downloadPath: string;
  maxConcurrentDownloads: number;
  defaultQuality: QualityOption;
  filenameTemplate: string;

  // UI
  theme: ThemeMode;
  accentColor: string;
  sidebarCollapsed: boolean;

  // Player
  defaultVolume: number;
  rememberPosition: boolean;

  // Network
  speedLimitKbps: number | null;
  proxyUrl: string | null;

  // Logging
  logPath: string; // Internal log storage path (.logs folder)
  logExportPath: string; // Export destination
  maxLogSessions: number; // Number of sessions to keep (default: 10)

  // Presets
  downloadPresets: DownloadPreset[];
}

// File Warning for file operations
export interface FileWarning {
  type: 'permission' | 'system' | 'location' | 'security' | 'size';
  title: string;
  message: string;
  isCritical: boolean;
}

// Cache Information
export interface CacheInfo {
  downloadCount: number;
  completedCount: number;
  pendingCount: number;
  videoCount: number;
  totalLibrarySize: number;
  orphanedFilesCount: number;
  orphanedFilesSize: number;
}

// Notification types
export type NotificationType = 'success' | 'error' | 'warning' | 'info';

export interface NotificationOptions {
  title: string;
  message?: string;
  type?: NotificationType;
  autoClose?: number;
}

// Confirm dialog options
export interface ConfirmOptions {
  title: string;
  message: string;
  confirmLabel?: string;
  cancelLabel?: string;
  confirmColor?: 'red' | 'orange' | 'blue' | 'green' | 'yted' | 'gray' | 'yellow';
  onConfirm: () => void;
  onCancel?: () => void;
}

export const DEFAULT_SETTINGS: UserSettings = {
  downloadPath: '', // Will be set to ~/Downloads/YTed on init
  maxConcurrentDownloads: 3,
  defaultQuality: 'best',
  filenameTemplate: '%(title).100s [%(id)s][%(format_id)s].%(ext)s',
  theme: 'dark',
  accentColor: '#ff0000',
  sidebarCollapsed: false,
  defaultVolume: 80,
  rememberPosition: true,
  speedLimitKbps: null,
  proxyUrl: null,
  logPath: '', // Will be set to ~/.yted/.logs on init
  logExportPath: '', // Will be set to ~/Downloads on init
  maxLogSessions: 10, // Default: keep last 10 sessions
  downloadPresets: [
    {
      id: '1',
      name: '4K (2160p)',
      format: 'bestvideo[height<=2160][vcodec^=avc1]+bestaudio/bestvideo[height<=2160]+bestaudio',
      quality: '2160p',
      extension: 'mp4',
    },
    {
      id: '2',
      name: '1440p (2K)',
      format: 'bestvideo[height<=1440][vcodec^=avc1]+bestaudio/bestvideo[height<=1440]+bestaudio',
      quality: '1440p',
      extension: 'mp4',
    },
    {
      id: '3',
      name: '1080p HD',
      format: 'bestvideo[height<=1080][vcodec^=avc1]+bestaudio/bestvideo[height<=1080]+bestaudio',
      quality: '1080p',
      extension: 'mp4',
    },
    {
      id: '4',
      name: '720p HD',
      format: 'bestvideo[height<=720][vcodec^=avc1]+bestaudio/bestvideo[height<=720]+bestaudio',
      quality: '720p',
      extension: 'mp4',
    },
    {
      id: '5',
      name: '480p',
      format: 'bestvideo[height<=480]+bestaudio/best[height<=480]',
      quality: '480p',
      extension: 'mp4',
    },
    {
      id: '6',
      name: 'Best Available',
      format: 'bestvideo+bestaudio/best',
      quality: 'best',
      extension: 'mp4',
    },
    { id: '7', name: 'Audio Only (MP3)', format: 'bestaudio', quality: 'audio', extension: 'mp3' },
    {
      id: '8',
      name: 'Audio Only (M4A)',
      format: 'bestaudio[ext=m4a]',
      quality: 'audio',
      extension: 'm4a',
    },
  ],
};
