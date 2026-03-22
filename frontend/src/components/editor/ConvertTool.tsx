import {
  Alert,
  Group,
  NumberInput,
  Paper,
  Select,
  Slider,
  Stack,
  Switch,
  Text,
} from '@mantine/core';
import { IconFile, IconInfoCircle, IconSettings } from '@tabler/icons-react';

import { EditSettings, OUTPUT_CODECS, OUTPUT_FORMATS } from '../../types/editor';

interface ConvertToolProps {
  settings: EditSettings;
  onChange: (settings: Partial<EditSettings>) => void;
}

export function ConvertTool({ settings, onChange }: ConvertToolProps) {
  const selectedFormat = OUTPUT_FORMATS.find(f => f.id === settings.outputFormat);
  const availableCodecs = selectedFormat?.codecs || [];

  return (
    <Stack gap="md">
      <Paper withBorder p="sm" bg="gray.0">
        <Stack gap="sm">
          <Text size="sm" fw={500}>
            <IconFile size={14} style={{ marginRight: 6 }} />
            Output Format
          </Text>
          <Select
            value={settings.outputFormat || 'mp4'}
            onChange={val => {
              onChange({ outputFormat: val as EditSettings['outputFormat'] });
              // Reset codec if not compatible
              const format = OUTPUT_FORMATS.find(f => f.id === val);
              if (format && format.codecs.length > 0) {
                onChange({ outputCodec: format.codecs[0] });
              }
            }}
            data={OUTPUT_FORMATS.map(f => ({
              value: f.id,
              label: `${f.name} - ${f.description}`,
            }))}
          />
        </Stack>
      </Paper>

      <Select
        label="Video Codec"
        value={settings.outputCodec || 'h264'}
        onChange={val => onChange({ outputCodec: val as EditSettings['outputCodec'] })}
        data={availableCodecs.map(codecId => {
          const codec = OUTPUT_CODECS.find(c => c.id === codecId);
          return {
            value: codecId,
            label: codec ? `${codec.name} (${codec.quality} quality, ${codec.speed})` : codecId,
          };
        })}
        disabled={availableCodecs.length === 0}
      />

      <Select
        label="Resolution"
        value={settings.outputResolution || 'original'}
        onChange={val => onChange({ outputResolution: val as EditSettings['outputResolution'] })}
        data={OUTPUT_RESOLUTIONS.map(r => ({
          value: r.id,
          label: r.name,
        }))}
      />

      <Paper withBorder p="sm">
        <Stack gap="xs">
          <Group justify="space-between">
            <Text size="sm" fw={500}>
              <IconSettings size={14} style={{ marginRight: 6 }} />
              Quality (CRF)
            </Text>
            <Text size="sm" c="dimmed">
              {settings.outputQuality ?? 23} - {getQualityLabel(settings.outputQuality ?? 23)}
            </Text>
          </Group>
          <Slider
            value={settings.outputQuality ?? 23}
            onChange={val => onChange({ outputQuality: val })}
            min={18}
            max={35}
            step={1}
            marks={[
              { value: 18, label: 'Best' },
              { value: 23, label: 'Good' },
              { value: 28, label: 'Med' },
              { value: 35, label: 'Low' },
            ]}
          />
          <Text size="xs" c="dimmed">
            Lower values = better quality, larger file size
          </Text>
        </Stack>
      </Paper>

      <Switch
        label="Remove Audio"
        checked={settings.removeAudio || false}
        onChange={e => onChange({ removeAudio: e.currentTarget.checked })}
      />

      {settings.outputFormat === 'gif' && (
        <Alert color="yellow" icon={<IconInfoCircle size={16} />}>
          <Text size="sm">
            GIF format will remove audio and limit quality. Best for short clips.
          </Text>
        </Alert>
      )}
    </Stack>
  );
}

function getQualityLabel(crf: number): string {
  if (crf <= 20) return 'High Quality';
  if (crf <= 23) return 'Good Quality';
  if (crf <= 28) return 'Medium Quality';
  return 'Smaller File';
}

// Add missing constant
const OUTPUT_RESOLUTIONS = [
  { id: 'original', name: 'Original Resolution' },
  { id: '2160p', name: '4K (2160p)' },
  { id: '1440p', name: '2K (1440p)' },
  { id: '1080p', name: '1080p Full HD' },
  { id: '720p', name: '720p HD' },
  { id: '480p', name: '480p SD' },
  { id: '360p', name: '360p' },
] as const;
