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
export type QualityOption = 'best' | '1080p' | '720p' | '480p' | '360p' | 'audio';

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
  logPath: string;          // Internal log storage path (.logs folder)
  logExportPath: string;    // Export destination
  maxLogSessions: number;   // Number of sessions to keep (default: 10)
  
  // Presets
  downloadPresets: DownloadPreset[];
}

export const DEFAULT_SETTINGS: UserSettings = {
  downloadPath: '', // Will be set to ~/Downloads/YTed on init
  maxConcurrentDownloads: 3,
  defaultQuality: 'best',
  filenameTemplate: '%(title)s.%(ext)s',
  theme: 'dark',
  accentColor: '#ff0000',
  sidebarCollapsed: false,
  defaultVolume: 80,
  rememberPosition: true,
  speedLimitKbps: null,
  proxyUrl: null,
  logPath: '',          // Will be set to ~/.yted/.logs on init
  logExportPath: '',    // Will be set to ~/Downloads on init
  maxLogSessions: 10,   // Default: keep last 10 sessions
  downloadPresets: [
    { id: '1', name: 'Best Quality', format: 'best', quality: 'best', extension: 'mp4' },
    { id: '2', name: '1080p', format: 'bestvideo[height<=1080]+bestaudio', quality: '1080p', extension: 'mp4' },
    { id: '3', name: '720p', format: 'bestvideo[height<=720]+bestaudio', quality: '720p', extension: 'mp4' },
    { id: '4', name: 'Audio Only', format: 'bestaudio', quality: 'audio', extension: 'mp3' },
  ],
};
