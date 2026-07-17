import { Button, Group, NumberInput, Paper, Stack, Text, TextInput } from '@mantine/core';
import { IconFolder } from '@tabler/icons-react';

import { config } from '../../../wailsjs/go/models';

interface LoggingSectionProps {
  dark: boolean;
  settings: config.Config;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
  handleBrowseLogPath: () => void;
  handleBrowseLogExportPath: () => void;
}

export function LoggingSection({
  dark,
  settings,
  updateSetting,
  handleBrowseLogPath,
  handleBrowseLogExportPath,
}: LoggingSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          Logging
        </Text>

        <Group align="flex-end" gap="sm">
          <TextInput
            readOnly
            description="Where application logs are stored (default: ~/.yted/.logs)"
            label="Log Storage Path"
            style={{ flex: 1 }}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.log_path || ''}
          />
          <Button
            color="yted"
            leftSection={<IconFolder size={16} />}
            variant="light"
            onClick={handleBrowseLogPath}
          >
            Browse
          </Button>
        </Group>

        <NumberInput
          description="Number of log sessions to keep (1-100). Each session is one app start."
          label="Max Log Sessions"
          max={100}
          min={1}
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.max_log_sessions || 10}
          w={200}
          onChange={v => updateSetting('max_log_sessions', v || 10)}
        />

        <Group align="flex-end" gap="sm">
          <TextInput
            readOnly
            description="Where exported logs are saved"
            label="Log Export Path"
            style={{ flex: 1 }}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.log_export_path || ''}
          />
          <Button
            color="yted"
            leftSection={<IconFolder size={16} />}
            variant="light"
            onClick={handleBrowseLogExportPath}
          >
            Browse
          </Button>
        </Group>
      </Stack>
    </Paper>
  );
}
