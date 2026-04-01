import {
  FileInput,
  Group,
  NumberInput,
  Paper,
  Radio,
  Select,
  Slider,
  Stack,
  Text,
  TextInput,
} from '@mantine/core';
import { IconPhoto, IconTypography } from '@tabler/icons-react';

import { EditSettings, WATERMARK_POSITIONS } from '../../types/editor';

interface WatermarkToolProps {
  settings: EditSettings;
  onChange: (settings: Partial<EditSettings>) => void;
}

export function WatermarkTool({ settings, onChange }: WatermarkToolProps) {
  return (
    <Stack gap="md">
      <Radio.Group
        value={settings.watermarkType || 'text'}
        onChange={val => onChange({ watermarkType: val as 'text' | 'image' })}
      >
        <Group>
          <Radio label="Text Watermark" value="text" />
          <Radio label="Image Watermark" value="image" />
        </Group>
      </Radio.Group>

      {settings.watermarkType === 'text' ? (
        <Paper withBorder bg="gray.0" p="sm">
          <Stack gap="sm">
            <Text fw={500} size="sm">
              <IconTypography size={14} style={{ marginRight: 6 }} />
              Text Settings
            </Text>
            <TextInput
              label="Watermark Text"
              placeholder="Enter watermark text"
              value={settings.watermarkText || ''}
              onChange={e => onChange({ watermarkText: e.currentTarget.value })}
            />
            <NumberInput
              label="Font Size"
              max={120}
              min={8}
              value={settings.watermarkSize || 24}
              onChange={val => onChange({ watermarkSize: Number(val) || 24 })}
            />
          </Stack>
        </Paper>
      ) : (
        <Paper withBorder bg="gray.0" p="sm">
          <Stack gap="sm">
            <Text fw={500} size="sm">
              <IconPhoto size={14} style={{ marginRight: 6 }} />
              Image Settings
            </Text>
            <FileInput
              accept="image/png"
              label="Watermark Image"
              placeholder="Select PNG image with transparency"
              value={settings.watermarkImage ? new File([], settings.watermarkImage) : null}
              onChange={file => {
                if (file) {
                  // Note: In a real app, you'd need to handle file path properly
                  // This is a simplified version
                  onChange({ watermarkImage: file.name });
                }
              }}
            />
            <NumberInput
              label="Scale (%)"
              max={500}
              min={10}
              value={settings.watermarkSize || 100}
              onChange={val => onChange({ watermarkSize: Number(val) || 100 })}
            />
          </Stack>
        </Paper>
      )}

      <Select
        data={WATERMARK_POSITIONS}
        label="Position"
        value={settings.watermarkPosition || 'bottom-right'}
        onChange={val => onChange({ watermarkPosition: val as EditSettings['watermarkPosition'] })}
      />

      <Stack gap="xs">
        <Group justify="space-between">
          <Text size="sm">Opacity</Text>
          <Text c="dimmed" size="sm">
            {Math.round((settings.watermarkOpacity ?? 0.7) * 100)}%
          </Text>
        </Group>
        <Slider
          max={100}
          min={10}
          value={(settings.watermarkOpacity ?? 0.7) * 100}
          onChange={val => onChange({ watermarkOpacity: val / 100 })}
        />
      </Stack>
    </Stack>
  );
}
