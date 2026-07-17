import { create } from 'zustand';

import { CancelEditJob, GetEditOptions, SubmitEditJob } from '../../wailsjs/go/app/App';
import { app, db } from '../../wailsjs/go/models';
import { EventsOn } from '../../wailsjs/runtime';

export type EditJobStatus = 'pending' | 'processing' | 'completed' | 'error';

export interface EditJobState {
  jobId: string;
  videoId: string;
  operation: string;
  progress: number;
  status: EditJobStatus;
  error?: string;
}

interface EditorState {
  options: app.EditOptions | null;
  ffmpegError: string | null;
  jobs: Record<string, EditJobState>;

  // Actions
  loadOptions: () => Promise<void>;
  submitJob: (videoId: string, operation: string, settings: db.EditSettings) => Promise<string>;
  cancelJob: (jobId: string) => Promise<void>;
  initEvents: () => () => void;
}

export const useEditorStore = create<EditorState>((set, get) => ({
  options: null,
  ffmpegError: null,
  jobs: {},

  loadOptions: async () => {
    try {
      const options = await GetEditOptions();
      set({ options, ffmpegError: null });
    } catch (err) {
      set({
        ffmpegError: err instanceof Error ? err.message : 'Failed to load editor options',
      });
    }
  },

  submitJob: async (videoId, operation, settings) => {
    const jobId = await SubmitEditJob(videoId, operation, settings);
    set(state => ({
      jobs: {
        ...state.jobs,
        [jobId]: { jobId, videoId, operation, progress: 0, status: 'pending' },
      },
    }));
    return jobId;
  },

  cancelJob: async jobId => {
    try {
      await CancelEditJob(jobId);
    } catch (err) {
      console.error('Failed to cancel edit job:', err);
    }
  },

  initEvents: () => {
    const updateJob = (jobId: string, updates: Partial<EditJobState>) => {
      const job = get().jobs[jobId];
      if (!job) return;
      set(state => ({
        jobs: {
          ...state.jobs,
          [jobId]: { ...job, ...updates },
        },
      }));
    };

    const cancelProgress = EventsOn(
      'editor:progress',
      (data: { jobId?: string; progress?: number }) => {
        if (data?.jobId && typeof data.progress === 'number') {
          updateJob(data.jobId, { progress: data.progress, status: 'processing' });
        }
      }
    );

    const cancelCompleted = EventsOn(
      'editor:completed',
      (data: { jobId?: string; outputVideoId?: string }) => {
        if (data?.jobId) {
          updateJob(data.jobId, { progress: 1, status: 'completed' });
        }
      }
    );

    const cancelError = EventsOn('editor:error', (data: { jobId?: string; error?: string }) => {
      if (data?.jobId) {
        updateJob(data.jobId, { status: 'error', error: data.error || 'Unknown error' });
      }
    });

    return () => {
      cancelProgress();
      cancelCompleted();
      cancelError();
    };
  },
}));
