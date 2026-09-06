import {
  ActionIcon,
  Badge,
  Button,
  Group,
  Modal,
  Paper,
  Select,
  Stack,
  Table,
  Text,
  TextInput,
  Tooltip,
} from '@mantine/core';
import { IconEdit, IconPlus, IconTrash } from '@tabler/icons-react';
import { memo, useEffect, useState } from 'react';

import { config } from '../../../wailsjs/go/models';

interface PresetsSectionProps {
  dark: boolean;
  settings: config.Config;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
}

// Memoized: skips re-render unless this section's own props change (see
// DownloadsSection for the rationale).
export const PresetsSection = memo(function PresetsSection({
  dark,
  settings,
  updateSetting,
}: PresetsSectionProps) {
  const [presetModalOpen, setPresetModalOpen] = useState(false);
  const [editingPreset, setEditingPreset] = useState<config.DownloadPreset | null>(null);

  return (
    <>
      <Paper
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        p="md"
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Group justify="space-between">
            <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
              Download Presets
            </Text>
            <Button
              color="yted"
              leftSection={<IconPlus size={16} />}
              size="sm"
              onClick={() => {
                setEditingPreset(null);
                setPresetModalOpen(true);
              }}
            >
              Add Preset
            </Button>
          </Group>

          <Table>
            <Table.Thead>
              <Table.Tr>
                <Table.Th c={dark ? '#c1c2c5' : '#495057'}>Name</Table.Th>
                <Table.Th c={dark ? '#c1c2c5' : '#495057'}>Format</Table.Th>
                <Table.Th c={dark ? '#c1c2c5' : '#495057'}>Quality</Table.Th>
                <Table.Th c={dark ? '#c1c2c5' : '#495057'}>Extension</Table.Th>
                <Table.Th c={dark ? '#c1c2c5' : '#495057'}>Actions</Table.Th>
              </Table.Tr>
            </Table.Thead>
            <Table.Tbody>
              {settings.download_presets?.map(preset => (
                <Table.Tr key={preset.id}>
                  <Table.Td c={dark ? '#fff' : '#000'}>{preset.name}</Table.Td>
                  <Table.Td>
                    <code style={{ color: dark ? '#c1c2c5' : '#495057' }}>{preset.format}</code>
                  </Table.Td>
                  <Table.Td>
                    <Badge color="yted">{preset.quality}</Badge>
                  </Table.Td>
                  <Table.Td c={dark ? '#c1c2c5' : '#495057'}>{preset.extension}</Table.Td>
                  <Table.Td>
                    <Group gap={4}>
                      <Tooltip label="Edit">
                        <ActionIcon
                          size="sm"
                          onClick={() => {
                            setEditingPreset(preset);
                            setPresetModalOpen(true);
                          }}
                        >
                          <IconEdit size={14} />
                        </ActionIcon>
                      </Tooltip>
                      <Tooltip label="Delete">
                        <ActionIcon
                          color="red"
                          size="sm"
                          onClick={() => {
                            const newPresets = settings.download_presets.filter(
                              p => p.id !== preset.id
                            );
                            updateSetting('download_presets', newPresets);
                          }}
                        >
                          <IconTrash size={14} />
                        </ActionIcon>
                      </Tooltip>
                    </Group>
                  </Table.Td>
                </Table.Tr>
              ))}
            </Table.Tbody>
          </Table>
        </Stack>
      </Paper>

      {/* Preset Modal */}
      <Modal
        opened={presetModalOpen}
        title={editingPreset ? 'Edit Preset' : 'Add Preset'}
        onClose={() => setPresetModalOpen(false)}
      >
        <PresetForm
          preset={editingPreset}
          onCancel={() => setPresetModalOpen(false)}
          onSave={preset => {
            const currentPresets = settings.download_presets || [];
            if (editingPreset) {
              const updatedPresets = currentPresets.map(p => (p.id === preset.id ? preset : p));
              updateSetting('download_presets', updatedPresets);
            } else {
              const newPreset = { ...preset, id: crypto.randomUUID() };
              updateSetting('download_presets', [...currentPresets, newPreset]);
            }
            setPresetModalOpen(false);
          }}
        />
      </Modal>
    </>
  );
});

function PresetForm({
  preset,
  onSave,
  onCancel,
}: {
  preset: config.DownloadPreset | null;
  onSave: (preset: config.DownloadPreset) => void;
  onCancel: () => void;
}) {
  const [name, setName] = useState(preset?.name || '');
  const [format, setFormat] = useState(preset?.format || 'best');
  const [quality, setQuality] = useState(preset?.quality || 'best');
  const [extension, setExtension] = useState(preset?.extension || 'mp4');

  useEffect(() => {
    setName(preset?.name || '');
    setFormat(preset?.format || 'best');
    setQuality(preset?.quality || 'best');
    setExtension(preset?.extension || 'mp4');
  }, [preset]);

  return (
    <Stack gap="md">
      <TextInput
        label="Name"
        placeholder="e.g., 1080p Video"
        value={name}
        onChange={e => setName(e.currentTarget.value)}
      />
      <TextInput
        label="Format"
        placeholder="e.g., bestvideo[height<=1080]+bestaudio"
        value={format}
        onChange={e => setFormat(e.currentTarget.value)}
      />
      <Select
        data={[
          { value: 'best', label: 'Best' },
          { value: '1080p', label: '1080p' },
          { value: '720p', label: '720p' },
          { value: '480p', label: '480p' },
          { value: 'audio', label: 'Audio' },
        ]}
        label="Quality"
        value={quality}
        onChange={v => v && setQuality(v)}
      />
      <TextInput
        label="Extension"
        placeholder="e.g., mp4"
        value={extension}
        onChange={e => setExtension(e.currentTarget.value)}
      />
      <Group justify="flex-end" mt="md">
        <Button variant="light" onClick={onCancel}>
          Cancel
        </Button>
        <Button
          color="yted"
          disabled={!name || !format}
          onClick={() =>
            onSave({
              id: preset?.id || '',
              name,
              format,
              quality,
              extension,
              // eslint-disable-next-line @typescript-eslint/no-explicit-any
            } as any)
          }
        >
          Save
        </Button>
      </Group>
    </Stack>
  );
}
