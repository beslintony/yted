import {
  Alert,
  Button,
  Checkbox,
  Group,
  Image,
  Loader,
  Modal,
  NumberInput,
  Progress,
  SegmentedControl,
  Select,
  Slider,
  Stack,
  Tabs,
  Text,
  TextInput,
} from '@mantine/core';
import { IconAlertCircle } from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import { GenerateEditPreview, GetEditVideoMetadata } from '../../wailsjs/go/app/App';
import { app, db, editor } from '../../wailsjs/go/models';
import { useEditorStore, useNotifications } from '../stores';

interface EditorModalProps {
  video: app.VideoResult | null;
  opened: boolean;
  onClose: () => void;
}

const TAB_OPERATIONS: Record<string, string> = {
  trim: 'crop',
  convert: 'convert',
  watermark: 'watermark',
  effects: 'effects',
};

const RESOLUTION_OPTIONS = [
  { value: 'original', label: 'Original' },
  { value: '1080p', label: '1080p' },
  { value: '720p', label: '720p' },
  { value: '480p', label: '480p' },
];

export function EditorModal({ video, opened, onClose }: EditorModalProps) {
  const { options, ffmpegError, jobs, loadOptions, submitJob, cancelJob } = useEditorStore();
  const { success, error: notifyError } = useNotifications();

  const [metadata, setMetadata] = useState<editor.VideoMetadata | null>(null);
  const [activeTab, setActiveTab] = useState<string | null>('trim');

  // Trim tab state
  const [trimStart, setTrimStart] = useState<number | string>(0);
  const [trimEnd, setTrimEnd] = useState<number | string>(0);

  // Convert tab state
  const [outputFormat, setOutputFormat] = useState<string | null>(null);
  const [outputCodec, setOutputCodec] = useState<string | null>(null);
  const [outputQuality, setOutputQuality] = useState(23);
  const [outputResolution, setOutputResolution] = useState('original');

  // Watermark tab state
  const [watermarkType, setWatermarkType] = useState('text');
  const [watermarkText, setWatermarkText] = useState('');
  const [watermarkImage, setWatermarkImage] = useState('');
  const [watermarkPosition, setWatermarkPosition] = useState<string | null>(null);
  const [watermarkOpacity, setWatermarkOpacity] = useState(1);
  const [watermarkSize, setWatermarkSize] = useState<number | string>(24);

  // Effects tab state
  const [brightness, setBrightness] = useState(1);
  const [contrast, setContrast] = useState(1);
  const [saturation, setSaturation] = useState(1);
  const [speed, setSpeed] = useState(1);
  const [volume, setVolume] = useState(1);
  const [rotation, setRotation] = useState<string | null>('0');
  const [removeAudio, setRemoveAudio] = useState(false);

  // Preview and active job state
  const [previewUrl, setPreviewUrl] = useState<string | null>(null);
  const [previewLoading, setPreviewLoading] = useState(false);
  const [activeJobId, setActiveJobId] = useState<string | null>(null);

  const videoId = video?.id;
  const activeJob = activeJobId ? jobs[activeJobId] : undefined;

  // Reset form state and load options/metadata when the modal opens for a video
  useEffect(() => {
    if (!opened || !videoId) return;

    setMetadata(null);
    setActiveTab('trim');
    setTrimStart(0);
    setTrimEnd(0);
    setOutputFormat(null);
    setOutputCodec(null);
    setOutputQuality(23);
    setOutputResolution('original');
    setWatermarkType('text');
    setWatermarkText('');
    setWatermarkImage('');
    setWatermarkPosition(null);
    setWatermarkOpacity(1);
    setWatermarkSize(24);
    setRotation('0');
    setRemoveAudio(false);
    setPreviewUrl(null);
    setActiveJobId(null);

    loadOptions();

    GetEditVideoMetadata(videoId)
      .then(meta => {
        setMetadata(meta);
        setTrimEnd(meta.duration);
      })
      .catch(err => console.error('Failed to load video metadata:', err));
  }, [opened, videoId, loadOptions]);

  // Initialize effect sliders from the backend-provided ranges
  useEffect(() => {
    if (!options?.effect_ranges) return;
    const range = (id: string) => options.effect_ranges.find(r => r.id === id);
    setBrightness(range('brightness')?.default ?? 1);
    setContrast(range('contrast')?.default ?? 1);
    setSaturation(range('saturation')?.default ?? 1);
    setSpeed(range('speed')?.default ?? 1);
    setVolume(range('volume')?.default ?? 1);
  }, [options]);

  // Notify and close when the active job finishes or fails
  useEffect(() => {
    if (!activeJob) return;
    if (activeJob.status === 'completed') {
      success('Edit complete', 'The edited video has been added to your library');
      setActiveJobId(null);
      onClose();
    } else if (activeJob.status === 'error') {
      notifyError('Edit failed', activeJob.error || 'Unknown error');
      setActiveJobId(null);
    }
  }, [activeJob, success, notifyError, onClose]);

  const buildSettings = (): db.EditSettings => {
    switch (activeTab) {
      case 'trim':
        return {
          crop_start: Number(trimStart) || 0,
          crop_end: Number(trimEnd) || 0,
        };
      case 'convert':
        return {
          output_format: outputFormat || undefined,
          output_codec: outputCodec || undefined,
          output_quality: outputQuality,
          output_resolution: outputResolution,
        };
      case 'watermark':
        return {
          watermark_type: watermarkType,
          watermark_text: watermarkType === 'text' ? watermarkText : undefined,
          watermark_image: watermarkType === 'image' ? watermarkImage : undefined,
          watermark_position: watermarkPosition || undefined,
          watermark_opacity: watermarkOpacity,
          watermark_size: Number(watermarkSize) || 24,
        };
      case 'effects':
        return {
          brightness,
          contrast,
          saturation,
          speed,
          volume,
          rotation: Number(rotation) || 0,
          remove_audio: removeAudio,
        };
      default:
        return {};
    }
  };

  const handlePreview = async () => {
    if (!video) return;
    setPreviewLoading(true);
    try {
      // GenerateEditPreview returns a data: URL (base64 JPEG, see
      // internal/app/editor.go) — not a blob object URL — so there is
      // nothing to revoke on replace/unmount/close; dropping the state is enough.
      const url = await GenerateEditPreview(video.id, buildSettings(), Number(trimStart) || 0);
      setPreviewUrl(url);
    } catch (err) {
      notifyError(
        'Preview failed',
        err instanceof Error ? err.message : 'Failed to generate preview'
      );
    } finally {
      setPreviewLoading(false);
    }
  };

  const handleSubmit = async () => {
    if (!video) return;
    try {
      const jobId = await submitJob(video.id, TAB_OPERATIONS[activeTab || 'trim'], buildSettings());
      setActiveJobId(jobId);
    } catch (err) {
      notifyError('Edit failed', err instanceof Error ? err.message : 'Failed to submit edit job');
    }
  };

  const handleCancel = async () => {
    if (!activeJobId) return;
    await cancelJob(activeJobId);
    setActiveJobId(null);
  };

  const formatData = (options?.formats ?? []).map(f => ({ value: f.id, label: f.name }));
  const selectedFormat = options?.formats.find(f => f.id === outputFormat);
  const codecData = (options?.codecs ?? [])
    .filter(c => !selectedFormat || (selectedFormat.codecs ?? []).includes(c.id))
    .map(c => ({ value: c.id, label: c.name }));
  const positionData = Object.entries(options?.watermark_positions ?? {}).map(([value, label]) => ({
    value,
    label,
  }));
  const rotationData = (options?.rotations ?? []).map(r => ({
    value: String(r.value),
    label: r.label,
  }));

  const renderEffectSlider = (
    id: string,
    value: number,
    setValue: (value: number) => void,
    fallbackLabel: string
  ) => {
    const range = options?.effect_ranges?.find(r => r.id === id);
    return (
      <div key={id}>
        <Text size="sm">{range?.description || fallbackLabel}</Text>
        <Slider
          max={range?.max ?? 2}
          min={range?.min ?? 0}
          step={range?.step ?? 0.1}
          value={value}
          onChange={setValue}
        />
      </div>
    );
  };

  return (
    <Modal opened={opened} size="xl" title={`Edit: ${video?.title ?? ''}`} onClose={onClose}>
      {ffmpegError ? (
        <Alert color="yellow" icon={<IconAlertCircle size={16} />} title="FFmpeg unavailable">
          {ffmpegError}
        </Alert>
      ) : (
        <Stack gap="md">
          <Tabs value={activeTab} onChange={setActiveTab}>
            <Tabs.List>
              <Tabs.Tab value="trim">Trim</Tabs.Tab>
              <Tabs.Tab value="convert">Convert</Tabs.Tab>
              <Tabs.Tab value="watermark">Watermark</Tabs.Tab>
              <Tabs.Tab value="effects">Effects</Tabs.Tab>
            </Tabs.List>

            <Tabs.Panel pt="md" value="trim">
              <Stack gap="sm">
                <Group grow>
                  <NumberInput
                    decimalScale={2}
                    label="Start (seconds)"
                    max={metadata?.duration}
                    min={0}
                    value={trimStart}
                    onChange={setTrimStart}
                  />
                  <NumberInput
                    decimalScale={2}
                    label="End (seconds)"
                    max={metadata?.duration}
                    min={0}
                    value={trimEnd}
                    onChange={setTrimEnd}
                  />
                </Group>
                <Text c="dimmed" size="xs">
                  Total duration: {metadata ? `${metadata.duration.toFixed(1)}s` : 'loading...'}
                </Text>
              </Stack>
            </Tabs.Panel>

            <Tabs.Panel pt="md" value="convert">
              <Stack gap="sm">
                <Group grow>
                  <Select
                    data={formatData}
                    label="Format"
                    placeholder="Select format"
                    value={outputFormat}
                    onChange={value => {
                      setOutputFormat(value);
                      setOutputCodec(null);
                    }}
                  />
                  <Select
                    data={codecData}
                    label="Codec"
                    placeholder="Select codec"
                    value={outputCodec}
                    onChange={setOutputCodec}
                  />
                </Group>
                <div>
                  <Text size="sm">Quality (CRF): {outputQuality}</Text>
                  <Slider
                    marks={[
                      { value: 18, label: '18' },
                      { value: 23, label: '23' },
                      { value: 28, label: '28' },
                    ]}
                    max={28}
                    mb="lg"
                    min={18}
                    value={outputQuality}
                    onChange={setOutputQuality}
                  />
                  <Text c="dimmed" size="xs">
                    Lower = better quality
                  </Text>
                </div>
                <Select
                  data={RESOLUTION_OPTIONS}
                  label="Resolution"
                  value={outputResolution}
                  onChange={value => value && setOutputResolution(value)}
                />
              </Stack>
            </Tabs.Panel>

            <Tabs.Panel pt="md" value="watermark">
              <Stack gap="sm">
                <SegmentedControl
                  data={[
                    { value: 'text', label: 'Text' },
                    { value: 'image', label: 'Image' },
                  ]}
                  value={watermarkType}
                  onChange={setWatermarkType}
                />
                {watermarkType === 'text' ? (
                  <TextInput
                    label="Watermark text"
                    placeholder="Enter watermark text"
                    value={watermarkText}
                    onChange={e => setWatermarkText(e.currentTarget.value)}
                  />
                ) : (
                  <TextInput
                    label="Watermark image"
                    placeholder="/absolute/path/to/image.png"
                    value={watermarkImage}
                    onChange={e => setWatermarkImage(e.currentTarget.value)}
                  />
                )}
                <Select
                  data={positionData}
                  label="Position"
                  placeholder="Select position"
                  value={watermarkPosition}
                  onChange={setWatermarkPosition}
                />
                <div>
                  <Text size="sm">Opacity: {watermarkOpacity.toFixed(1)}</Text>
                  <Slider
                    max={1}
                    min={0}
                    step={0.1}
                    value={watermarkOpacity}
                    onChange={setWatermarkOpacity}
                  />
                </div>
                <NumberInput
                  description="Font size for text watermarks, scale for image watermarks"
                  label="Size"
                  min={1}
                  value={watermarkSize}
                  onChange={setWatermarkSize}
                />
              </Stack>
            </Tabs.Panel>

            <Tabs.Panel pt="md" value="effects">
              <Stack gap="sm">
                {renderEffectSlider('brightness', brightness, setBrightness, 'Brightness')}
                {renderEffectSlider('contrast', contrast, setContrast, 'Contrast')}
                {renderEffectSlider('saturation', saturation, setSaturation, 'Saturation')}
                {renderEffectSlider('speed', speed, setSpeed, 'Speed')}
                {renderEffectSlider('volume', volume, setVolume, 'Volume')}
                <Select
                  data={rotationData}
                  label="Rotation"
                  value={rotation}
                  onChange={setRotation}
                />
                <Checkbox
                  checked={removeAudio}
                  label="Remove audio"
                  onChange={e => setRemoveAudio(e.currentTarget.checked)}
                />
              </Stack>
            </Tabs.Panel>
          </Tabs>

          {activeJobId && activeJob && (
            <Stack gap="xs">
              <Group justify="space-between">
                <Text size="sm">Processing {activeJob.operation}...</Text>
                <Text size="sm">{Math.round(activeJob.progress * 100)}%</Text>
              </Group>
              <Progress animated value={activeJob.progress * 100} />
              <Group justify="flex-end">
                <Button color="red" size="xs" variant="light" onClick={handleCancel}>
                  Cancel
                </Button>
              </Group>
            </Stack>
          )}

          {previewLoading && (
            <Group justify="center">
              <Loader size="sm" />
            </Group>
          )}
          {!previewLoading && previewUrl && (
            <Image alt="Edit preview" fit="contain" mah={300} radius="sm" src={previewUrl} />
          )}

          <Group justify="flex-end">
            <Button
              disabled={previewLoading || activeJobId !== null}
              variant="light"
              onClick={handlePreview}
            >
              Preview
            </Button>
            <Button color="yted" disabled={activeJobId !== null} onClick={handleSubmit}>
              Submit
            </Button>
          </Group>
        </Stack>
      )}
    </Modal>
  );
}
