import { Button, Group, Paper, Stack, Text, Tooltip } from '@mantine/core';
import { IconTrash } from '@tabler/icons-react';
import { memo } from 'react';

import {
  ClearCompletedDownloads,
  ClearCompletedDownloadsCache,
  ClearDownloadCache,
} from '../../../wailsjs/go/app/App';
import { ConfirmOptions } from '../../types';

interface CacheSectionProps {
  dark: boolean;
  success: (title: string, message?: string) => void;
  error: (title: string, message?: string) => void;
  confirm: (options: ConfirmOptions) => void;
}

// Memoized: this section takes no settings state and its callbacks are
// stable, so keystrokes elsewhere never re-render it.
export const CacheSection = memo(function CacheSection({
  dark,
  success,
  error,
  confirm,
}: CacheSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          Cache Management
        </Text>
        <Text c={dark ? 'dimmed' : 'gray.6'} size="sm">
          Clear cached data to free up space or fix issues. This action cannot be undone.
        </Text>

        <Group gap="sm">
          <Tooltip label="Clear all download queue history">
            <Button
              color="orange"
              leftSection={<IconTrash size={16} />}
              variant="light"
              onClick={() => {
                confirm({
                  title: 'Clear Download Cache?',
                  message:
                    'This will remove ALL pending and completed downloads from the database. This action cannot be undone.',
                  confirmLabel: 'Clear All',
                  confirmColor: 'orange',
                  onConfirm: async () => {
                    try {
                      await ClearDownloadCache();
                      success('Cache Cleared', 'Download cache has been cleared successfully');
                      // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    } catch (err: any) {
                      error('Clear Failed', 'Failed to clear download cache: ' + err?.message);
                    }
                  },
                });
              }}
            >
              Clear Download Cache
            </Button>
          </Tooltip>

          <Tooltip label="Clear only completed downloads from history">
            <Button
              color="gray"
              leftSection={<IconTrash size={16} />}
              variant="light"
              onClick={() => {
                confirm({
                  title: 'Clear Completed Downloads?',
                  message:
                    'This will remove completed download records from the database. Active downloads will not be affected.',
                  confirmLabel: 'Clear Completed',
                  confirmColor: 'orange',
                  onConfirm: async () => {
                    try {
                      await ClearCompletedDownloadsCache();
                      success('Cache Cleared', 'Completed downloads cache has been cleared');
                      // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    } catch (err: any) {
                      error('Clear Failed', 'Failed to clear completed downloads: ' + err?.message);
                    }
                  },
                });
              }}
            >
              Clear Completed Only
            </Button>
          </Tooltip>

          <Tooltip label="Clear completed downloads from queue">
            <Button
              color="gray"
              leftSection={<IconTrash size={16} />}
              variant="light"
              onClick={() => {
                confirm({
                  title: 'Clear Queue History?',
                  message:
                    'This will remove completed downloads from the queue view. Downloads will remain in the library.',
                  confirmLabel: 'Clear Queue',
                  confirmColor: 'gray',
                  onConfirm: async () => {
                    try {
                      await ClearCompletedDownloads();
                      success('Queue Cleared', 'Completed downloads cleared from queue');
                      // eslint-disable-next-line @typescript-eslint/no-explicit-any
                    } catch (err: any) {
                      error('Clear Failed', 'Failed to clear queue: ' + err?.message);
                    }
                  },
                });
              }}
            >
              Clear Queue Completed
            </Button>
          </Tooltip>
        </Group>
      </Stack>
    </Paper>
  );
});
