import { useEffect, useState } from 'react';
import {
  Stack,
  Paper,
  Text,
  TextInput,
  NumberInput,
  Select,
  Group,
  Button,
  Switch,
  ActionIcon,
  Tooltip,
  Table,
  Modal,
  Badge,
  useMantineColorScheme,
  ColorInput,
} from '@mantine/core';
import {
  IconFolder,
  IconDeviceFloppy,
  IconRefresh,
  IconPlus,
  IconTrash,
  IconEdit,
} from '@tabler/icons-react';
import { useSettingsStore } from '../stores';
import { GetSettings, SaveSettings, ShowOpenDirectoryDialog } from '../../wailsjs/go/app/App';
import { config } from '../../wailsjs/go/models';

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

  const dark = colorScheme === 'dark';

  useEffect(() => {
    loadSettings();
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
        setDefaultQuality(result.default_quality as any);
        setTheme(result.theme as any);
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
      
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      console.error('Failed to save settings:', err);
      setSaveError(err?.message || 'Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const handleBrowseDownloadPath = async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        setSettings((s) => s ? { ...s, download_path: path } as any : null);
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
        setSettings((s) => s ? { ...s, log_export_path: path } as any : null);
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  };

  const handleBrowseLogPath = async () => {
    try {
      const path = await ShowOpenDirectoryDialog();
      if (path) {
        setSettings((s) => s ? { ...s, log_path: path } as any : null);
      }
    } catch (err) {
      console.error('Failed to browse:', err);
    }
  };

  const handleReset = () => {
    if (originalSettings) {
      setSettings(JSON.parse(JSON.stringify(originalSettings)));
    }
  };

  const handleThemeChange = (value: string) => {
    setSettings((s) => s ? { ...s, theme: value } as any : null);
    setTheme(value as any);
    if (value === 'dark') {
      setColorScheme('dark');
    } else if (value === 'light') {
      setColorScheme('light');
    }
  };

  const updateSetting = (key: string, value: any) => {
    setSettings((s) => s ? { ...s, [key]: value } as any : null);
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
        <Text size="xl" fw={700} c={dark ? '#fff' : '#000'}>Settings</Text>
        {hasChanges && (
          <Badge color="yellow" variant="filled">Unsaved Changes</Badge>
        )}
      </Group>

      {saveError && (
        <Paper p="sm" withBorder bg="#2c1b1b" style={{ borderColor: '#c92a2a' }}>
          <Text c="red">{saveError}</Text>
        </Paper>
      )}

      {saveSuccess && (
        <Paper p="sm" withBorder bg="#1b2c1b" style={{ borderColor: '#2ac92a' }}>
          <Text c="green">Settings saved successfully!</Text>
        </Paper>
      )}

      {/* Downloads */}
      <Paper 
        p="md" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Downloads</Text>
          
          <Group align="flex-end" gap="sm">
            <TextInput
              label="Download Path"
              description="Where downloaded videos are saved"
              value={settings.download_path}
              readOnly
              style={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
            />
            <Button
              variant="light"
              leftSection={<IconFolder size={16} />}
              onClick={handleBrowseDownloadPath}
              color="yted"
            >
              Browse
            </Button>
          </Group>

          <NumberInput
            label="Max Concurrent Downloads"
            description="Number of simultaneous downloads (1-10)"
            value={settings.max_concurrent_downloads}
            onChange={(v) => updateSetting('max_concurrent_downloads', v || 1)}
            min={1}
            max={10}
            w={200}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />

          <Select
            label="Default Quality"
            description="Preferred quality for new downloads"
            value={settings.default_quality}
            onChange={(v) => v && updateSetting('default_quality', v)}
            data={[
              { value: 'best', label: 'Best Quality' },
              { value: '1080p', label: '1080p' },
              { value: '720p', label: '720p' },
              { value: '480p', label: '480p' },
              { value: '360p', label: '360p' },
              { value: 'audio', label: 'Audio Only' },
            ]}
            w={200}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />

          <TextInput
            label="Filename Template"
            description="Template for output filenames (yt-dlp format)"
            value={settings.filename_template}
            onChange={(e) => updateSetting('filename_template', e.currentTarget.value)}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />
        </Stack>
      </Paper>

      {/* Download Presets */}
      <Paper 
        p="md" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Group justify="space-between">
            <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Download Presets</Text>
            <Button
              size="sm"
              leftSection={<IconPlus size={16} />}
              onClick={() => {
                setEditingPreset(null);
                setPresetModalOpen(true);
              }}
              color="yted"
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
              {settings.download_presets?.map((preset) => (
                <Table.Tr key={preset.id}>
                  <Table.Td c={dark ? '#fff' : '#000'}>{preset.name}</Table.Td>
                  <Table.Td>
                    <code style={{ color: dark ? '#c1c2c5' : '#495057' }}>{preset.format}</code>
                  </Table.Td>
                  <Table.Td><Badge color="yted">{preset.quality}</Badge></Table.Td>
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
                          size="sm"
                          color="red"
                          onClick={() => {
                            const newPresets = settings.download_presets.filter((p) => p.id !== preset.id);
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
        p="md" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>UI</Text>

          <Select
            label="Theme"
            value={settings.theme}
            onChange={(v) => v && handleThemeChange(v)}
            data={[
              { value: 'dark', label: 'Dark' },
              { value: 'light', label: 'Light' },
              { value: 'auto', label: 'Auto' },
            ]}
            w={200}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />

          <ColorInput
            label="Accent Color"
            value={settings.accent_color}
            onChange={(v) => updateSetting('accent_color', v)}
            w={200}
          />

          <Switch
            label="Collapse Sidebar"
            checked={settings.sidebar_collapsed}
            onChange={(e) => {
              updateSetting('sidebar_collapsed', e.currentTarget.checked);
              toggleSidebar();
            }}
            styles={{
              label: {
                color: dark ? '#c1c2c5' : '#495057',
              },
            }}
          />
        </Stack>
      </Paper>

      {/* Player */}
      <Paper 
        p="md" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Player</Text>

          <NumberInput
            label="Default Volume"
            value={settings.default_volume}
            onChange={(v) => updateSetting('default_volume', v || 80)}
            min={0}
            max={100}
            w={200}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />

          <Switch
            label="Remember Watch Position"
            checked={settings.remember_position}
            onChange={(e) => updateSetting('remember_position', e.currentTarget.checked)}
            styles={{
              label: {
                color: dark ? '#c1c2c5' : '#495057',
              },
            }}
          />
        </Stack>
      </Paper>

      {/* Logging */}
      <Paper 
        p="md" 
        withBorder
        bg={dark ? '#25262b' : '#fff'}
        style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
      >
        <Stack gap="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Logging</Text>
          
          <Group align="flex-end" gap="sm">
            <TextInput
              label="Log Storage Path"
              description="Where application logs are stored (default: ~/.yted/.logs)"
              value={settings.log_path || ''}
              readOnly
              style={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
            />
            <Button
              variant="light"
              leftSection={<IconFolder size={16} />}
              onClick={handleBrowseLogPath}
              color="yted"
            >
              Browse
            </Button>
          </Group>

          <NumberInput
            label="Max Log Sessions"
            description="Number of log sessions to keep (1-100). Each session is one app start."
            value={settings.max_log_sessions || 10}
            onChange={(v) => updateSetting('max_log_sessions', v || 10)}
            min={1}
            max={100}
            w={200}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
          />
          
          <Group align="flex-end" gap="sm">
            <TextInput
              label="Log Export Path"
              description="Where exported logs are saved"
              value={settings.log_export_path || ''}
              readOnly
              style={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
            />
            <Button
              variant="light"
              leftSection={<IconFolder size={16} />}
              onClick={handleBrowseLogExportPath}
              color="yted"
            >
              Browse
            </Button>
          </Group>
        </Stack>
      </Paper>

      {/* Actions */}
      <Group justify="flex-end">
        {hasChanges && (
          <Button 
            variant="light" 
            leftSection={<IconRefresh size={16} />} 
            onClick={handleReset}
            color="gray"
          >
            Reset
          </Button>
        )}
        <Button 
          leftSection={<IconDeviceFloppy size={16} />} 
          onClick={handleSave}
          loading={saving}
          disabled={!hasChanges}
          color="yted"
        >
          Save Settings
        </Button>
      </Group>

      {/* Preset Modal */}
      <Modal 
        opened={presetModalOpen} 
        onClose={() => setPresetModalOpen(false)} 
        title={editingPreset ? 'Edit Preset' : 'Add Preset'}
      >
        <PresetForm
          preset={editingPreset}
          onSave={(preset) => {
            const currentPresets = settings.download_presets || [];
            if (editingPreset) {
              const updatedPresets = currentPresets.map((p) =>
                p.id === preset.id ? preset : p
              );
              updateSetting('download_presets', updatedPresets);
            } else {
              const newPreset = { ...preset, id: crypto.randomUUID() };
              updateSetting('download_presets', [...currentPresets, newPreset]);
            }
            setPresetModalOpen(false);
          }}
          onCancel={() => setPresetModalOpen(false)}
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
        value={name}
        onChange={(e) => setName(e.currentTarget.value)}
        placeholder="e.g., 1080p Video"
      />
      <TextInput
        label="Format"
        value={format}
        onChange={(e) => setFormat(e.currentTarget.value)}
        placeholder="e.g., bestvideo[height<=1080]+bestaudio"
      />
      <Select
        label="Quality"
        value={quality}
        onChange={(v) => v && setQuality(v)}
        data={[
          { value: 'best', label: 'Best' },
          { value: '1080p', label: '1080p' },
          { value: '720p', label: '720p' },
          { value: '480p', label: '480p' },
          { value: 'audio', label: 'Audio' },
        ]}
      />
      <TextInput
        label="Extension"
        value={extension}
        onChange={(e) => setExtension(e.currentTarget.value)}
        placeholder="e.g., mp4"
      />
      <Group justify="flex-end" mt="md">
        <Button variant="light" onClick={onCancel}>Cancel</Button>
        <Button
          onClick={() => onSave({
            id: preset?.id || '',
            name,
            format,
            quality,
            extension,
          } as any)}
          disabled={!name || !format}
          color="yted"
        >
          Save
        </Button>
      </Group>
    </Stack>
  );
}
