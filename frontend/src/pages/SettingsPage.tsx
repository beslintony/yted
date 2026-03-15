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
  Divider,
  Switch,
  ColorInput,
  ActionIcon,
  Tooltip,
  Table,
  Modal,
  Badge,
  useMantineColorScheme,
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
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
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
    theme: userTheme,
    setTheme,
    toggleSidebar,
    setDownloadPath,
    setMaxConcurrentDownloads,
    setDefaultQuality,
    downloadPresets,
    addDownloadPreset,
    removeDownloadPreset,
    updateDownloadPreset,
    saveSettings: saveSettingsToStore,
  } = useSettingsStore();

  const dark = colorScheme === 'dark';

  useEffect(() => {
    loadSettings();
  }, []);

  // Track changes
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
        
        // Sync with store
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
      
      // Also save to store
      await saveSettingsToStore();
      
      // Apply theme if changed
      if (settings.theme !== userTheme) {
        setTheme(settings.theme as any);
        if (settings.theme === 'dark' && colorScheme !== 'dark') {
          toggleColorScheme('dark');
        } else if (settings.theme === 'light' && colorScheme !== 'light') {
          toggleColorScheme('light');
        }
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

  const handleReset = () => {
    if (originalSettings) {
      setSettings(JSON.parse(JSON.stringify(originalSettings)));
    }
  };

  const handleThemeChange = (value: string) => {
    setSettings((s) => s ? { ...s, theme: value } as any : null);
    setTheme(value as any);
    if (value === 'dark' && colorScheme !== 'dark') {
      toggleColorScheme('dark');
    } else if (value === 'light' && colorScheme !== 'light') {
      toggleColorScheme('light');
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
    <Stack spacing="lg">
      <Group position="apart">
        <Text size="xl" fw={700} c={dark ? '#fff' : '#000'}>Settings</Text>
        {hasChanges && (
          <Badge color="yellow" variant="filled">Unsaved Changes</Badge>
        )}
      </Group>

      {saveError && (
        <Paper p="sm" withBorder sx={{ background: '#2c1b1b', borderColor: '#c92a2a' }}>
          <Text c="red">{saveError}</Text>
        </Paper>
      )}

      {saveSuccess && (
        <Paper p="sm" withBorder sx={{ background: '#1b2c1b', borderColor: '#2ac92a' }}>
          <Text c="green">Settings saved successfully!</Text>
        </Paper>
      )}

      {/* Downloads */}
      <Paper 
        p="md" 
        withBorder
        sx={{
          background: dark ? '#25262b' : '#fff',
          borderColor: dark ? '#373a40' : '#dee2e6',
        }}
      >
        <Stack spacing="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Downloads</Text>
          
          <Group align="flex-end">
            <TextInput
              label="Download Path"
              description="Where downloaded videos are saved"
              value={settings.download_path}
              readOnly
              sx={{ flex: 1 }}
              styles={{
                input: {
                  background: dark ? '#1a1b1e' : '#f8f9fa',
                  color: dark ? '#c1c2c5' : '#212529',
                },
              }}
            />
            <Button
              variant="light"
              leftIcon={<IconFolder size={16} />}
              onClick={handleBrowseDownloadPath}
              color="red"
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
        sx={{
          background: dark ? '#25262b' : '#fff',
          borderColor: dark ? '#373a40' : '#dee2e6',
        }}
      >
        <Stack spacing="md">
          <Group position="apart">
            <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>Download Presets</Text>
            <Button
              size="sm"
              leftIcon={<IconPlus size={16} />}
              onClick={() => {
                setEditingPreset(null);
                setPresetModalOpen(true);
              }}
              color="red"
            >
              Add Preset
            </Button>
          </Group>

          <Table>
            <thead>
              <tr style={{ color: dark ? '#c1c2c5' : '#495057' }}>
                <th>Name</th>
                <th>Format</th>
                <th>Quality</th>
                <th>Extension</th>
                <th>Actions</th>
              </tr>
            </thead>
            <tbody>
              {settings.download_presets?.map((preset) => (
                <tr key={preset.id}>
                  <td style={{ color: dark ? '#fff' : '#000' }}>{preset.name}</td>
                  <td><code style={{ color: dark ? '#c1c2c5' : '#495057' }}>{preset.format}</code></td>
                  <td><Badge color="red">{preset.quality}</Badge></td>
                  <td style={{ color: dark ? '#c1c2c5' : '#495057' }}>{preset.extension}</td>
                  <td>
                    <Group spacing={4}>
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
                  </td>
                </tr>
              ))}
            </tbody>
          </Table>
        </Stack>
      </Paper>

      {/* UI */}
      <Paper 
        p="md" 
        withBorder
        sx={{
          background: dark ? '#25262b' : '#fff',
          borderColor: dark ? '#373a40' : '#dee2e6',
        }}
      >
        <Stack spacing="md">
          <Text size="lg" fw={600} c={dark ? '#fff' : '#000'}>UI</Text>

          <Select
            label="Theme"
            value={settings.theme}
            onChange={handleThemeChange}
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
        sx={{
          background: dark ? '#25262b' : '#fff',
          borderColor: dark ? '#373a40' : '#dee2e6',
        }}
      >
        <Stack spacing="md">
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

      {/* Actions */}
      <Group position="right">
        {hasChanges && (
          <Button 
            variant="light" 
            leftIcon={<IconRefresh size={16} />} 
            onClick={handleReset}
            color="gray"
          >
            Reset
          </Button>
        )}
        <Button 
          leftIcon={<IconDeviceFloppy size={16} />} 
          onClick={handleSave}
          loading={saving}
          disabled={!hasChanges}
          color="red"
        >
          Save Settings
        </Button>
      </Group>

      {/* Preset Modal */}
      <PresetModal
        opened={presetModalOpen}
        onClose={() => setPresetModalOpen(false)}
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
      />
    </Stack>
  );

  function PresetModal({
    opened,
    onClose,
    preset,
    onSave,
  }: {
    opened: boolean;
    onClose: () => void;
    preset: config.DownloadPreset | null;
    onSave: (preset: config.DownloadPreset) => void;
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
      <Modal opened={opened} onClose={onClose} title={preset ? 'Edit Preset' : 'Add Preset'}>
        <Stack spacing="md">
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
          <Group position="right" mt="md">
            <Button variant="light" onClick={onClose}>Cancel</Button>
            <Button
              onClick={() => onSave({
                id: preset?.id || '',
                name,
                format,
                quality,
                extension,
              } as any)}
              disabled={!name || !format}
              color="red"
            >
              Save
            </Button>
          </Group>
        </Stack>
      </Modal>
    );
  }
}
