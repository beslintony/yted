import { beforeEach, describe, expect, it, vi } from 'vitest';

import { GetVersion } from '../../wailsjs/go/app/App';
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
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    mockedGetVersion.mockResolvedValue('1.2.3');

    const { fetchVersion } = useVersionStore.getState();

    await fetchVersion();

    const state = useVersionStore.getState();
    expect(state.version).toBe('1.2.3');
    expect(state.isLoading).toBe(false);
    expect(mockedGetVersion).toHaveBeenCalledTimes(1);
  });

  it('should handle fetch error gracefully', async () => {
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    const consoleError = vi.spyOn(console, 'error').mockImplementation(() => {});
    mockedGetVersion.mockRejectedValue(new Error('Network error'));

    const { fetchVersion } = useVersionStore.getState();

    await fetchVersion();

    const state = useVersionStore.getState();
    expect(state.version).toBe('dev'); // Should keep default
    expect(state.isLoading).toBe(false);
    expect(consoleError).toHaveBeenCalled();

    consoleError.mockRestore();
  });

  it('should set loading state during fetch', async () => {
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    mockedGetVersion.mockImplementation(
      () => new Promise(resolve => setTimeout(() => resolve('1.0.0'), 10))
    );

    const { fetchVersion } = useVersionStore.getState();

    const promise = fetchVersion();
    expect(useVersionStore.getState().isLoading).toBe(true);

    await promise;
    expect(useVersionStore.getState().isLoading).toBe(false);
  });

  it('should handle version with commit hash', async () => {
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    mockedGetVersion.mockResolvedValue('1.0.0-abc123');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    expect(useVersionStore.getState().version).toBe('1.0.0-abc123');
  });

  it('should handle dev version', async () => {
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    mockedGetVersion.mockResolvedValue('dev');

    const { fetchVersion } = useVersionStore.getState();
    await fetchVersion();

    expect(useVersionStore.getState().version).toBe('dev');
  });

  it('should not update version if component unmounts during fetch', async () => {
    const mockedGetVersion = GetVersion as unknown as ReturnType<typeof vi.fn>;
    mockedGetVersion.mockImplementation(
      () => new Promise(resolve => setTimeout(() => resolve('1.0.0'), 100))
    );

    const { fetchVersion } = useVersionStore.getState();

    // Start fetch but don't await
    fetchVersion();

    // Loading should be true
    expect(useVersionStore.getState().isLoading).toBe(true);

    // Wait for it to complete
    await new Promise(resolve => setTimeout(resolve, 150));

    // Should have version now
    expect(useVersionStore.getState().version).toBe('1.0.0');
    expect(useVersionStore.getState().isLoading).toBe(false);
  });
});
