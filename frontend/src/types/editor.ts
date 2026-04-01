// Video Editor Types

export type EditOperation = 'crop' | 'watermark' | 'convert' | 'effects' | 'combine';

export interface FFmpegCheckResult {
  installed: boolean;
  version: string;
  path: string;
  canAutoInstall: boolean;
  installMethod: 'package_manager' | 'download' | 'manual';
  installCommand: string;
  installGuide: string;
  downloadURL: string;
  requiresAdmin: boolean;
}

export interface InstallGuide {
  title: string;
  description: string;
  steps: string[];
  command: string;
  commandDescription: string;
  alternativeURL: string;
  tips: string[];
}

export interface EditSettings {
  // Crop settings
  cropStart?: number;
  cropEnd?: number;
  cropX?: number;
  cropY?: number;
  cropWidth?: number;
  cropHeight?: number;

  // Watermark settings
  watermarkType?: 'text' | 'image';
  watermarkText?: string;
  watermarkImage?: string;
  watermarkPosition?: WatermarkPosition;
  watermarkOpacity?: number;
  watermarkSize?: number;

  // Convert settings
  outputFormat?: OutputFormat;
  outputCodec?: OutputCodec;
  outputQuality?: number;
  outputResolution?: OutputResolution;

  // Effects settings
  brightness?: number;
  contrast?: number;
  saturation?: number;
  rotation?: Rotation;
  speed?: number;
  volume?: number;
  removeAudio?: boolean;

  // Output options
  outputFilename?: string;
  replaceOriginal?: boolean;
}

export type WatermarkPosition =
  | 'top-left'
  | 'top-center'
  | 'top-right'
  | 'center-left'
  | 'center'
  | 'center-right'
  | 'bottom-left'
  | 'bottom-center'
  | 'bottom-right';

export type OutputFormat = 'mp4' | 'mkv' | 'webm' | 'mov' | 'avi' | 'gif';

export type OutputCodec = 'h264' | 'h265' | 'vp9' | 'av1';

export type OutputResolution = 'original' | '2160p' | '1440p' | '1080p' | '720p' | '480p' | '360p';

export type Rotation = 0 | 90 | 180 | 270;

export interface EditJob {
  id: string;
  sourceVideoId: string;
  outputVideoId?: string;
  status: 'pending' | 'processing' | 'completed' | 'error';
  operation: EditOperation;
  settings: EditSettings;
  progress: number;
  errorMessage?: string;
  createdAt: number;
  completedAt?: number;
}

export interface VideoMetadata {
  duration: number;
  width: number;
  height: number;
  fps: number;
  bitrate: number;
  codec: string;
  audioCodec?: string;
  hasAudio: boolean;
}

export interface EditPreset {
  id: string;
  name: string;
  description: string;
  icon: string;
  operation: EditOperation;
  defaultSettings: Partial<EditSettings>;
}

// Crop preset with aspect ratio
export interface CropPreset {
  id: string;
  name: string;
  ratio: string;
  width: number;
  height: number;
}

// Format info for conversion
export interface FormatInfo {
  id: OutputFormat;
  name: string;
  extension: string;
  description: string;
  codecs: OutputCodec[];
}

// Codec info
export interface CodecInfo {
  id: OutputCodec;
  name: string;
  description: string;
  quality: 'good' | 'better' | 'best';
  speed: 'fast' | 'medium' | 'slow';
}

// Effect parameter range
export interface EffectRange {
  min: number;
  max: number;
  default: number;
  step: number;
  description: string;
}

// Editor UI state
export interface EditorUIState {
  selectedVideoId: string | null;
  currentOperation: EditOperation | null;
  settings: EditSettings;
  previewFrame: string | null;
  isGeneratingPreview: boolean;
  activeTab: 'tools' | 'presets' | 'history';
  showReplaceConfirm: boolean;
}

// Output resolutions for conversion
export const OUTPUT_RESOLUTIONS = [
  { id: 'original', name: 'Original Resolution' },
  { id: '2160p', name: '4K (2160p)' },
  { id: '1440p', name: '2K (1440p)' },
  { id: '1080p', name: '1080p Full HD' },
  { id: '720p', name: '720p HD' },
  { id: '480p', name: '480p SD' },
  { id: '360p', name: '360p' },
] as const;

