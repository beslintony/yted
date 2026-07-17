import { Badge, Button, Group, Paper, Stack, Text, TextInput } from '@mantine/core';
import { IconFolder, IconRefresh } from '@tabler/icons-react';

import { app, config } from '../../../wailsjs/go/models';

interface FfmpegSectionProps {
  dark: boolean;
  settings: config.Config;
  ffmpegStatus: app.FFmpegCheckResult | null;
  loadingFfmpeg: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
  handleBrowseFFmpeg: () => void;
  refreshFfmpegStatus: () => void;
  getSelectedFfmpegInfo: () => app.FFmpegLocation | null;
}

export function FfmpegSection({
  dark,
  settings,
  ffmpegStatus,
  loadingFfmpeg,
  updateSetting,
  handleBrowseFFmpeg,
  refreshFfmpegStatus,
  getSelectedFfmpegInfo,
}: FfmpegSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Group justify="space-between">
          <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
            FFmpeg Configuration
          </Text>
          <Group gap="xs">
            <Button
              color="gray"
              leftSection={<IconRefresh size={14} />}
              loading={loadingFfmpeg}
              size="sm"
              variant="light"
              onClick={refreshFfmpegStatus}
            >
              Refresh
            </Button>
            {loadingFfmpeg && (
              <Badge color="gray" variant="light">
                Checking...
              </Badge>
            )}
            {!loadingFfmpeg && ffmpegStatus && (
              <>
                {ffmpegStatus.installed ? (
                  <Badge color="green" variant="light">
                    Detected
                  </Badge>
                ) : (
                  <Badge color="red" variant="light">
                    Not Found
                  </Badge>
                )}
              </>
            )}
          </Group>
        </Group>

        {/* Detected FFmpeg Info */}
        {!loadingFfmpeg && ffmpegStatus?.installed && getSelectedFfmpegInfo() && (
          <Paper
            bg={dark ? '#1b2c1b' : '#f0f9f0'}
            p="sm"
            style={{
              borderColor: dark ? '#2ac92a' : '#28a745',
              borderWidth: 1,
              borderStyle: 'solid',
            }}
          >
            <Stack gap={4}>
              <Text c={dark ? '#2ac92a' : '#28a745'} fw={600} size="sm">
                FFmpeg Detected
              </Text>
              <Text c={dark ? '#c1c2c5' : '#495057'} size="sm">
                <strong>Path:</strong> {getSelectedFfmpegInfo()?.path}
              </Text>
              <Text c={dark ? '#c1c2c5' : '#495057'} size="sm">
                <strong>Version:</strong> {getSelectedFfmpegInfo()?.version}
              </Text>
              <Text c={dark ? '#909296' : '#6c757d'} size="xs">
                Source: {getSelectedFfmpegInfo()?.source}
              </Text>
            </Stack>
          </Paper>
        )}

        {/* Not Found Warning */}
        {!loadingFfmpeg && ffmpegStatus && !ffmpegStatus.installed && (
          <Paper
            bg={dark ? '#2c1b1b' : '#fdf2f2'}
            p="sm"
            style={{
              borderColor: dark ? '#c92a2a' : '#dc3545',
              borderWidth: 1,
              borderStyle: 'solid',
            }}
          >
            <Stack gap="xs">
              <Text c={dark ? '#ff6b6b' : '#dc3545'} fw={600} size="sm">
                FFmpeg Not Found
              </Text>
              <Text c={dark ? '#c1c2c5' : '#495057'} size="sm">
                FFmpeg is required for merging separate video and audio streams into a single file.
                Without FFmpeg configured:
              </Text>
              <ul
                style={{
                  margin: 0,
                  paddingLeft: 20,
                  color: dark ? '#c1c2c5' : '#495057',
                  fontSize: '0.875rem',
                }}
              >
                <li>Videos may be downloaded without audio</li>
                <li>Separate video and audio files will not be merged</li>
                <li>Some formats may not work correctly</li>
              </ul>
              <Text c={dark ? '#c1c2c5' : '#495057'} mt={4} size="sm">
                Please specify the path to your FFmpeg binary below, or install FFmpeg and restart
                YTed.
              </Text>
            </Stack>
          </Paper>
        )}

        {/* Other Found Locations */}
        {!loadingFfmpeg &&
          ffmpegStatus &&
          ffmpegStatus.allLocations &&
          ffmpegStatus.allLocations.length > 1 && (
            <Paper
              bg={dark ? '#25262b' : '#f8f9fa'}
              p="sm"
              style={{
                borderColor: dark ? '#373a40' : '#dee2e6',
                borderWidth: 1,
                borderStyle: 'solid',
              }}
            >
              <Text c={dark ? '#c1c2c5' : '#495057'} fw={600} mb={8} size="sm">
                Other FFmpeg Installations Found ({ffmpegStatus.allLocations.length - 1})
              </Text>
              <Stack gap={4}>
                {ffmpegStatus.allLocations.map((loc, idx) => {
                  if (idx === ffmpegStatus.selectedIndex) return null;
                  return (
                    <Group key={loc.path} gap="xs">
                      <Text c={dark ? '#909296' : '#6c757d'} size="xs" style={{ flex: 1 }}>
                        {loc.path}
                      </Text>
                      <Text c={dark ? '#909296' : '#6c757d'} size="xs">
                        v{loc.version}
                      </Text>
                      <Badge color="gray" size="xs" variant="outline">
                        {loc.source}
                      </Badge>
                    </Group>
                  );
                })}
              </Stack>
            </Paper>
          )}

        <Text c={dark ? 'dimmed' : 'gray.6'} size="sm">
          You can specify a custom FFmpeg path below, or leave empty to auto-detect from system
          PATH.
        </Text>

        <Group align="flex-end" gap="sm">
          <TextInput
            description="Path to ffmpeg executable (optional)"
            label="FFmpeg Path"
            placeholder="Auto-detect from PATH"
            style={{ flex: 1 }}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.ffmpeg_path || ''}
            onChange={e => updateSetting('ffmpeg_path', e.currentTarget.value)}
          />
          <Button
            color="yted"
            leftSection={<IconFolder size={16} />}
            variant="light"
            onClick={handleBrowseFFmpeg}
          >
            Browse
          </Button>
        </Group>

        {settings.ffmpeg_path && (
          <Button
            color="gray"
            leftSection={<IconRefresh size={16} />}
            size="sm"
            variant="light"
            onClick={async () => {
              updateSetting('ffmpeg_path', '');
              await refreshFfmpegStatus();
            }}
          >
            Reset to Auto-Detect
          </Button>
        )}
      </Stack>
    </Paper>
  );
}
