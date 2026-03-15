import { useState, useEffect } from 'react';
import {
  Paper,
  Text,
  Group,
  Button,
  ScrollArea,
  Badge,
  Stack,
  ActionIcon,
  Tooltip,
  Select,
  TextInput,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconTrash,
  IconDownload,
  IconRefresh,
  IconSearch,
  IconX,
} from '@tabler/icons-react';
import { useLogStore, LogEntry, LogLevel } from '../stores';
import { GetLogs, ClearLogs, ExportLogs } from '../../wailsjs/go/app/App';
import { EventsOn } from '../../wailsjs/runtime';

export function LoggerViewer() {
  const { entries, setEntries, clearLogs, addEntry } = useLogStore();
  const [filter, setFilter] = useState<LogLevel | 'ALL'>('ALL');
  const [search, setSearch] = useState('');
  const [autoScroll, setAutoScroll] = useState(true);
  const { colorScheme } = useMantineColorScheme();
  const dark = colorScheme === 'dark';

  // Listen for log events from backend
  useEffect(() => {
    const cancel = EventsOn('log:new', (data: LogEntry) => {
      addEntry(data);
    });
    return () => cancel();
  }, []);

  // Load initial logs
  const loadLogs = async () => {
    try {
      const logs = await GetLogs(100);
      setEntries(logs as any);
    } catch (err) {
      console.error('Failed to load logs:', err);
    }
  };

  useEffect(() => {
    loadLogs();
  }, []);

  const handleClear = async () => {
    await ClearLogs();
    clearLogs();
  };

  const handleExport = async () => {
    try {
      const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
      const filename = `yted-logs-${timestamp}.txt`;
      await ExportLogs(filename);
    } catch (err) {
      console.error('Failed to export logs:', err);
    }
  };

  const filteredEntries = entries
    .filter((entry) => filter === 'ALL' || entry.level === filter)
    .filter((entry) => {
      if (!search) return true;
      const searchLower = search.toLowerCase();
      return (
        entry.message.toLowerCase().includes(searchLower) ||
        entry.component.toLowerCase().includes(searchLower) ||
        entry.error?.toLowerCase().includes(searchLower)
      );
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
    <Paper
      p="md"
      withBorder
      sx={{
        background: dark ? '#25262b' : '#fff',
        borderColor: dark ? '#373a40' : '#dee2e6',
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      <Group position="apart" mb="md">
        <Text fw={600} c={dark ? '#fff' : '#000'}>
          Application Logs ({filteredEntries.length})
        </Text>
        <Group spacing="xs">
          <Tooltip label="Refresh">
            <ActionIcon onClick={loadLogs} variant="light" color="gray">
              <IconRefresh size={16} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Export">
            <ActionIcon onClick={handleExport} variant="light" color="blue">
              <IconDownload size={16} />
            </ActionIcon>
          </Tooltip>
          <Tooltip label="Clear">
            <ActionIcon onClick={handleClear} variant="light" color="red">
              <IconTrash size={16} />
            </ActionIcon>
          </Tooltip>
        </Group>
      </Group>

      <Group spacing="sm" mb="md">
        <Select
          value={filter}
          onChange={(v) => setFilter(v as LogLevel | 'ALL')}
          data={[
            { value: 'ALL', label: 'All Levels' },
            { value: 'DEBUG', label: 'Debug' },
            { value: 'INFO', label: 'Info' },
            { value: 'WARN', label: 'Warning' },
            { value: 'ERROR', label: 'Error' },
          ]}
          size="sm"
          w={150}
        />
        <TextInput
          placeholder="Search logs..."
          value={search}
          onChange={(e) => setSearch(e.currentTarget.value)}
          size="sm"
          sx={{ flex: 1 }}
          icon={<IconSearch size={14} />}
          rightSection={
            search && (
              <ActionIcon onClick={() => setSearch('')} size="sm" color="gray">
                <IconX size={14} />
              </ActionIcon>
            )
          }
        />
      </Group>

      <ScrollArea
        sx={{
          flex: 1,
          background: dark ? '#1a1b1e' : '#f8f9fa',
          borderRadius: 4,
          border: `1px solid ${dark ? '#373a40' : '#dee2e6'}`,
        }}
      >
        <Stack spacing={0} p="xs">
          {filteredEntries.length === 0 ? (
            <Text c={dark ? 'dimmed' : 'gray.6'} align="center" py="xl">
              No logs to display
            </Text>
          ) : (
            filteredEntries.map((entry, index) => (
              <Paper
                key={index}
                p="xs"
                sx={{
                  background: 'transparent',
                  borderBottom: `1px solid ${dark ? '#373a40' : '#e9ecef'}`,
                  '&:last-child': { borderBottom: 'none' },
                }}
              >
                <Group spacing="xs" align="flex-start" noWrap>
                  <Badge
                    size="xs"
                    color={getLevelColor(entry.level)}
                    sx={{ minWidth: 60 }}
                  >
                    {entry.level}
                  </Badge>
                  <Text size="xs" c={dark ? 'dimmed' : 'gray.6'} sx={{ minWidth: 130 }}>
                    {entry.timestamp}
                  </Text>
                  <Text size="xs" fw={500} c={dark ? 'gray.4' : 'gray.7'} sx={{ minWidth: 100 }}>
                    [{entry.component}]
                  </Text>
                  <Text size="xs" c={dark ? '#fff' : '#000'} sx={{ flex: 1 }}>
                    {entry.message}
                  </Text>
                </Group>
                {entry.error && (
                  <Text size="xs" c="red" mt={4} ml={300}>
                    Error: {entry.error}
                  </Text>
                )}
              </Paper>
            ))
          )}
        </Stack>
      </ScrollArea>
    </Paper>
  );
}
