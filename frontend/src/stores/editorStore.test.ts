import { beforeEach, describe, expect, it, vi } from 'vitest';

import { CancelEditJob, GetEditOptions, SubmitEditJob } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';

import { useEditorStore } from './editorStore';

// Mock the Wails API
vi.mock('../../wailsjs/go/app/App', () => ({
  GetEditOptions: vi.fn(),
  SubmitEditJob: vi.fn(),
  CancelEditJob: vi.fn(),
}));

// Mock the wails runtime
vi.mock('../../wailsjs/runtime', () => ({
  EventsOn: vi.fn(() => vi.fn()),
}));

const mockedGetEditOptions = GetEditOptions as unknown as ReturnType<typeof vi.fn>;
const mockedSubmitEditJob = SubmitEditJob as unknown as ReturnType<typeof vi.fn>;
const mockedCancelEditJob = CancelEditJob as unknown as ReturnType<typeof vi.fn>;
const mockedEventsOn = vi.mocked(EventsOn);

const mockOptions = {
  formats: [
    { id: 'mp4', name: 'MP4', extension: 'mp4', description: 'MP4 video', codecs: ['h264'] },
  ],
  codecs: [{ id: 'h264', name: 'H.264', description: '', quality: 'good', speed: 'fast' }],
  crop_presets: [],
  effect_ranges: [
    { id: 'brightness', min: 0, max: 2, default: 1, step: 0.1, description: 'Brightness' },
  ],
  rotations: [{ value: 0, label: 'None', description: '' }],
  watermark_positions: { bottom_right: 'Bottom right' },
};

const addTestJob = () => {
  useEditorStore.setState({
    jobs: {
      'job-1': {
        jobId: 'job-1',
        videoId: 'video-1',
        operation: 'crop',
        progress: 0,
        status: 'pending',
      },
    },
  });
};

describe('editorStore', () => {
  let eventHandlers: Record<string, (data: unknown) => void>;
  let unsubscribers: Array<ReturnType<typeof vi.fn>>;

  beforeEach(() => {
    useEditorStore.setState({ options: null, ffmpegError: null, jobs: {} });
    vi.clearAllMocks();
    eventHandlers = {};
    unsubscribers = [];
    mockedEventsOn.mockImplementation((eventName, callback) => {
      eventHandlers[eventName] = callback;
      const unsubscribe = vi.fn();
      unsubscribers.push(unsubscribe);
      return unsubscribe;
    });
  });

  it('should have correct initial state', () => {
    const state = useEditorStore.getState();

    expect(state.options).toBeNull();
    expect(state.ffmpegError).toBeNull();
    expect(state.jobs).toEqual({});
  });

  it('should load options successfully', async () => {
    mockedGetEditOptions.mockResolvedValue(mockOptions);

    await useEditorStore.getState().loadOptions();

    const state = useEditorStore.getState();
    expect(state.options).toEqual(mockOptions);
    expect(state.ffmpegError).toBeNull();
  });

  it('should set ffmpegError when loading options fails', async () => {
    mockedGetEditOptions.mockRejectedValue(new Error('ffmpeg not found'));

    await useEditorStore.getState().loadOptions();

    const state = useEditorStore.getState();
    expect(state.ffmpegError).toBe('ffmpeg not found');
    expect(state.options).toBeNull();
  });

  it('should add a job on successful submit', async () => {
    mockedSubmitEditJob.mockResolvedValue('job-1');

    const settings = { crop_start: 0, crop_end: 10 };
    const jobId = await useEditorStore.getState().submitJob('video-1', 'crop', settings);

    expect(jobId).toBe('job-1');
    expect(mockedSubmitEditJob).toHaveBeenCalledWith('video-1', 'crop', settings);

    const job = useEditorStore.getState().jobs['job-1'];
    expect(job).toBeDefined();
    expect(job.videoId).toBe('video-1');
    expect(job.operation).toBe('crop');
    expect(job.progress).toBe(0);
    expect(job.status).toBe('pending');
  });

  it('should throw and not add a job when submit fails', async () => {
    mockedSubmitEditJob.mockRejectedValue(new Error('submit failed'));

    await expect(useEditorStore.getState().submitJob('video-1', 'crop', {})).rejects.toThrow(
      'submit failed'
    );
    expect(useEditorStore.getState().jobs).toEqual({});
  });

  it('should cancel a job', async () => {
    mockedCancelEditJob.mockResolvedValue(undefined);

    await useEditorStore.getState().cancelJob('job-1');

    expect(mockedCancelEditJob).toHaveBeenCalledWith('job-1');
  });

  it('should subscribe to editor events and return a combined unsubscribe', () => {
    const cleanup = useEditorStore.getState().initEvents();

    expect(mockedEventsOn).toHaveBeenCalledTimes(3);
    expect(mockedEventsOn).toHaveBeenCalledWith('editor:progress', expect.any(Function));
    expect(mockedEventsOn).toHaveBeenCalledWith('editor:completed', expect.any(Function));
    expect(mockedEventsOn).toHaveBeenCalledWith('editor:error', expect.any(Function));

    cleanup();
    expect(unsubscribers).toHaveLength(3);
    unsubscribers.forEach(unsubscribe => expect(unsubscribe).toHaveBeenCalled());
  });

  it('should update job progress on editor:progress events', () => {
    addTestJob();
    useEditorStore.getState().initEvents();

    eventHandlers['editor:progress']({ jobId: 'job-1', progress: 0.5 });

    const job = useEditorStore.getState().jobs['job-1'];
    expect(job.progress).toBe(0.5);
    expect(job.status).toBe('processing');
  });

  it('should mark job as completed on editor:completed events', () => {
    addTestJob();
    useEditorStore.getState().initEvents();

    eventHandlers['editor:completed']({ jobId: 'job-1', outputVideoId: 'video-2' });

    const job = useEditorStore.getState().jobs['job-1'];
    expect(job.status).toBe('completed');
    expect(job.progress).toBe(1);
  });

  it('should mark job as error on editor:error events', () => {
    addTestJob();
    useEditorStore.getState().initEvents();

    eventHandlers['editor:error']({ jobId: 'job-1', error: 'encoder crashed' });

    const job = useEditorStore.getState().jobs['job-1'];
    expect(job.status).toBe('error');
    expect(job.error).toBe('encoder crashed');
  });

  it('should ignore events for unknown jobs', () => {
    useEditorStore.getState().initEvents();

    eventHandlers['editor:progress']({ jobId: 'unknown', progress: 0.5 });

    expect(useEditorStore.getState().jobs).toEqual({});
  });
});
