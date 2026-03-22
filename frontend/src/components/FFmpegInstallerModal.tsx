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
  IconDownload,
  IconExternalLink,
  IconInfoCircle,
  IconTerminal,
  IconTool,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';

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
  const [copied, setCopied] = useState(false);
  const { success, error } = useNotifications();

  useEffect(() => {
    if (opened) {
      checkFFmpegStatus();
    }
  }, [opened]);

  const checkFFmpegStatus = async () => {
    console.log('[FFmpegInstallerModal] Checking FFmpeg status...');
    setIsChecking(true);
    try {
      const result = (await CheckFFmpegWithGuidance()) as FFmpegCheckResult;
      console.log('[FFmpegInstallerModal] FFmpeg check result:', result);
      setCheckResult(result);

      if (!result.installed) {
        console.log('[FFmpegInstallerModal] FFmpeg not installed, getting guide...');
        const guide = (await GetFFmpegInstallGuide()) as InstallGuide;
        console.log('[FFmpegInstallerModal] Install guide:', guide);
        setInstallGuide(guide);
      } else {
        console.log('[FFmpegInstallerModal] FFmpeg installed, calling onInstalled');
        onInstalled();
      }
    } catch (err) {
      console.error('[FFmpegInstallerModal] Error checking FFmpeg:', err);
      error('Error', 'Failed to check FFmpeg status');
    } finally {
      setIsChecking(false);
    }
  };

  const handleCopyCommand = () => {
    if (installGuide?.command) {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
      success('Copied!', 'Command copied to clipboard');
    }
  };

  const handleOpenDownloadPage = async () => {
    try {
      await OpenFFmpegDownloadPage();
    } catch (err) {
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

  // Don't render if FFmpeg is already installed
  if (checkResult?.installed) {
    console.log('[FFmpegInstallerModal] FFmpeg installed, not rendering');
    return null;
  }

  return (
    <Modal
      opened={opened}
      onClose={onClose}
      size="lg"
      title={
        <Group gap="sm">
          <ThemeIcon color="red" size="lg" variant="filled">
            <IconTool size={20} />
          </ThemeIcon>
          <Title order={4}>FFmpeg Required</Title>
        </Group>
      }
    >
      <Stack gap="md">
        <Alert color="yellow" icon={<IconInfoCircle size={16} />}>
          <Text size="sm">
            FFmpeg is required for the video editor to work. It&apos;s used for processing
            video operations like cropping, watermarking, and format conversion.
          </Text>
        </Alert>

        {installGuide && (
          <>
            <Paper withBorder p="md">
              <Stack gap="md">
                <Group gap="xs">
                  <IconInfoCircle size={18} color="var(--mantine-color-blue-6)" />
                  <Text fw={600}>{installGuide.title}</Text>
                </Group>

                <Text size="sm" c="dimmed">
                  {installGuide.description}
                </Text>

                {installGuide.steps && installGuide.steps.length > 0 && (
                  <Timeline active={-1} bulletSize={24} lineWidth={2}>
                    {installGuide.steps.map((step, index) => (
                      <Timeline.Item
                        key={index}
                        bullet={<Text size="xs">{index + 1}</Text>}
                        title={<Text size="sm">{step}</Text>}
                      />
                    ))}
                  </Timeline>
                )}

                {installGuide.command && (
                  <Paper withBorder p="sm" bg="var(--mantine-color-default)">
                    <Stack gap="xs">
                      <Group justify="space-between">
                        <Text size="xs" c="dimmed">
                          {installGuide.commandDescription}
                        </Text>
                        <CopyButton value={installGuide.command} timeout={2000}>
                          {({ copied: isCopied, copy }) => (
                            <Button
                              size="xs"
                              variant="light"
                              leftSection={isCopied ? <IconCheck size={14} /> : <IconCopy size={14} />}
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

                {installGuide.tips && installGuide.tips.length > 0 && (
                  <Stack gap="xs">
                    <Text size="sm" fw={500}>
                      Tips:
                    </Text>
                    <List size="sm" c="dimmed">
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
                variant="light"
                leftSection={<IconExternalLink size={16} />}
                onClick={handleOpenDownloadPage}
              >
                Download FFmpeg
              </Button>
              {checkResult?.canAutoInstall && (
                <Button
                  variant="light"
                  leftSection={<IconTerminal size={16} />}
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
