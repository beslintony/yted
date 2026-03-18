
import {
  ActionIcon,
  AppShell,
  Burger,
  Drawer,
  Group,
  Stack,
  Tooltip,
  useMantineColorScheme,
} from '@mantine/core';
import {
  IconDownload,
  IconFileText,
  IconMoon,
  IconSettings,
  IconSun,
  IconVideo,
} from '@tabler/icons-react';
import { useEffect, useState } from 'react';

import { LoggerViewer } from './components/LoggerViewer';
import { DownloadPage } from './pages/DownloadPage';
import { LibraryPage } from './pages/LibraryPage';
import { SettingsPage } from './pages/SettingsPage';
import { useSettingsStore, useVersionStore } from './stores';

function App() {
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const [mobileOpened, setMobileOpened] = useState(false);
  const [activeTab, setActiveTab] = useState<'downloads' | 'library' | 'settings'>('downloads');
  const [loggerOpened, setLoggerOpened] = useState(false);
  const { sidebarCollapsed, toggleSidebar } = useSettingsStore();
  const { version, fetchVersion } = useVersionStore();

  // Fetch version on mount
  useEffect(() => {
    fetchVersion();
  }, [fetchVersion]);

  const dark = colorScheme === 'dark';

  const navItems = [
    { id: 'downloads' as const, label: 'Downloads', icon: IconDownload },
    { id: 'library' as const, label: 'Library', icon: IconVideo },
    { id: 'settings' as const, label: 'Settings', icon: IconSettings },
  ];

  const handleThemeToggle = () => {
    const newScheme = dark ? 'light' : 'dark';
    setColorScheme(newScheme);
  };

  return (
    <AppShell
      footer={{ height: 40 }}
      header={{ height: 60 }}
      navbar={{
        width: sidebarCollapsed ? 80 : 200,
        breakpoint: 'sm',
        collapsed: { mobile: !mobileOpened },
      }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" justify="space-between" px="md">
          <Group>
            {/* Desktop hamburger - toggles sidebar */}
            <Burger
              aria-label="Toggle sidebar"
              color={dark ? '#c1c2c5' : '#495057'}
              opened={!sidebarCollapsed}
              size="sm"
              onClick={toggleSidebar}
            />

            {/* Mobile hamburger - toggles mobile menu */}
            <Burger
              aria-label="Toggle mobile menu"
              color={dark ? '#c1c2c5' : '#495057'}
              hiddenFrom="sm"
              opened={mobileOpened}
              size="sm"
              onClick={() => setMobileOpened(o => !o)}
            />

            {/* Logo */}
            <Tooltip label={`YTed v${version}`}>
              <img
                alt="YTed"
                src="/logo.svg"
                style={{ height: 36, width: 36, cursor: 'pointer' }}
              />
            </Tooltip>
          </Group>

          <Group gap="xs">
            <Tooltip label="View logs">
              <ActionIcon
                color="gray"
                size="lg"
                variant="subtle"
                onClick={() => setLoggerOpened(true)}
              >
                <IconFileText size={22} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label={dark ? 'Switch to light mode' : 'Switch to dark mode'}>
              <ActionIcon color="gray" size="lg" variant="subtle" onClick={handleThemeToggle}>
                {dark ? <IconSun size={22} /> : <IconMoon size={22} />}
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="xs">
          {navItems.map(item => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <Tooltip
                key={item.id}
                disabled={!sidebarCollapsed}
                label={item.label}
                position="right"
              >
                <ActionIcon
                  color={isActive ? 'yted' : 'gray'}
                  h={44}
                  style={{
                    justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                    padding: sidebarCollapsed ? 0 : '0 12px',
                    borderRadius: 8,
                  }}
                  variant={isActive ? 'filled' : 'subtle'}
                  w={sidebarCollapsed ? 60 : '100%'}
                  onClick={() => {
                    setActiveTab(item.id);
                    setMobileOpened(false);
                  }}
                >
                  <Group
                    gap={12}
                    style={{
                      justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                      width: '100%',
                    }}
                    wrap="nowrap"
                  >
                    <Icon size={20} />
                    {!sidebarCollapsed && (
                      <span style={{ fontSize: 14, fontWeight: 500, textAlign: 'left' }}>
                        {item.label}
                      </span>
                    )}
                  </Group>
                </ActionIcon>
              </Tooltip>
            );
          })}
        </Stack>
      </AppShell.Navbar>

      <AppShell.Main bg={dark ? '#1a1b1e' : '#f8f9fa'}>
        {activeTab === 'downloads' && <DownloadPage />}
        {activeTab === 'library' && <LibraryPage />}
        {activeTab === 'settings' && <SettingsPage />}
      </AppShell.Main>

      <AppShell.Footer>
        <Group h="100%" justify="space-between" px="md">
          <span style={{ fontSize: 12, color: dark ? '#909296' : '#868e96' }}>YTed v{version}</span>
          <span style={{ fontSize: 12, color: dark ? '#909296' : '#868e96' }}>Ready</span>
        </Group>
      </AppShell.Footer>

      {/* Logger Drawer */}
      <Drawer
        opened={loggerOpened}
        padding="md"
        position="right"
        size="lg"
        title="Application Logs"
        onClose={() => setLoggerOpened(false)}
      >
        <LoggerViewer />
      </Drawer>
    </AppShell>
  );
}

export default App;
