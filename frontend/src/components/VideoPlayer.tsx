import { Alert, Paper, Stack, Text, useMantineColorScheme } from '@mantine/core';
import { IconAlertCircle, IconVideo } from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import { GetVideoFile } from '../../wailsjs/go/app/App';
import { VideoMetadata } from '../types/editor';

interface VideoPlayerProps {
  videoId: string;
  format?: string;
  previewFrame?: string | null;
  isGeneratingPreview?: boolean;
  metadata?: VideoMetadata | null;
  onPlay?: () => void;
}

// Get MIME type from file format (extension)
function getVideoMimeType(format?: string): string {
  if (!format) return 'video/mp4';

  const formatLower = format.toLowerCase().replace(/^\./, '');
  switch (formatLower) {
    case 'webm':
      return 'video/webm';
    case 'ogv':
    case 'ogg':
      return 'video/ogg';
    case 'mkv':
      return 'video/x-matroska';
    case 'mov':
      return 'video/quicktime';
    case 'avi':
      return 'video/x-msvideo';
    case 'mp4':
    default:
      return 'video/mp4';
  }
}

export function VideoPlayer({
  videoId,
  format,
  previewFrame,
  isGeneratingPreview,
  metadata,
  onPlay,
}: VideoPlayerProps) {
  useMantineColorScheme();
  const [videoUrl, setVideoUrl] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (!videoId) {
      setVideoUrl(null);
      setError(null);
      return;
    }

    // Load video file as blob URL
    let objectUrl: string | null = null;

    const loadVideo = async () => {
      setIsLoading(true);
      setError(null);
      try {
        console.warn('[VideoPlayer] Loading video:', videoId, 'format:', format);
        const data = await GetVideoFile(videoId);
        console.warn('[VideoPlayer] Video data loaded, size:', data.length);

        // Convert byte array to blob with proper MIME type
        const mimeType = getVideoMimeType(format);
        console.warn('[VideoPlayer] Using MIME type:', mimeType);

        const blob = new Blob([new Uint8Array(data)], { type: mimeType });
        objectUrl = URL.createObjectURL(blob);
        console.warn('[VideoPlayer] Object URL created:', objectUrl);

        setVideoUrl(objectUrl);
      } catch (err) {
        console.error('[VideoPlayer] Failed to load video:', err);
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
  }, [videoId, format]);

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
            alt="Preview"
            src={previewFrame}
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              objectFit: 'contain',
            }}
          />
        ) : videoUrl ? (
          <video
            key={videoUrl}
            controls
            controlsList="nodownload"
            preload="metadata"
            style={{
              maxWidth: '100%',
              maxHeight: '100%',
              width: '100%',
              height: '100%',
            }}
            onClick={onPlay}
          >
            <source src={videoUrl} type={getVideoMimeType(format)} />
            <track kind="captions" />
            Your browser does not support the video tag.
          </video>
        ) : error ? (
          <Stack align="center" gap="md" p="xl">
            <IconAlertCircle color="var(--mantine-color-red-5)" size={48} />
            <Alert color="red" title="Failed to load video" variant="light">
              <Text size="sm">{error}</Text>
            </Alert>
          </Stack>
        ) : (
          <Stack align="center" gap="md">
            <IconVideo color="var(--mantine-color-gray-5)" size={64} />
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
        <Paper withBorder bg="var(--mantine-color-body)" p="sm">
          <Text c="dimmed" size="sm">
            {metadata.width}x{metadata.height} • {metadata.codec?.toUpperCase() || 'Unknown Codec'}
            {metadata.duration > 0 && ` • ${formatDuration(metadata.duration)}`}
            {metadata.hasAudio ? ' • With Audio' : ' • No Audio'}
          </Text>
        </Paper>
      )}
    </Stack>
  );
}
