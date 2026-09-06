import { NumberInput, Paper, Stack, Switch, Text } from '@mantine/core';
import { memo } from 'react';

import { config } from '../../../wailsjs/go/models';

interface PlayerSectionProps {
  dark: boolean;
  settings: config.Config;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
}

// Memoized: skips re-render unless this section's own props change (see
// DownloadsSection for the rationale).
export const PlayerSection = memo(function PlayerSection({
  dark,
  settings,
  updateSetting,
}: PlayerSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          Player
        </Text>

        <NumberInput
          label="Default Volume"
          max={100}
          min={0}
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.default_volume}
          w={200}
          onChange={v => updateSetting('default_volume', v || 80)}
        />

        <Switch
          checked={settings.remember_position}
          label="Remember Watch Position"
          styles={{
            label: {
              color: dark ? '#c1c2c5' : '#495057',
            },
          }}
          onChange={e => updateSetting('remember_position', e.currentTarget.checked)}
        />
      </Stack>
    </Paper>
  );
});
