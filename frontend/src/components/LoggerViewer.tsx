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
import { useCallback, useEffect, useState } from 'react';

import { ClearLogs, ExportLogs, GetLogs } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';
import { LogEntry, LogLevel, useLogStore } from '../stores';

export function LoggerViewer() {
  const { entries, setEntries, clearLogs, addEntry } = useLogStore();
  const [filter, setFilter] = useState<LogLevel | 'ALL'>('ALL');
  const [search, setSearch] = useState('');
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

  // Load initial logs
  const loadLogs = useCallback(async () => {
    try {
      const logs = await GetLogs(100);
      setEntries(logs as LogEntry[]);
    } catch (err) {
      console.error('Failed to load logs:', err);
    }
  }, [setEntries]);

  // Listen for log events from backend
  useEffect(() => {
    const cancel = EventsOn('log:new', (data: LogEntry) => {
      addEntry(data);
    });
    return () => cancel();
  }, [addEntry]);

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

  const filteredEntries = entries.filter((entry: LogEntry) => {
    const matchesFilter = filter === 'ALL' || entry.level === filter;
    const matchesSearch =
      search === '' ||
      entry.message.toLowerCase().includes(search.toLowerCase()) ||
      entry.component.toLowerCase().includes(search.toLowerCase());
    return matchesFilter && matchesSearch;
  });

  const getLevelColor = (level: LogLevel) => {
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
  };

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
            {filteredEntries.length === 0 ? (
              <Text c="dimmed" ta="center">
                No logs found
              </Text>
            ) : (
              filteredEntries.map((entry: LogEntry, index: number) => (
                <Paper
                  key={index}
                  p="xs"
                  style={{
                    background: dark ? '#25262b' : '#fff',
                    borderLeft: `3px solid ${
                      entry.level === 'ERROR'
                        ? '#fa5252'
                        : entry.level === 'WARN'
                          ? '#fab005'
                          : '#228be6'
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
              ))
            )}
          </Stack>
        </ScrollArea>
      </Paper>
    </Stack>
  );
}
