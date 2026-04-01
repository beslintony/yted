import {
  Alert,
  Button,
  Code,
  CopyButton,
  Group,
  List,
  Modal,
  Paper,
  Stack,
  Text,
  ThemeIcon,
  Timeline,
  Title,
} from '@mantine/core';
import {
  IconCheck,
  IconCopy,
  IconExternalLink,
  IconInfoCircle,
  IconTerminal,
  IconTool,
} from '@tabler/icons-react';
import { useCallback, useEffect, useState } from 'react';

import {
  CheckFFmpegWithGuidance,
  GetFFmpegInstallGuide,
  OpenFFmpegDownloadPage,
} from '../../wailsjs/go/app/App';
import { useNotifications } from '../stores';
import { FFmpegCheckResult, InstallGuide } from '../types/editor';

interface FFmpegInstallerModalProps {
  opened: boolean;
  onClose: () => void;
  onInstalled: () => void;
}

export function FFmpegInstallerModal({ opened, onClose, onInstalled }: FFmpegInstallerModalProps) {
  const [checkResult, setCheckResult] = useState<FFmpegCheckResult | null>(null);
  const [installGuide, setInstallGuide] = useState<InstallGuide | null>(null);
  const [isChecking, setIsChecking] = useState(false);

  const { success, error } = useNotifications();

  const checkFFmpegStatus = useCallback(async () => {
    setIsChecking(true);
    try {
      const result = (await CheckFFmpegWithGuidance()) as FFmpegCheckResult;
      setCheckResult(result);

      if (!result.installed) {
        const guide = (await GetFFmpegInstallGuide()) as InstallGuide;
        setInstallGuide(guide);
      } else {
        onInstalled();
      }
    } catch {
      error('Error', 'Failed to check FFmpeg status');
    } finally {
      setIsChecking(false);
    }
  }, [error, onInstalled]);

  useEffect(() => {
    if (opened) {
      checkFFmpegStatus();
    }
  }, [opened, checkFFmpegStatus]);



  const handleCopyCommand = () => {
    if (installGuide?.command) {
      success('Copied!', 'Command copied to clipboard');
    }
  };

  const handleOpenDownloadPage = async () => {
    try {
      await OpenFFmpegDownloadPage();
    } catch {
      error('Error', 'Failed to open download page');
    }
  };

  const handleCheckAgain = async () => {
    await checkFFmpegStatus();
    if (checkResult?.installed) {
      success('FFmpeg Found', 'FFmpeg is now installed and ready to use');
      onInstalled();
    }
  };

  // Don't show if FFmpeg is already installed
  if (checkResult?.installed) {
    return null;
  }

  return (
    <Modal
      opened={opened}
      size="lg"
      title={
        <Group gap="sm">
          <ThemeIcon color="red" size="lg" variant="filled">
            <IconTool size={20} />
          </ThemeIcon>
          <Title order={4}>FFmpeg Required</Title>
        </Group>
      }
      onClose={onClose}
    >
      <Stack gap="md">
        <Alert color="yellow" icon={<IconInfoCircle size={16} />}>
          <Text size="sm">
            FFmpeg is required for the video editor to work. It&apos;s used for processing video
            operations like cropping, watermarking, and format conversion.
          </Text>
        </Alert>

        {installGuide && (
          <>
            <Paper withBorder p="md">
              <Stack gap="md">
                <Group gap="xs">
                  <IconInfoCircle color="#228be6" size={18} />
                  <Text fw={600}>{installGuide.title}</Text>
                </Group>

                <Text c="dimmed" size="sm">
                  {installGuide.description}
                </Text>

                <Timeline active={-1} bulletSize={24} lineWidth={2}>
                  {installGuide.steps.map((step, index) => (
                    <Timeline.Item
                      key={index}
                      bullet={<Text size="xs">{index + 1}</Text>}
                      title={<Text size="sm">{step}</Text>}
                    />
                  ))}
                </Timeline>

                {installGuide.command && (
                  <Paper withBorder bg="gray.0" p="sm">
                    <Stack gap="xs">
                      <Group justify="space-between">
                        <Text c="dimmed" size="xs">
                          {installGuide.commandDescription}
                        </Text>
                        <CopyButton timeout={2000} value={installGuide.command}>
                          {({ copied: isCopied, copy }) => (
                            <Button
                              leftSection={
                                isCopied ? <IconCheck size={14} /> : <IconCopy size={14} />
                              }
                              size="xs"
                              variant="light"
                              onClick={() => {
                                copy();
                                handleCopyCommand();
                              }}
                            >
                              {isCopied ? 'Copied!' : 'Copy'}
                            </Button>
                          )}
                        </CopyButton>
                      </Group>
                      <Code block>{installGuide.command}</Code>
                    </Stack>
                  </Paper>
                )}

                {installGuide.tips.length > 0 && (
                  <Stack gap="xs">
                    <Text fw={500} size="sm">
                      Tips:
                    </Text>
                    <List c="dimmed" size="sm">
                      {installGuide.tips.map((tip, index) => (
                        <List.Item key={index}>{tip}</List.Item>
                      ))}
                    </List>
                  </Stack>
                )}
              </Stack>
            </Paper>

            <Group grow>
              <Button
                leftSection={<IconExternalLink size={16} />}
                variant="light"
                onClick={handleOpenDownloadPage}
              >
                Download FFmpeg
              </Button>
              {checkResult?.canAutoInstall && (
                <Button
                  leftSection={<IconTerminal size={16} />}
                  variant="light"
                  onClick={() => {
                    // On macOS with Homebrew, we could potentially auto-install
                    // but for security, we'll just show the command
                    handleCopyCommand();
                  }}
                >
                  Auto-Install (Copy)
                </Button>
              )}
            </Group>
          </>
        )}

        <Group justify="space-between">
          <Button variant="subtle" onClick={onClose}>
            Skip for Now
          </Button>
          <Button
            leftSection={<IconCheck size={16} />}
            loading={isChecking}
            onClick={handleCheckAgain}
          >
            I&apos;ve Installed It
          </Button>
        </Group>
      </Stack>
    </Modal>
  );
}
