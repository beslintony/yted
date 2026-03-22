import { create } from 'zustand';

import {
  CheckFFmpegWithGuidance,
  GetEditJobStatus,
  GetFFmpegInstallGuide,
  GetPreviewImage,
  GetVideoMetadata,
  ListEditJobs,
  PreviewEdit,
  SubmitEditJob,
} from '../../wailsjs/go/app/App';
import { app } from '../../wailsjs/go/models';
import {
  DEFAULT_EDIT_SETTINGS,
  EditJob,
  EditOperation,
  EditSettings,
  EditorPanelTab,
  FFmpegCheckResult,
  InstallGuide,
  VideoMetadata,
} from '../types/editor';

interface EditorState {
  // FFmpeg status
  ffmpegStatus: FFmpegCheckResult | null;
  installGuide: InstallGuide | null;
  isCheckingFFmpeg: boolean;

  // Selected video
  selectedVideoId: string | null;
  videoMetadata: VideoMetadata | null;
  isLoadingMetadata: boolean;

  // Editor settings
  currentOperation: EditOperation | null;
  settings: EditSettings;

  // Preview
  previewFrame: string | null;
  isGeneratingPreview: boolean;

  // Jobs
  jobs: EditJob[];
  isLoadingJobs: boolean;

  // Submit
  isSubmitting: boolean;
  lastJobId: string | null;

  // UI state
  activeTab: EditorPanelTab;
  showReplaceConfirm: boolean;

  // Polling
  pollInterval: ReturnType<typeof setInterval> | null;

  // Actions
  checkFFmpeg: () => Promise<void>;
  getInstallGuide: () => Promise<InstallGuide>;
  selectVideo: (videoId: string | null) => void;
  loadVideoMetadata: (videoId: string) => Promise<void>;
  setOperation: (operation: EditOperation | null) => void;
  updateSettings: (settings: Partial<EditSettings>) => void;
  resetSettings: () => void;
  generatePreview: () => Promise<void>;
  loadJobs: (videoId: string) => Promise<void>;
  submitJob: (replaceOriginal?: boolean) => Promise<string | null>;
  cancelJob: (jobId: string) => Promise<void>;
  setActiveTab: (tab: EditorPanelTab) => void;
  setShowReplaceConfirm: (show: boolean) => void;
  reset: () => void;
  startJobPolling: (videoId: string) => void;
  stopJobPolling: () => void;
}

