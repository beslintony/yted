
import {
  ActionIcon,
  Badge,
  Button,
  ColorInput,
  Group,
  Modal,
  NumberInput,
  Paper,
  Select,
  Stack,
  Switch,
  Table,
  Text,
  TextInput,
  Tooltip,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconDeviceFloppy,
  IconEdit,
  IconFolder,
  IconPlus,
  IconRefresh,
  IconTrash,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import {
  ClearCompletedDownloads,
  ClearCompletedDownloadsCache,
  ClearDownloadCache,
  GetSettings,
  SaveSettings,
  ShowFFmpegDialog,
  ShowOpenDirectoryDialog,
} from '../../wailsjs/go/app/App';
import { config } from '../../wailsjs/go/models';
import { useNotifications, useSettingsStore } from '../stores';
import { QualityOption, ThemeMode } from '../types';

export function SettingsPage() {
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const [settings, setSettings] = useState<config.Config | null>(null);
  const [originalSettings, setOriginalSettings] = useState<config.Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [presetModalOpen, setPresetModalOpen] = useState(false);
  const [editingPreset, setEditingPreset] = useState<config.DownloadPreset | null>(null);
  const [saveError, setSaveError] = useState<string | null>(null);
  const [saveSuccess, setSaveSuccess] = useState(false);

  const {
    setTheme,
    toggleSidebar,
    setDownloadPath,
    setMaxConcurrentDownloads,
    setDefaultQuality,
    saveSettings: saveSettingsToStore,
  } = useSettingsStore();

  const { success, error, confirm } = useNotifications();

  const dark = colorScheme === 'dark';

  useEffect(() => {
    loadSettings();
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  useEffect(() => {
    if (originalSettings && settings) {
      const changed = JSON.stringify(originalSettings) !== JSON.stringify(settings);
      setHasChanges(changed);
    }
  }, [settings, originalSettings]);

  const loadSettings = async () => {
    try {
      const result = await GetSettings();
      if (result) {
        setSettings(result);
        setOriginalSettings(JSON.parse(JSON.stringify(result)));

        setDownloadPath(result.download_path);
        setMaxConcurrentDownloads(result.max_concurrent_downloads);
        setDefaultQuality(result.default_quality as QualityOption);
        setTheme(result.theme as ThemeMode);
      }
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSave = async () => {
    if (!settings) return;

    setSaving(true);
    setSaveError(null);
    setSaveSuccess(false);

    try {
      await SaveSettings(settings);
      setOriginalSettings(JSON.parse(JSON.stringify(settings)));
      setHasChanges(false);
      setSaveSuccess(true);

      await saveSettingsToStore();

      if (settings.theme === 'dark' && colorScheme !== 'dark') {
        setColorScheme('dark');
      } else if (settings.theme === 'light' && colorScheme !== 'light') {
        setColorScheme('light');
      }

      success('Settings Saved', 'Your settings have been saved successfully');
      setTimeout(() => setSaveSuccess(false), 3000);
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (err: any) {
      console.error('Failed to save settings:', err);
      const msg = err?.message || 'Failed to save settings';
      setSaveError(msg);
      error('Save Failed', msg);
    } finally {
      setSaving(false);
    }
  };

  const handleBrowseDownloadPath = async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, download_path: path } as any) : null));
        setDownloadPath(path);
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  };

  const handleBrowseLogExportPath = async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, log_export_path: path } as any) : null));
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  };

  const handleBrowseLogPath = async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, log_path: path } as any) : null));
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  };

  const handleBrowseFFmpeg = async () => {
    try {
      const path = await ShowFFmpegDialog();
      if (path) {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        setSettings(s => (s ? ({ ...s, ffmpeg_path: path } as any) : null));
      }
    } catch (err) {
      console.error('Failed to browse for ffmpeg:', err);
    }
  };

  const handleReset = () => {
    if (originalSettings) {
      setSettings(JSON.parse(JSON.stringify(originalSettings)));
    }
  };

  const handleThemeChange = (value: string) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    setSettings(s => (s ? ({ ...s, theme: value } as any) : null));
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    setTheme(value as any);
    if (value === 'dark') {
      setColorScheme('dark');
    } else if (value === 'light') {
      setColorScheme('light');
    }
  };

  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const updateSetting = (key: string, value: any) => {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    setSettings(s => (s ? ({ ...s, [key]: value } as any) : null));
  };

  if (loading) {
    return <Text>Loading settings...</Text>;
  }

  if (!settings) {
    return <Text c="red">Failed to load settings</Text>;
  }

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Text c={dark ? '#fff' : '#000'} fw={700} size="xl">
          Settings
        </Text>
        {hasChanges && (
          <Badge color="yellow" variant="filled">
            Unsaved Changes
          </Badge>
        )}
      </Group>

      {saveError && (
        <Paper withBorder bg="#2c1b1b" p="sm" style={{ borderColor: '#c92a2a' }}>
          <Text c="red">{saveError}</Text>
        </Paper>
      )}

      {saveSuccess && (
        <Paper withBorder bg="#1b2c1b" p="sm" style={{ borderColor: '#2ac92a' }}>
          <Text c="green">Settings saved successfully!</Text>
        </Paper>
      )}

      {/* Downloads */}
      <Paper
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        p="md"
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
            Downloads
          </Text>

          <Group align="flex-end" gap="sm">
            <TextInput
              readOnly
              description="Where downloaded videos are saved"
              label="Download Path"
              style={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
              value={settings.download_path}
            />
            <Button
              color="yted"
              leftSection={<IconFolder size={16} />}
              variant="light"
              onClick={handleBrowseDownloadPath}
            >
              Browse
            </Button>
          </Group>

          <NumberInput
            description="Number of simultaneous downloads (1-10)"
            label="Max Concurrent Downloads"
            max={10}
            min={1}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.max_concurrent_downloads}
            w={200}
            onChange={v => updateSetting('max_concurrent_downloads', v || 1)}
          />

          <Select
            data={[
              { value: 'best', label: 'Best Quality' },
              { value: '1080p', label: '1080p' },
              { value: '720p', label: '720p' },
              { value: '480p', label: '480p' },
              { value: '360p', label: '360p' },
              { value: 'audio', label: 'Audio Only' },
            ]}
            description="Preferred quality for new downloads"
            label="Default Quality"
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.default_quality}
            w={200}
            onChange={v => v && updateSetting('default_quality', v)}
          />

          <TextInput
            description="Template for output filenames (yt-dlp format)"
            label="Filename Template"
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.filename_template}
            onChange={e => updateSetting('filename_template', e.currentTarget.value)}
          />
        </Stack>
      </Paper>

      {/* Download Presets */}
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

      {/* UI */}
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

      {/* Player */}
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

      {/* Logging */}
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

      {/* FFmpeg */}
      <Paper
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        p="md"
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
            FFmpeg Configuration
          </Text>
          <Text c={dark ? 'dimmed' : 'gray.6'} size="sm">
            FFmpeg is required for merging separate video and audio streams into a single file.
            Leave empty to auto-detect from system PATH.
          </Text>

          <Group align="flex-end" gap="sm">
            <TextInput
              description="Path to ffmpeg executable (optional)"
              label="FFmpeg Path"
              placeholder="Auto-detect from PATH"
              style={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
              value={settings.ffmpeg_path || ''}
              onChange={e => updateSetting('ffmpeg_path', e.currentTarget.value)}
            />
            <Button
              color="yted"
              leftSection={<IconFolder size={16} />}
              variant="light"
              onClick={handleBrowseFFmpeg}
            >
              Browse
            </Button>
          </Group>
        </Stack>
      </Paper>

      {/* Cache Management */}
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
                        error(
                          'Clear Failed',
                          'Failed to clear completed downloads: ' + err?.message
                        );
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

      {/* Actions */}
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
    </Stack>
  );
}

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