// Default settings
export const DEFAULT_EDIT_SETTINGS: EditSettings = {
  cropStart: 0,
  cropEnd: 0,
  watermarkType: 'text',
  watermarkText: 'YTed',
  watermarkPosition: 'bottom-right',
  watermarkOpacity: 0.7,
  watermarkSize: 24,
  outputFormat: 'mp4',
  outputCodec: 'h264',
  outputQuality: 23,
  outputResolution: 'original',
  brightness: 0,
  contrast: 0,
  saturation: 1,
  rotation: 0,
  speed: 1,
  volume: 1,
  removeAudio: false,
  replaceOriginal: false,
};

// Crop presets
export const CROP_PRESETS: CropPreset[] = [
  { id: 'free', name: 'Freeform', ratio: 'Free', width: 0, height: 0 },
  { id: '16:9', name: 'Widescreen', ratio: '16:9', width: 16, height: 9 },
  { id: '4:3', name: 'Standard', ratio: '4:3', width: 4, height: 3 },
  { id: '1:1', name: 'Square', ratio: '1:1', width: 1, height: 1 },
  { id: '9:16', name: 'Vertical', ratio: '9:16', width: 9, height: 16 },
  { id: '21:9', name: 'Cinema', ratio: '21:9', width: 21, height: 9 },
];

// Watermark positions
export const WATERMARK_POSITIONS: { value: WatermarkPosition; label: string }[] = [
  { value: 'top-left', label: 'Top Left' },
  { value: 'top-center', label: 'Top Center' },
  { value: 'top-right', label: 'Top Right' },
  { value: 'center-left', label: 'Center Left' },
  { value: 'center', label: 'Center' },
  { value: 'center-right', label: 'Center Right' },
  { value: 'bottom-left', label: 'Bottom Left' },
  { value: 'bottom-center', label: 'Bottom Center' },
  { value: 'bottom-right', label: 'Bottom Right' },
];

// Output formats
export const OUTPUT_FORMATS: FormatInfo[] = [
  {
    id: 'mp4',
    name: 'MP4',
    extension: 'mp4',
    description: 'Best compatibility',
    codecs: ['h264', 'h265'],
  },
  {
    id: 'mkv',
    name: 'MKV',
    extension: 'mkv',
    description: 'Universal container',
    codecs: ['h264', 'h265', 'vp9', 'av1'],
  },
  { id: 'webm', name: 'WebM', extension: 'webm', description: 'Web optimized', codecs: ['vp9'] },
  {
    id: 'mov',
    name: 'QuickTime',
    extension: 'mov',
    description: 'Apple format',
    codecs: ['h264', 'h265'],
  },
  { id: 'avi', name: 'AVI', extension: 'avi', description: 'Legacy format', codecs: ['h264'] },
  { id: 'gif', name: 'GIF', extension: 'gif', description: 'Animated image', codecs: [] },
];

// Output codecs
export const OUTPUT_CODECS: CodecInfo[] = [
  {
    id: 'h264',
    name: 'H.264 (AVC)',
    description: 'Best compatibility',
    quality: 'good',
    speed: 'fast',
  },
  {
    id: 'h265',
    name: 'H.265 (HEVC)',
    description: 'Better compression',
    quality: 'better',
    speed: 'medium',
  },
  { id: 'vp9', name: 'VP9', description: 'Web streaming', quality: 'better', speed: 'slow' },
  { id: 'av1', name: 'AV1', description: 'Next generation', quality: 'best', speed: 'slow' },
];

// Effect ranges
export const EFFECT_RANGES: Record<string, EffectRange> = {
  brightness: { min: -1, max: 1, default: 0, step: 0.1, description: 'Brightness adjustment' },
  contrast: { min: -1, max: 1, default: 0, step: 0.1, description: 'Contrast adjustment' },
  saturation: { min: 0, max: 2, default: 1, step: 0.1, description: 'Saturation adjustment' },
  speed: { min: 0.5, max: 2, default: 1, step: 0.1, description: 'Playback speed' },
  volume: { min: 0, max: 2, default: 1, step: 0.1, description: 'Volume level' },
  quality: { min: 18, max: 35, default: 23, step: 1, description: 'Quality (lower is better)' },
  opacity: { min: 0.1, max: 1, default: 0.7, step: 0.1, description: 'Watermark opacity' },
};

// Rotation options
export const ROTATION_OPTIONS: { value: Rotation; label: string; icon: string }[] = [
  { value: 0, label: 'No Rotation', icon: 'rotate' },
  { value: 90, label: '90° Clockwise', icon: 'rotate-clockwise' },
  { value: 180, label: '180°', icon: 'rotate-180' },
  { value: 270, label: '90° Counter-Clockwise', icon: 'rotate-counter-clockwise' },
];
