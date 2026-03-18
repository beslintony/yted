import { beforeEach, describe, expect, it, vi } from 'vitest';

import { LogLevel, useLogStore } from './logStore';

// Mock the Wails API
vi.mock('../../wailsjs/go/app/App', () => ({
  ExportLogs: vi.fn(),
}));

describe('logStore', () => {
  beforeEach(() => {
    useLogStore.setState({
      entries: [],
      isLoading: false,
      error: null,
    });
    vi.clearAllMocks();
  });

  it('should have correct initial state', () => {
    const state = useLogStore.getState();

    expect(state.entries).toEqual([]);
    expect(state.isLoading).toBe(false);
    expect(state.error).toBeNull();
  });

  it('should add entry', () => {
    const { addEntry } = useLogStore.getState();

    const newEntry = {
      timestamp: '2024-01-01 12:00:00',
      level: 'INFO' as LogLevel,
      component: 'Test',
      message: 'Test message',
    };

    addEntry(newEntry);

    const entries = useLogStore.getState().entries;
    expect(entries).toHaveLength(1);
    expect(entries[0]).toEqual(newEntry);
  });

  it('should limit entries to 1000', () => {
    const { addEntry } = useLogStore.getState();

    // Add 1005 entries
    for (let i = 0; i < 1005; i++) {
      addEntry({
        timestamp: '2024-01-01 12:00:00',
        level: 'INFO',
        component: 'Test',
        message: `Message ${i}`,
      });
    }

    const entries = useLogStore.getState().entries;
    expect(entries.length).toBeLessThanOrEqual(1000);
  });

  it('should keep most recent entries when limiting', () => {
    const { addEntry } = useLogStore.getState();

    // Add 1005 entries
    for (let i = 0; i < 1005; i++) {
      addEntry({
        timestamp: '2024-01-01 12:00:00',
        level: 'INFO',
        component: 'Test',
        message: `Message ${i}`,
      });
    }

    const entries = useLogStore.getState().entries;
    const lastEntry = entries[entries.length - 1];
    expect(lastEntry.message).toBe('Message 1004');
  });

  it('should set entries', () => {
    const { setEntries } = useLogStore.getState();

    const mockEntries = [
      {
        timestamp: '2024-01-01 12:00:00',
        level: 'INFO' as LogLevel,
        component: 'Test',
        message: 'Test 1',
      },
      {
        timestamp: '2024-01-01 12:01:00',
        level: 'ERROR' as LogLevel,
        component: 'Test',
        message: 'Test 2',
      },
    ];

    setEntries(mockEntries);

    expect(useLogStore.getState().entries).toEqual(mockEntries);
  });

  it('should clear logs', () => {
    const { setEntries, clearLogs } = useLogStore.getState();

    setEntries([
      {
        timestamp: '2024-01-01 12:00:00',
        level: 'INFO',
        component: 'Test',
        message: 'Test',
      },
    ]);

    clearLogs();

    expect(useLogStore.getState().entries).toEqual([]);
  });

  it('should get logs by level', () => {
    const { setEntries } = useLogStore.getState();

    setEntries([
      { timestamp: '12:00:00', level: 'DEBUG' as LogLevel, component: 'Test', message: 'Debug' },
      { timestamp: '12:01:00', level: 'INFO' as LogLevel, component: 'Test', message: 'Info' },
      { timestamp: '12:02:00', level: 'ERROR' as LogLevel, component: 'Test', message: 'Error' },
      { timestamp: '12:03:00', level: 'ERROR' as LogLevel, component: 'Test', message: 'Error 2' },
    ]);

    const errorLogs = useLogStore.getState().getLogsByLevel('ERROR');
    expect(errorLogs).toHaveLength(2);
    expect(errorLogs.every(e => e.level === 'ERROR')).toBe(true);
  });

  it('should get recent logs', () => {
    const { setEntries } = useLogStore.getState();

    setEntries([
      { timestamp: '12:00:00', level: 'INFO', component: 'Test', message: '1' },
      { timestamp: '12:01:00', level: 'INFO', component: 'Test', message: '2' },
      { timestamp: '12:02:00', level: 'INFO', component: 'Test', message: '3' },
      { timestamp: '12:03:00', level: 'INFO', component: 'Test', message: '4' },
      { timestamp: '12:04:00', level: 'INFO', component: 'Test', message: '5' },
    ]);

    const recent = useLogStore.getState().getRecentLogs(3);
    expect(recent).toHaveLength(3);
    expect(recent[0].message).toBe('3');
    expect(recent[2].message).toBe('5');
  });

  it('should handle getRecentLogs with more than available', () => {
    const { setEntries } = useLogStore.getState();

    setEntries([
      { timestamp: '12:00:00', level: 'INFO', component: 'Test', message: '1' },
      { timestamp: '12:01:00', level: 'INFO', component: 'Test', message: '2' },
    ]);

    const recent = useLogStore.getState().getRecentLogs(10);
    expect(recent).toHaveLength(2);
  });

  it('should load logs', async () => {
    const { loadLogs } = useLogStore.getState();

    await loadLogs();

    expect(useLogStore.getState().isLoading).toBe(false);
    expect(useLogStore.getState().error).toBeNull();
  });

  it('should export logs successfully', async () => {
    const { exportLogs } = useLogStore.getState();
    const { ExportLogs } = await import('../../wailsjs/go/app/App');

    await exportLogs();

    expect(ExportLogs).toHaveBeenCalledWith('');
    expect(useLogStore.getState().isLoading).toBe(false);
    expect(useLogStore.getState().error).toBeNull();
  });

  it('should handle export logs error', async () => {
    const { exportLogs } = useLogStore.getState();
    const { ExportLogs } = await import('../../wailsjs/go/app/App');

    const mockError = new Error('Export failed');
    (ExportLogs as unknown as ReturnType<typeof vi.fn>).mockRejectedValue(mockError);

    await exportLogs();

    expect(useLogStore.getState().error).toBe('Export failed');
    expect(useLogStore.getState().isLoading).toBe(false);
  });

  it('should handle non-Error export failure', async () => {
    const { exportLogs } = useLogStore.getState();
    const { ExportLogs } = await import('../../wailsjs/go/app/App');

    (ExportLogs as unknown as ReturnType<typeof vi.fn>).mockRejectedValue('string error');

    await exportLogs();

    expect(useLogStore.getState().error).toBe('Failed to export logs');
  });

  it('should get logs by different levels', () => {
    const { setEntries } = useLogStore.getState();

    setEntries([
      { timestamp: '12:00:00', level: 'DEBUG' as LogLevel, component: 'Test', message: 'Debug' },
      { timestamp: '12:01:00', level: 'INFO' as LogLevel, component: 'Test', message: 'Info' },
      { timestamp: '12:02:00', level: 'WARN' as LogLevel, component: 'Test', message: 'Warn' },
      { timestamp: '12:03:00', level: 'ERROR' as LogLevel, component: 'Test', message: 'Error' },
    ]);

    expect(useLogStore.getState().getLogsByLevel('DEBUG')).toHaveLength(1);
    expect(useLogStore.getState().getLogsByLevel('INFO')).toHaveLength(1);
    expect(useLogStore.getState().getLogsByLevel('WARN')).toHaveLength(1);
    expect(useLogStore.getState().getLogsByLevel('ERROR')).toHaveLength(1);
  });

  it('should return empty array for level with no entries', () => {
    const { setEntries } = useLogStore.getState();

    setEntries([
      { timestamp: '12:00:00', level: 'INFO' as LogLevel, component: 'Test', message: 'Info' },
    ]);

    expect(useLogStore.getState().getLogsByLevel('ERROR')).toEqual([]);
  });

  it('should maintain entry order', () => {
    const { addEntry } = useLogStore.getState();

    addEntry({ timestamp: '12:00:00', level: 'INFO', component: 'Test', message: 'First' });
    addEntry({ timestamp: '12:01:00', level: 'INFO', component: 'Test', message: 'Second' });
    addEntry({ timestamp: '12:02:00', level: 'INFO', component: 'Test', message: 'Third' });

    const entries = useLogStore.getState().entries;
    expect(entries[0].message).toBe('First');
    expect(entries[1].message).toBe('Second');
    expect(entries[2].message).toBe('Third');
  });
});
