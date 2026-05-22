import { beforeEach, describe, expect, it, vi } from 'vitest';

import { useVersionStore } from './versionStore';

// Mock the Wails API
vi.mock('../../wailsjs/go/app/App', () => ({
  GetVersion: vi.fn(),
}));

describe('versionStore', () => {
  beforeEach(() => {
    useVersionStore.setState({
      version: 'dev',
      isLoading: false,
    });
    vi.clearAllMocks();
  });

  it('should have correct initial state', () => {
    const state = useVersionStore.getState();

    expect(state.version).toBe('dev');
    expect(state.isLoading).toBe(false);
  });

  it('should fetch version successfully', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    mockedGetVersion.mockResolvedValue('1.4.1');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    expect(useVersionStore.getState().version).toBe('1.4.1');
    expect(useVersionStore.getState().isLoading).toBe(false);
  });

  it('should handle fetch version error gracefully', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    mockedGetVersion.mockRejectedValue(new Error('Failed to get version'));

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    // On error, version should remain unchanged and loading should be false
    expect(useVersionStore.getState().version).toBe('dev');
    expect(useVersionStore.getState().isLoading).toBe(false);
  });

  it('should set loading state during fetch', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    // Create a promise that we can resolve manually
    let resolvePromise: (value: string) => void;
    const promise = new Promise<string>(resolve => {
      resolvePromise = resolve;
    });
    mockedGetVersion.mockReturnValue(promise);

    const { fetchVersion } = useVersionStore.getState();
    const fetchPromise = fetchVersion();

    // Loading should be true while fetching
    expect(useVersionStore.getState().isLoading).toBe(true);

    resolvePromise!('1.4.1');
    await fetchPromise;

    expect(useVersionStore.getState().isLoading).toBe(false);
    expect(useVersionStore.getState().version).toBe('1.4.1');
  });

  it('should handle empty version string', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    mockedGetVersion.mockResolvedValue('');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    expect(useVersionStore.getState().version).toBe('');
  });

  it('should handle version with build metadata', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    mockedGetVersion.mockResolvedValue('1.4.1+build.123');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    expect(useVersionStore.getState().version).toBe('1.4.1+build.123');
  });

  it('should handle non-Error rejection', async () => {
    const { GetVersion } = await import('../../wailsjs/go/app/App');
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;

    mockedGetVersion.mockRejectedValue('string error');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    // Should handle gracefully without crashing
    expect(useVersionStore.getState().version).toBe('dev');
    expect(useVersionStore.getState().isLoading).toBe(false);
  });
});
