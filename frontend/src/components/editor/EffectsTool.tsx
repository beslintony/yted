import {
  Button,
  Divider,
  Group,
  Paper,
  SegmentedControl,
  Slider,
  Stack,
  Switch,
  Text,
} from '@mantine/core';
import {
  IconBrightness,
  IconContrast,
  IconMoon,
  IconRefresh,
  IconRotate,
  IconSpeedboat,
  IconSun,
  IconVolume,
} from '@tabler/icons-react';
import React from 'react';

import { EditSettings, EFFECT_RANGES, ROTATION_OPTIONS } from '../../types/editor';

interface EffectsToolProps {
  settings: EditSettings;
  onChange: (settings: Partial<EditSettings>) => void;
}

export function EffectsTool({ settings, onChange }: EffectsToolProps) {
  const handleReset = () => {
    onChange({
      brightness: 0,
      contrast: 0,
      saturation: 1,
      rotation: 0,
      speed: 1,
      volume: 1,
      removeAudio: false,
    });
  };

  const renderSlider = (
    label: string,
    icon: React.ReactNode,
    value: number,
    range: { min: number; max: number; default: number; step: number },
    onChangeValue: (val: number) => void,
    formatValue?: (val: number) => string
  ) => (
    <Stack gap="xs">
      <Group justify="space-between">
        <Group gap="xs">
          {icon}
          <Text fw={500} size="sm">
            {label}
          </Text>
        </Group>
        <Text c="dimmed" size="sm">
          {formatValue ? formatValue(value) : value.toFixed(1)}
        </Text>
      </Group>
      <Slider
        marks={[
          { value: range.min, label: range.min.toString() },
          { value: range.default, label: range.default.toString() },
          { value: range.max, label: range.max.toString() },
        ]}
        max={range.max}
        min={range.min}
        step={range.step}
        value={value}
        onChange={onChangeValue}
      />
    </Stack>
  );

  return (
    <Stack gap="lg">
      <Group justify="space-between">
        <Text fw={600}>Video Effects</Text>
        <Button
          leftSection={<IconRefresh size={14} />}
          size="xs"
          variant="light"
          onClick={handleReset}
        >
          Reset All
        </Button>
      </Group>

      <Paper withBorder bg="gray.0" p="sm">
        <Stack gap="md">
          <Text fw={500} size="sm">
            <IconBrightness size={14} style={{ marginRight: 6 }} />
            Color Adjustments
          </Text>

          {renderSlider(
            'Brightness',
            <IconSun size={16} />,
            settings.brightness ?? 0,
            EFFECT_RANGES.brightness,
            val => onChange({ brightness: val }),
            val => (val > 0 ? `+${val.toFixed(1)}` : val.toFixed(1))
          )}

          {renderSlider(
            'Contrast',
            <IconContrast size={16} />,
            settings.contrast ?? 0,
            EFFECT_RANGES.contrast,
            val => onChange({ contrast: val }),
            val => (val > 0 ? `+${val.toFixed(1)}` : val.toFixed(1))
          )}

          {renderSlider(
            'Saturation',
            <IconMoon size={16} />,
            settings.saturation ?? 1,
            EFFECT_RANGES.saturation,
            val => onChange({ saturation: val })
          )}
        </Stack>
      </Paper>

      <Paper withBorder bg="gray.0" p="sm">
        <Stack gap="md">
          <Text fw={500} size="sm">
            <IconRotate size={14} style={{ marginRight: 6 }} />
            Transform
          </Text>

          <Stack gap="xs">
            <Text size="sm">Rotation</Text>
            <SegmentedControl
              data={ROTATION_OPTIONS.map(opt => ({
                value: opt.value.toString(),
                label: opt.label,
              }))}
              value={(settings.rotation ?? 0).toString()}
              onChange={val => onChange({ rotation: Number(val) as EditSettings['rotation'] })}
            />
          </Stack>

          {renderSlider(
            'Speed',
            <IconSpeedboat size={16} />,
            settings.speed ?? 1,
            EFFECT_RANGES.speed,
            val => onChange({ speed: val }),
            val => `${val.toFixed(1)}x`
          )}
        </Stack>
      </Paper>

      <Paper withBorder bg="gray.0" p="sm">
        <Stack gap="md">
          <Text fw={500} size="sm">
            <IconVolume size={14} style={{ marginRight: 6 }} />
            Audio
          </Text>

          {renderSlider(
            'Volume',
            <IconVolume size={16} />,
            settings.volume ?? 1,
            EFFECT_RANGES.volume,
            val => onChange({ volume: val }),
            val => `${Math.round(val * 100)}%`
          )}

          <Divider />

          <Switch
            checked={settings.removeAudio || false}
            description="Create a silent video"
            label="Remove Audio Track"
            onChange={e => onChange({ removeAudio: e.currentTarget.checked })}
          />
        </Stack>
      </Paper>
    </Stack>
  );
}
