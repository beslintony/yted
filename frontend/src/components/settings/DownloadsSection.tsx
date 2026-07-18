import { Button, Group, NumberInput, Paper, Select, Stack, Text, TextInput } from '@mantine/core';
import { IconFolder } from '@tabler/icons-react';

import { config } from '../../../wailsjs/go/models';

interface DownloadsSectionProps {
  dark: boolean;
  settings: config.Config;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  updateSetting: (key: string, value: any) => void;
  handleBrowseDownloadPath: () => void;
}

export function DownloadsSection({
  dark,
  settings,
  updateSetting,
  handleBrowseDownloadPath,
}: DownloadsSectionProps) {
  return (
    <Paper
      withBorder
      bg={dark ? '#25262b' : '#fff'}
      p="md"
      style={{ borderColor: dark ? '#373a40' : '#dee2e6' }}
    >
      <Stack gap="md">
        <Text c={dark ? '#fff' : '#000'} fw={600} size="lg">
          Downloads
        </Text>

        <Group align="flex-end" gap="sm">
          <TextInput
            readOnly
            description="Where downloaded videos are saved"
            label="Download Path"
            style={{ flex: 1 }}
            styles={{
              input: {
                background: dark ? '#1a1b1e' : '#f8f9fa',
                color: dark ? '#c1c2c5' : '#212529',
              },
            }}
            value={settings.download_path}
          />
          <Button
            color="yted"
            leftSection={<IconFolder size={16} />}
            variant="light"
            onClick={handleBrowseDownloadPath}
          >
            Browse
          </Button>
        </Group>

        <NumberInput
          description="Number of simultaneous downloads (1-10)"
          label="Max Concurrent Downloads"
          max={10}
          min={1}
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.max_concurrent_downloads}
          w={200}
          onChange={v => updateSetting('max_concurrent_downloads', v || 1)}
        />

        <Select
          data={[
            { value: 'best', label: 'Best Quality' },
            { value: '1080p', label: '1080p' },
            { value: '720p', label: '720p' },
            { value: '480p', label: '480p' },
            { value: '360p', label: '360p' },
            { value: 'audio', label: 'Audio Only' },
          ]}
          description="Preferred quality for new downloads"
          label="Default Quality"
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.default_quality}
          w={200}
          onChange={v => v && updateSetting('default_quality', v)}
        />

        <TextInput
          description="Template for output filenames (yt-dlp format)"
          label="Filename Template"
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.filename_template}
          onChange={e => updateSetting('filename_template', e.currentTarget.value)}
        />

        <Select
          clearable
          data={[
            { value: 'firefox', label: 'Firefox' },
            { value: 'chrome', label: 'Chrome' },
            { value: 'chromium', label: 'Chromium' },
            { value: 'brave', label: 'Brave' },
            { value: 'edge', label: 'Edge' },
          ]}
          description="Use your browser's YouTube login to bypass bot checks (sign in to YouTube in that browser first)"
          label="Cookies from Browser"
          placeholder="Disabled"
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.cookies_browser || null}
          w={250}
          onChange={v => updateSetting('cookies_browser', v || '')}
        />

        <TextInput
          description="Path to a Netscape cookies.txt file (takes precedence over browser cookies)"
          label="Cookies File (optional)"
          placeholder="/path/to/cookies.txt"
          styles={{
            input: {
              background: dark ? '#1a1b1e' : '#f8f9fa',
              color: dark ? '#c1c2c5' : '#212529',
            },
          }}
          value={settings.cookies_file}
          onChange={e => updateSetting('cookies_file', e.currentTarget.value)}
        />
      </Stack>
    </Paper>
  );
}
