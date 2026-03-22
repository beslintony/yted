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
import { IconLetterT, IconPhoto, IconTypography } from '@tabler/icons-react';

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
          <Radio value="text" label="Text Watermark" />
          <Radio value="image" label="Image Watermark" />
        </Group>
      </Radio.Group>

      {settings.watermarkType === 'text' ? (
        <Paper withBorder p="sm" >
          <Stack gap="sm">
            <Text size="sm" fw={500}>
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
              value={settings.watermarkSize || 24}
              onChange={val => onChange({ watermarkSize: Number(val) || 24 })}
              min={8}
              max={120}
            />
          </Stack>
        </Paper>
      ) : (
        <Paper withBorder p="sm" >
          <Stack gap="sm">
            <Text size="sm" fw={500}>
              <IconPhoto size={14} style={{ marginRight: 6 }} />
              Image Settings
            </Text>
            <FileInput
              label="Watermark Image"
              placeholder="Select PNG image with transparency"
              accept="image/png"
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
              value={settings.watermarkSize || 100}
              onChange={val => onChange({ watermarkSize: Number(val) || 100 })}
              min={10}
              max={500}
            />
          </Stack>
        </Paper>
      )}

      <Select
        label="Position"
        value={settings.watermarkPosition || 'bottom-right'}
        onChange={val => onChange({ watermarkPosition: val as EditSettings['watermarkPosition'] })}
        data={WATERMARK_POSITIONS}
      />

      <Stack gap="xs">
        <Group justify="space-between">
          <Text size="sm">Opacity</Text>
          <Text size="sm" c="dimmed">
            {Math.round((settings.watermarkOpacity ?? 0.7) * 100)}%
          </Text>
        </Group>
        <Slider
          value={(settings.watermarkOpacity ?? 0.7) * 100}
          onChange={val => onChange({ watermarkOpacity: val / 100 })}
          min={10}
          max={100}
        />
      </Stack>
    </Stack>
  );
}