export const useEditorStore = create<EditorState>((set, get) => ({
  // Initial state
  ffmpegStatus: null,
  installGuide: null,
  isCheckingFFmpeg: false,

  selectedVideoId: null,
  videoMetadata: null,
  isLoadingMetadata: false,

  currentOperation: null,
  settings: { ...DEFAULT_EDIT_SETTINGS },

  previewFrame: null,
  isGeneratingPreview: false,

  jobs: [],
  isLoadingJobs: false,

  isSubmitting: false,
  lastJobId: null,

  activeTab: 'preview',
  showReplaceConfirm: false,

  pollInterval: null,
  checkFFmpeg: async () => {
    set({ isCheckingFFmpeg: true });
    try {
      const status = await CheckFFmpegWithGuidance();
      set({ ffmpegStatus: status as FFmpegCheckResult });
    } catch (err) {
      console.error('Failed to check FFmpeg:', err);
    } finally {
      set({ isCheckingFFmpeg: false });
    }
  },

  getInstallGuide: async () => {
    try {
      const guide = await GetFFmpegInstallGuide();
      set({ installGuide: guide as InstallGuide });
      return guide as InstallGuide;
    } catch (err) {
      console.error('Failed to get install guide:', err);
      throw err;
    }
  },

  selectVideo: (videoId: string | null) => {
    set({
      selectedVideoId: videoId,
      videoMetadata: null,
      previewFrame: null,
      settings: { ...DEFAULT_EDIT_SETTINGS },
      ...(videoId ? { activeTab: 'preview' as const } : {}),
    });
    if (videoId) {
      get().loadVideoMetadata(videoId);
      get().loadJobs(videoId);
    }
  },

  loadVideoMetadata: async (videoId: string) => {
    console.log('[EditorStore] Loading video metadata for:', videoId);
    set({ isLoadingMetadata: true });
    try {
      const metadata = await GetVideoMetadata(videoId);
      console.log('[EditorStore] Metadata received:', metadata);
      if (metadata) {
        // Convert backend type to frontend type
        const videoMeta: VideoMetadata = {
          duration: (metadata as app.VideoMetadataResult).duration || 0,
          width: (metadata as app.VideoMetadataResult).width || 0,
          height: (metadata as app.VideoMetadataResult).height || 0,
          fps: (metadata as app.VideoMetadataResult).fps || 0,
          bitrate: (metadata as app.VideoMetadataResult).bitrate || 0,
          codec: (metadata as app.VideoMetadataResult).codec || 'unknown',
          audioCodec: (metadata as app.VideoMetadataResult).audio_codec,
          hasAudio: (metadata as app.VideoMetadataResult).has_audio || false,
        };
        set({ videoMetadata: videoMeta });
        // Set default crop end to video duration
        if (videoMeta.duration > 0) {
          set(state => ({
            settings: {
              ...state.settings,
              cropEnd: videoMeta.duration,
            },
          }));
        }
      } else {
        console.warn('[EditorStore] No metadata returned for video:', videoId);
      }
    } catch (err) {
      console.error('[EditorStore] Failed to load video metadata:', err);
    } finally {
      set({ isLoadingMetadata: false });
    }
  },

  setOperation: (operation: EditOperation | null) => {
    set({ currentOperation: operation });
    // Reset settings to defaults when switching operations
    if (operation) {
      set({ settings: { ...DEFAULT_EDIT_SETTINGS } });
    }
  },

  updateSettings: (newSettings: Partial<EditSettings>) => {
    set(state => ({
      settings: { ...state.settings, ...newSettings },
    }));
  },

  resetSettings: () => {
    set({ settings: { ...DEFAULT_EDIT_SETTINGS } });
  },

  generatePreview: async () => {
    const { selectedVideoId, currentOperation, settings } = get();
    if (!selectedVideoId || !currentOperation) return;

    set({ isGeneratingPreview: true });
    try {
      const previewPath = await PreviewEdit(
        selectedVideoId,
        currentOperation,
        settings as unknown as app.EditSettingsInput
      );
      if (previewPath) {
        // Load the preview image through the backend to get a data URL
        const dataUrl = await GetPreviewImage(previewPath);
        if (dataUrl) {
          set({ previewFrame: dataUrl });
        }
      }
    } catch (err) {
      console.error('Failed to generate preview:', err);
    } finally {
      set({ isGeneratingPreview: false });
    }
  },

  loadJobs: async (videoId: string) => {
    set({ isLoadingJobs: true });
    try {
      const jobs = await ListEditJobs(videoId);
      // Convert backend job results to frontend job types
      const convertedJobs: EditJob[] = ((jobs || []) as app.EditJobResult[]).map(j => ({
        id: j.id,
        sourceVideoId: j.source_video_id,
        outputVideoId: j.output_video_id || undefined,
        status: j.status as EditJob['status'],
        operation: j.operation as EditOperation,
        settings: {}, // Settings are stored as JSON string in backend
        progress: j.progress,
        errorMessage: j.error_message || undefined,
        createdAt: j.created_at,
        completedAt: j.completed_at || undefined,
      }));
      set({ jobs: convertedJobs });
    } catch (err) {
      console.error('Failed to load edit jobs:', err);
    } finally {
      set({ isLoadingJobs: false });
    }
  },

  submitJob: async (replaceOriginal = false) => {
    const { selectedVideoId, currentOperation, settings, isSubmitting } = get();
    if (!selectedVideoId || !currentOperation) {
      return null;
    }
    
    // Prevent duplicate submissions
    if (isSubmitting) {
      return null;
    }

    set({ isSubmitting: true });
    try {
      const jobSettings = {
        ...settings,
        replaceOriginal,
      };

      const jobId = await SubmitEditJob(
        selectedVideoId,
        currentOperation,
        jobSettings as unknown as app.EditSettingsInput
      );

      set({ lastJobId: jobId });

      // Reload jobs list
      await get().loadJobs(selectedVideoId);

      // Start polling for job status
      get().startJobPolling(selectedVideoId);

      return jobId;
    } catch (err) {
      console.error('Failed to submit edit job:', err);
      return null;
    } finally {
      set({ isSubmitting: false });
    }
  },

  // Poll for job status updates (since we don't have websocket events from backend)
  startJobPolling: (videoId: string) => {
    // Clear any existing interval
    const currentInterval = get().pollInterval;
    if (currentInterval) {
      clearInterval(currentInterval);
    }

    // Poll every 2 seconds for updates
    const interval = setInterval(async () => {
      const { jobs } = get();
      
      // Check if there are any active jobs
      const hasActiveJobs = jobs.some(
        job => job.status === 'pending' || job.status === 'processing'
      );
      
      if (hasActiveJobs) {
        await get().loadJobs(videoId);
      } else {
        // No active jobs, stop polling
        const current = get().pollInterval;
        if (current) {
          clearInterval(current);
          set({ pollInterval: null });
        }
      }
    }, 2000);

    set({ pollInterval: interval });
  },
  stopJobPolling: () => {
    const interval = get().pollInterval;
    if (interval) {
      clearInterval(interval);
      set({ pollInterval: null });
    }
  },

  cancelJob: async (jobId: string) => {
    try {
      // Note: CancelEditJob needs to be added to backend
      console.log('Cancel job:', jobId);
    } catch (err) {
      console.error('Failed to cancel job:', err);
    }
  },

  setActiveTab: (tab: EditorPanelTab) => {
    set({ activeTab: tab });
  },

  setShowReplaceConfirm: (show: boolean) => {
    set({ showReplaceConfirm: show });
  },

  reset: () => {
    // Stop any active polling
    const interval = get().pollInterval;
    if (interval) {
      clearInterval(interval);
    }
    set({
      selectedVideoId: null,
      videoMetadata: null,
      currentOperation: null,
      settings: { ...DEFAULT_EDIT_SETTINGS },
      previewFrame: null,
      jobs: [],
      lastJobId: null,
      showReplaceConfirm: false,
      pollInterval: null,
    });
  },
}));
