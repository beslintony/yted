import { ColorInput, Paper, Select, Stack, Switch, Text } from '@mantine/core';
import { memo } from 'react';

import { config } from '../../../wailsjs/go/models';

interface UiSectionProps {
  dark: boolean;
  settings: config.Config;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
  handleThemeChange: (value: string) => void;
  toggleSidebar: () => void;
}

// Memoized: skips re-render unless this section's own props change (see
// DownloadsSection for the rationale).
export const UiSection = memo(function UiSection({
  dark,
  settings,
  updateSetting,
  handleThemeChange,
  toggleSidebar,
}: UiSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          UI
        </Text>

        <Select
          data={[
            { value: 'dark', label: 'Dark' },
            { value: 'light', label: 'Light' },
            { value: 'auto', label: 'Auto' },
          ]}
          label="Theme"
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.theme}
          w={200}
          onChange={v => v && handleThemeChange(v)}
        />

        <ColorInput
          label="Accent Color"
          value={settings.accent_color}
          w={200}
          onChange={v => updateSetting('accent_color', v)}
        />

        <Switch
          checked={settings.sidebar_collapsed}
          label="Collapse Sidebar"
          styles={{
            label: {
              color: dark ? '#c1c2c5' : '#495057',
            },
          }}
          onChange={e => {
            updateSetting('sidebar_collapsed', e.currentTarget.checked);
            toggleSidebar();
          }}
        />
      </Stack>
    </Paper>
  );
});
