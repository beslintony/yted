import { Button, Group } from '@mantine/core';
import { IconDeviceFloppy, IconRefresh } from '@tabler/icons-react';
import { memo } from 'react';

interface ActionsSectionProps {
  hasChanges: boolean;
  saving: boolean;
  handleReset: () => void;
  handleSave: () => void;
}

// Memoized: only hasChanges/saving flips re-render this section; the
// handlers are stable across keystrokes (see SettingsPage).
export const ActionsSection = memo(function ActionsSection({
  hasChanges,
  saving,
  handleReset,
  handleSave,
}: ActionsSectionProps) {
  return (
    <Group justify="flex-end">
      {hasChanges && (
        <Button
          color="gray"
          leftSection={<IconRefresh size={16} />}
          variant="light"
          onClick={handleReset}
        >
          Reset
        </Button>
      )}
      <Button
        color="yted"
        disabled={!hasChanges}
        leftSection={<IconDeviceFloppy size={16} />}
        loading={saving}
        onClick={handleSave}
      >
        Save Settings
      </Button>
    </Group>
  );
});
