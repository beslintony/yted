import { Group, NumberInput, Paper, RangeSlider, Select, Stack, Text } from '@mantine/core';
import { IconAspectRatio, IconClock, IconCrop } from '@tabler/icons-react';

import { CROP_PRESETS, EditSettings } from '../../types/editor';

interface CropToolProps {
  settings: EditSettings;
  onChange: (settings: Partial<EditSettings>) => void;
  videoDuration: number;
  videoWidth: number;
  videoHeight: number;
}

export function CropTool({
  settings,
  onChange,
  videoDuration,
  videoWidth,
  videoHeight,
}: CropToolProps) {
  const handleCropPresetChange = (presetId: string | null) => {
    if (!presetId || !videoWidth || !videoHeight) return;

    const preset = CROP_PRESETS.find(p => p.id === presetId);
    if (!preset || preset.id === 'free') {
      onChange({ cropX: 0, cropY: 0, cropWidth: videoWidth, cropHeight: videoHeight });
      return;
    }

    // Calculate crop region based on aspect ratio
    const targetRatio = preset.width / preset.height;
    const currentRatio = videoWidth / videoHeight;

    let cropX = 0;
    let cropY = 0;
    let cropWidth = videoWidth;
    let cropHeight = videoHeight;

    if (currentRatio > targetRatio) {
      // Video is wider - crop width
      cropWidth = Math.round(videoHeight * targetRatio);
      cropX = Math.round((videoWidth - cropWidth) / 2);
    } else {
      // Video is taller - crop height
      cropHeight = Math.round(videoWidth / targetRatio);
      cropY = Math.round((videoHeight - cropHeight) / 2);
    }

    onChange({ cropX, cropY, cropWidth, cropHeight });
  };

  const formatTime = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <Stack gap="md">
      <Text fw={600}>
        <IconClock size={16} style={{ marginRight: 8 }} />
        Time Range
      </Text>

      {videoDuration > 0 && (
        <>
          <RangeSlider
            marks={[
              { value: 0, label: '0:00' },
              { value: videoDuration / 2, label: formatTime(videoDuration / 2) },
              { value: videoDuration, label: formatTime(videoDuration) },
            ]}
            max={videoDuration}
            min={0}
            step={0.1}
            value={[settings.cropStart ?? 0, settings.cropEnd ?? videoDuration]}
            onChange={([start, end]) => {
              onChange({ cropStart: start, cropEnd: end });
            }}
          />
          <Group justify="space-between">
            <Text c="dimmed" size="sm">
              Start: {formatTime(settings.cropStart ?? 0)}
            </Text>
            <Text c="dimmed" size="sm">
              End: {formatTime(settings.cropEnd ?? videoDuration)}
            </Text>
          </Group>
        </>
      )}

      <Paper withBorder bg="gray.0" p="sm">
        <Stack gap="xs">
          <Text fw={500} size="sm">
            <IconAspectRatio size={14} style={{ marginRight: 6 }} />
            Aspect Ratio Preset
          </Text>
          <Select
            data={CROP_PRESETS.map(p => ({
              value: p.id,
              label: `${p.name} (${p.ratio})`,
            }))}
            placeholder="Select crop ratio"
            onChange={handleCropPresetChange}
          />
        </Stack>
      </Paper>

      <Text fw={600}>
        <IconCrop size={16} style={{ marginRight: 8 }} />
        Spatial Crop (pixels)
      </Text>

      <Group grow>
        <NumberInput
          label="X Position"
          max={videoWidth}
          min={0}
          value={settings.cropX ?? 0}
          onChange={val => onChange({ cropX: Number(val) || 0 })}
        />
        <NumberInput
          label="Y Position"
          max={videoHeight}
          min={0}
          value={settings.cropY ?? 0}
          onChange={val => onChange({ cropY: Number(val) || 0 })}
        />
      </Group>

      <Group grow>
        <NumberInput
          label="Width"
          max={videoWidth}
          min={1}
          value={settings.cropWidth ?? videoWidth}
          onChange={val => onChange({ cropWidth: Number(val) || videoWidth })}
        />
        <NumberInput
          label="Height"
          max={videoHeight}
          min={1}
          value={settings.cropHeight ?? videoHeight}
          onChange={val => onChange({ cropHeight: Number(val) || videoHeight })}
        />
      </Group>

      {videoWidth > 0 && videoHeight > 0 && (
        <Paper withBorder p="sm">
          <Text c="dimmed" size="sm">
            Output: {settings.cropWidth ?? videoWidth} x {settings.cropHeight ?? videoHeight} @{' '}
            {(settings.cropStart ?? 0).toFixed(1)}s -{' '}
            {(settings.cropEnd ?? videoDuration).toFixed(1)}s
          </Text>
        </Paper>
      )}
    </Stack>
  );
}
