import { create } from 'zustand';

import {
  CheckFFmpegWithGuidance,
  GetEditJobStatus,
  GetFFmpegInstallGuide,
  GetVideoMetadata,
  ListEditJobs,
  PreviewEdit,
  SubmitEditJob,
} from '../../wailsjs/go/app/App';
import { app, editor } from '../../wailsjs/go/models';
import {
  DEFAULT_EDIT_SETTINGS,
  EditJob,
  EditOperation,
  EditSettings,
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
  activeTab: 'tools' | 'presets' | 'history';
  showReplaceConfirm: boolean;

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
  setActiveTab: (tab: 'tools' | 'presets' | 'history') => void;
  setShowReplaceConfirm: (show: boolean) => void;
  reset: () => void;
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

  activeTab: 'tools',
  showReplaceConfirm: false,

  // Actions
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
    });
    if (videoId) {
      get().loadVideoMetadata(videoId);
      get().loadJobs(videoId);
    }
  },

  loadVideoMetadata: async (videoId: string) => {
    set({ isLoadingMetadata: true });
    try {
      const metadata = await GetVideoMetadata(videoId);
      if (metadata) {
        set({ videoMetadata: metadata as VideoMetadata });
        // Set default crop end to video duration
        if ((metadata as VideoMetadata).duration > 0) {
          set(state => ({
            settings: {
              ...state.settings,
              cropEnd: (metadata as VideoMetadata).duration,
            },
          }));
        }
      }
    } catch (err) {
      console.error('Failed to load video metadata:', err);
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
        settings as editor.EditSettingsInput
      );
      if (previewPath) {
        // Convert path to file URL for display
        set({ previewFrame: `file://${previewPath}` });
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
      set({ jobs: (jobs || []) as EditJob[] });
    } catch (err) {
      console.error('Failed to load edit jobs:', err);
    } finally {
      set({ isLoadingJobs: false });
    }
  },

  submitJob: async (replaceOriginal = false) => {
    const { selectedVideoId, currentOperation, settings } = get();
    if (!selectedVideoId || !currentOperation) {
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
        jobSettings as editor.EditSettingsInput
      );

      set({ lastJobId: jobId });

      // Reload jobs list
      get().loadJobs(selectedVideoId);

      return jobId;
    } catch (err) {
      console.error('Failed to submit edit job:', err);
      return null;
    } finally {
      set({ isSubmitting: false });
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

  setActiveTab: (tab: 'tools' | 'presets' | 'history') => {
    set({ activeTab: tab });
  },

  setShowReplaceConfirm: (show: boolean) => {
    set({ showReplaceConfirm: show });
  },

  reset: () => {
    set({
      selectedVideoId: null,
      videoMetadata: null,
      currentOperation: null,
      settings: { ...DEFAULT_EDIT_SETTINGS },
      previewFrame: null,
      jobs: [],
      lastJobId: null,
      showReplaceConfirm: false,
    });
  },
}));
