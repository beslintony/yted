import { Paper, Stack, Text } from '@mantine/core';
import { IconPlayerPlay } from '@tabler/icons-react';

import { VideoMetadata } from '../types/editor';

interface VideoPlayerProps {
  src: string;
  previewFrame?: string | null;
  isGeneratingPreview?: boolean;
  metadata?: VideoMetadata | null;
}

export function VideoPlayer({
  src,
  previewFrame,
  isGeneratingPreview,
  metadata,
}: VideoPlayerProps) {
  const formatDuration = (seconds: number) => {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins}:${secs.toString().padStart(2, '0')}`;
  };

  return (
    <Stack gap="md" h="100%">
      <Paper
        withBorder
        style={{
          flex: 1,
          position: 'relative',
          overflow: 'hidden',
          backgroundColor: '#000',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
        }}
      >
        {previewFrame ? (
          <img
            src={previewFrame}
            alt="Preview"
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              objectFit: 'contain',
            }}
          />
        ) : (
          <video
            src={`file://${src}`}
            controls
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
            }}
          >
            <track kind="captions" />
          </video>
        )}

        {isGeneratingPreview && (
          <Paper
            style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              backgroundColor: 'rgba(0, 0, 0, 0.7)',
              display: 'flex',
              alignItems: 'center',
              justifyContent: 'center',
            }}
          >
            <Text c="white">Generating Preview...</Text>
          </Paper>
        )}
      </Paper>

      {metadata && (
        <Paper withBorder p="sm">
          <Text size="sm" c="dimmed">
            {metadata.width}x{metadata.height} • {metadata.codec?.toUpperCase() || 'Unknown Codec'}
            {metadata.duration > 0 && ` • ${formatDuration(metadata.duration)}`}
            {metadata.hasAudio ? ' • With Audio' : ' • No Audio'}
          </Text>
        </Paper>
      )}
    </Stack>
  );
}
