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
  const [settings, setSettings] = useState<config.Config | null>(null);
  const [loading, setLoading] = useState(true);
  const [presetModalOpen, setPresetModalOpen] = useState(false);
  const [editingPreset, setEditingPreset] = useState<config.DownloadPreset | null>(null);

  const {
    theme,
    setTheme,
    toggleSidebar,
    downloadPath,
    setDownloadPath,
    maxConcurrentDownloads,
    setMaxConcurrentDownloads,
    defaultQuality,
    setDefaultQuality,
    downloadPresets,
    addDownloadPreset,
    removeDownloadPreset,
    updateDownloadPreset,
  } = useSettingsStore();

  useEffect(() => {
    loadSettings();
  }, []);

  const loadSettings = async () => {
    try {
      const result = await GetSettings();
      setSettings(result);
      if (result) {
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
    
    try {
      await SaveSettings(settings);
    } catch (err) {
      console.error('Failed to save settings:', err);
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
    loadSettings();
  };

  if (loading) {
    return <Text>Loading settings...</Text>;
  }

  if (!settings) {
    return <Text c="red">Failed to load settings</Text>;
  }

  return (
    <Stack spacing="lg">
      <Text size="xl" fw={700}>Settings</Text>

      {/* Downloads */}
      <Paper p="md" withBorder>
        <Stack spacing="md">
          <Text size="lg" fw={600}>Downloads</Text>
          
          <Group align="flex-end">
            <TextInput
              label="Download Path"
              description="Where downloaded videos are saved"
              value={settings.download_path}
              readOnly
              style={{ flex: 1 }}
            />
            <Button
              variant="light"
              leftIcon={<IconFolder size={16} />}
              onClick={handleBrowseDownloadPath}
            >
              Browse
            </Button>
          </Group>

          <NumberInput
            label="Max Concurrent Downloads"
            description="Number of simultaneous downloads (1-10)"
            value={settings.max_concurrent_downloads}
            onChange={(v) => setSettings((s) => s ? { ...s, max_concurrent_downloads: v || 1 } as any : null)}
            min={1}
            max={10}
            w={200}
          />

          <Select
            label="Default Quality"
            description="Preferred quality for new downloads"
            value={settings.default_quality}
            onChange={(v) => v && setSettings((s) => s ? { ...s, default_quality: v } as any : null)}
            data={[
              { value: 'best', label: 'Best Quality' },
              { value: '1080p', label: '1080p' },
              { value: '720p', label: '720p' },
              { value: '480p', label: '480p' },
              { value: '360p', label: '360p' },
              { value: 'audio', label: 'Audio Only' },
            ]}
            w={200}
          />

          <TextInput
            label="Filename Template"
            description="Template for output filenames (yt-dlp format)"
            value={settings.filename_template}
            onChange={(e) => setSettings((s) => s ? { ...s, filename_template: e.currentTarget.value } as any : null)}
          />
        </Stack>
      </Paper>

      {/* Download Presets */}
      <Paper p="md" withBorder>
        <Stack spacing="md">
          <Group position="apart">
            <Text size="lg" fw={600}>Download Presets</Text>
            <Button
              size="sm"
              leftIcon={<IconPlus size={16} />}
              onClick={() => {
                setEditingPreset(null);
                setPresetModalOpen(true);
              }}
            >
              Add Preset
            </Button>
          </Group>

          <Table>
            <thead>
              <tr>
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
                  <td>{preset.name}</td>
                  <td><code>{preset.format}</code></td>
                  <td><Badge>{preset.quality}</Badge></td>
                  <td>{preset.extension}</td>
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
                          onClick={() => removeDownloadPreset(preset.id)}
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
      <Paper p="md" withBorder>
        <Stack spacing="md">
          <Text size="lg" fw={600}>UI</Text>

          <Select
            label="Theme"
            value={settings.theme}
            onChange={(v) => v && setSettings((s) => s ? { ...s, theme: v } as any : null)}
            data={[
              { value: 'dark', label: 'Dark' },
              { value: 'light', label: 'Light' },
              { value: 'auto', label: 'Auto' },
            ]}
            w={200}
          />

          <ColorInput
            label="Accent Color"
            value={settings.accent_color}
            onChange={(v) => setSettings((s) => s ? { ...s, accent_color: v } as any : null)}
            w={200}
          />

          <Switch
            label="Collapse Sidebar"
            checked={settings.sidebar_collapsed}
            onChange={(e) => {
              setSettings((s) => s ? { ...s, sidebar_collapsed: e.currentTarget.checked } as any : null);
              toggleSidebar();
            }}
          />
        </Stack>
      </Paper>

      {/* Player */}
      <Paper p="md" withBorder>
        <Stack spacing="md">
          <Text size="lg" fw={600}>Player</Text>

          <NumberInput
            label="Default Volume"
            value={settings.default_volume}
            onChange={(v) => setSettings((s) => s ? { ...s, default_volume: v || 80 } as any : null)}
            min={0}
            max={100}
            w={200}
          />

          <Switch
            label="Remember Watch Position"
            checked={settings.remember_position}
            onChange={(e) => setSettings((s) => s ? { ...s, remember_position: e.currentTarget.checked } as any : null)}
          />
        </Stack>
      </Paper>

      {/* Actions */}
      <Group position="right">
        <Button variant="light" leftIcon={<IconRefresh size={16} />} onClick={handleReset}>
          Reset
        </Button>
        <Button leftIcon={<IconDeviceFloppy size={16} />} onClick={handleSave}>
          Save Settings
        </Button>
      </Group>

      {/* Preset Modal */}
      <PresetModal
        opened={presetModalOpen}
        onClose={() => setPresetModalOpen(false)}
        preset={editingPreset}
        onSave={(preset) => {
          const fixedPreset = { ...preset, quality: preset.quality as any };
          if (editingPreset) {
            updateDownloadPreset(editingPreset.id, fixedPreset);
          } else {
            addDownloadPreset(fixedPreset as any);
          }
          setPresetModalOpen(false);
        }}
      />
    </Stack>
  );
}

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
          onChange={(v) => v && setQuality(v as any)}
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
            })}
            disabled={!name || !format}
          >
            Save
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
