import {
  ActionIcon,
  Badge,
  Group,
  Paper,
  ScrollArea,
  Select,
  Stack,
  Text,
  TextInput,
  Tooltip,
  useMantineColorScheme,
} from '@mantine/core';
import { IconDownload, IconRefresh, IconSearch, IconTrash, IconX } from '@tabler/icons-react';
import { memo, useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { ClearLogs, ExportLogs, GetLogs } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
import { LogEntry, LogLevel, useLogStore } from '../stores';

// Cap on rendered rows: the store keeps the last 1000 entries, but mounting
// 1000 Paper rows per update is the dominant cost of this view. Render only
// the latest 200 matches with a "showing X of Y" note instead.
export const MAX_VISIBLE_LOG_ROWS = 200;

function getLevelColor(level: LogLevel) {
  switch (level) {
    case 'ERROR':
      return 'red';
    case 'WARN':
      return 'yellow';
    case 'INFO':
      return 'blue';
    case 'DEBUG':
      return 'gray';
    default:
      return 'gray';
  }
}

interface LogRowProps {
  dark: boolean;
  entry: LogEntry;
}

// Memoized so that appends (which reuse the same entry object references via
// [...entries, entry]) don't re-render existing rows.
// Key note: LogEntry has no unique id — timestamp/level/component/message are
// all non-unique (see logStore.test.ts, which pins the exact entry shape, so
// we cannot inject ids without breaking store semantics). Callers therefore
// key rows by list offset (see visibleEntries below); index keys stay stable
// for appends at the tail.
export const LogRow = memo(function LogRow({ dark, entry }: LogRowProps) {
  return (
    <Paper
      p="xs"
      style={{
        background: dark ? '#25262b' : '#fff',
        borderLeft: `3px solid ${
          entry.level === 'ERROR' ? '#fa5252' : entry.level === 'WARN' ? '#fab005' : '#228be6'
        }`,
      }}
    >
      <Group gap="xs" wrap="nowrap">
        <Badge color={getLevelColor(entry.level)} size="sm" variant="light">
          {entry.level}
        </Badge>
        <Text c="dimmed" size="xs" style={{ whiteSpace: 'nowrap' }}>
          {new Date(entry.timestamp).toLocaleTimeString()}
        </Text>
        <Text fw={500} size="sm" style={{ whiteSpace: 'nowrap' }}>
          [{entry.component}]
        </Text>
        <Text size="sm" style={{ flex: 1, wordBreak: 'break-word' }}>
          {entry.message}
        </Text>
      </Group>
      {entry.error && (
        <Text c="red" mt="xs" pl="md" size="xs">
          {entry.error}
        </Text>
      )}
    </Paper>
  );
});

export function LoggerViewer() {
  // Select slices individually so unrelated store fields (isLoading/error)
  // don't re-render this view.
  const entries = useLogStore(s => s.entries);
  const setEntries = useLogStore(s => s.setEntries);
  const clearLogs = useLogStore(s => s.clearLogs);
  const [filter, setFilter] = useState<LogLevel | 'ALL'>('ALL');
  const [search, setSearch] = useState('');
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

  // Coalescing buffer for the high-frequency `log:new` event: rapid backend
  // log storms previously caused one store notify + full re-render per entry.
  // Events are buffered and flushed as a single addEntries batch on the next
  // microtask (same-tick flush). The store's synchronous addEntry path is
  // untouched — only this subscription coalesces.
  const pendingRef = useRef<LogEntry[]>([]);
  const flushScheduledRef = useRef(false);

  // Load initial logs
  const loadLogs = useCallback(async () => {
    try {
      const logs = await GetLogs(100);
      setEntries(logs as LogEntry[]);
    } catch (err) {
      console.error('Failed to load logs:', err);
    }
  }, [setEntries]);

  // Listen for log events from backend (coalesced, see above)
  useEffect(() => {
    const flush = () => {
      flushScheduledRef.current = false;
      if (pendingRef.current.length === 0) {
        return;
      }
      const batch = pendingRef.current;
      pendingRef.current = [];
      useLogStore.getState().addEntries(batch);
    };
    const scheduleFlush = () => {
      if (flushScheduledRef.current) {
        return;
      }
      flushScheduledRef.current = true;
      if (typeof queueMicrotask === 'function') {
        queueMicrotask(flush);
      } else {
        setTimeout(flush, 0);
      }
    };
    const cancel = EventsOn('log:new', (data: LogEntry) => {
      pendingRef.current.push(data);
      scheduleFlush();
    });
    return () => {
      cancel();
      // Don't drop logs that arrived in the same tick as unmount.
      flush();
    };
  }, []);

  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  const handleClear = async () => {
    await ClearLogs();
    clearLogs();
  };

  const handleExport = async () => {
    try {
      // ExportLogs now uses the configured log export path
      await ExportLogs('');
    } catch (err) {
      console.error('Failed to export logs:', err);
    }
  };

  const filteredEntries = useMemo(() => {
    const query = search.toLowerCase();
    return entries.filter((entry: LogEntry) => {
      const matchesFilter = filter === 'ALL' || entry.level === filter;
      const matchesSearch =
        query === '' ||
        entry.message.toLowerCase().includes(query) ||
        entry.component.toLowerCase().includes(query);
      return matchesFilter && matchesSearch;
    });
  }, [entries, filter, search]);

  const visibleEntries = useMemo(
    () => filteredEntries.slice(-MAX_VISIBLE_LOG_ROWS),
    [filteredEntries]
  );
  // Offset of the visible window within the filtered list. Used as the key
  // base so rows keep stable identities when the tail window slides on
  // append (plain slice indices would shift every row's key per entry).
  const visibleOffset = filteredEntries.length - visibleEntries.length;

  return (
    <Stack gap="md" style={{ height: '100%' }}>
      <Group gap="sm">
        <Select
          data={[
            { value: 'ALL', label: 'All Levels' },
            { value: 'DEBUG', label: 'Debug' },
            { value: 'INFO', label: 'Info' },
            { value: 'WARN', label: 'Warning' },
            { value: 'ERROR', label: 'Error' },
          ]}
          style={{ width: 150 }}
          value={filter}
          onChange={value => setFilter((value as LogLevel) || 'ALL')}
        />
        <TextInput
          leftSection={<IconSearch size={16} />}
          placeholder="Search logs..."
          rightSection={
            search ? (
              <ActionIcon color="gray" variant="subtle" onClick={() => setSearch('')}>
                <IconX size={14} />
              </ActionIcon>
            ) : undefined
          }
          style={{ flex: 1 }}
          value={search}
          onChange={e => setSearch(e.currentTarget.value)}
        />
        <Tooltip label="Refresh">
          <ActionIcon color="blue" variant="light" onClick={loadLogs}>
            <IconRefresh size={18} />
          </ActionIcon>
        </Tooltip>
        <Tooltip label="Export">
          <ActionIcon color="green" variant="light" onClick={handleExport}>
            <IconDownload size={18} />
          </ActionIcon>
        </Tooltip>
        <Tooltip label="Clear">
          <ActionIcon color="red" variant="light" onClick={handleClear}>
            <IconTrash size={18} />
          </ActionIcon>
        </Tooltip>
      </Group>

      {filteredEntries.length > visibleEntries.length && (
        <Text c="dimmed" size="xs">
          Showing {visibleEntries.length} of {filteredEntries.length} matching logs (latest{' '}
          {MAX_VISIBLE_LOG_ROWS})
        </Text>
      )}

      <Paper
        p="xs"
        style={{
          background: dark ? '#1a1b1e' : '#f8f9fa',
          border: `1px solid ${dark ? '#373a40' : '#dee2e6'}`,
          flex: 1,
          minHeight: 0,
        }}
      >
        <ScrollArea h="calc(100vh - 300px)">
          <Stack gap="xs">
            {visibleEntries.length === 0 ? (
              <Text c="dimmed" ta="center">
                No logs found
              </Text>
            ) : (
              visibleEntries.map((entry: LogEntry, index: number) => (
                <LogRow key={visibleOffset + index} dark={dark} entry={entry} />
              ))
            )}
          </Stack>
        </ScrollArea>
      </Paper>
    </Stack>
  );
}
