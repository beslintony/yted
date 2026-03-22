import { Paper, Stack, Text, useMantineColorScheme } from '@mantine/core';
import { IconVideo } from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import { GetVideoFile } from '../../wailsjs/go/app/App';
import { VideoMetadata } from '../types/editor';

interface VideoPlayerProps {
  videoId: string;
  previewFrame?: string | null;
  isGeneratingPreview?: boolean;
  metadata?: VideoMetadata | null;
  onPlay?: () => void;
}

export function VideoPlayer({
  videoId,
  previewFrame,
  isGeneratingPreview,
  metadata,
  onPlay,
}: VideoPlayerProps) {
  const { colorScheme } = useMantineColorScheme();
  const [videoUrl, setVideoUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!videoId) {
      setVideoUrl(null);
      return;
    }

    // Load video file as blob URL
    let objectUrl: string | null = null;

    const loadVideo = async () => {
      setIsLoading(true);
      setError(null);
      try {
        const data = await GetVideoFile(videoId);
        // Convert byte array to blob
        const blob = new Blob([new Uint8Array(data)], { type: 'video/mp4' });
        objectUrl = URL.createObjectURL(blob);
        setVideoUrl(objectUrl);
      } catch (err) {
        console.error('Failed to load video:', err);
        setError(err instanceof Error ? err.message : 'Failed to load video');
        setVideoUrl(null);
      } finally {
        setIsLoading(false);
      }
    };

    loadVideo();

    // Cleanup object URL on unmount or when videoId changes
    return () => {
      if (objectUrl) {
        URL.revokeObjectURL(objectUrl);
      }
    };
  }, [videoId]);

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
          backgroundColor: 'var(--mantine-color-dark-filled)',
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
        ) : videoUrl ? (
          <video
            src={videoUrl}
            controls
            onClick={onPlay}
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              width: '100%',
              height: '100%',
            }}
            controlsList="nodownload"
            preload="metadata"
          >
            <track kind="captions" />
          </video>
        ) : error ? (
          <Stack align="center" gap="md">
            <Text c="red" ta="center">
              Error: {error}
            </Text>
          </Stack>
        ) : (
          <Stack align="center" gap="md">
            <IconVideo size={64} color="var(--mantine-color-gray-5)" />
            <Text c="dimmed" ta="center">
              {isLoading ? 'Loading video...' : 'No video selected'}
            </Text>
          </Stack>
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
              zIndex: 10,
            }}
          >
            <Text c="white">Generating Preview...</Text>
          </Paper>
        )}
      </Paper>

      {metadata && (
        <Paper withBorder p="sm" bg="var(--mantine-color-body)">
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
