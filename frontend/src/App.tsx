import { useState } from 'react';
import {
  AppShell,
  Navbar,
  Header,
  Footer,
  Text,
  MediaQuery,
  Burger,
  useMantineTheme,
  Group,
  ActionIcon,
  Tooltip,
  Stack,
  rem,
} from '@mantine/core';
import {
  IconDownload,
  IconVideo,
  IconSettings,
  IconMenu2,
  IconSun,
  IconMoon,
} from '@tabler/icons-react';
import { useSettingsStore } from './stores';
import { DownloadPage } from './pages/DownloadPage';
import { LibraryPage } from './pages/LibraryPage';
import { SettingsPage } from './pages/SettingsPage';

function App() {
  const theme = useMantineTheme();
  const [opened, setOpened] = useState(false);
  const [activeTab, setActiveTab] = useState<'downloads' | 'library' | 'settings'>('downloads');
  const { theme: themeMode, toggleSidebar, sidebarCollapsed } = useSettingsStore();

  const navItems = [
    { id: 'downloads' as const, label: 'Downloads', icon: IconDownload },
    { id: 'library' as const, label: 'Library', icon: IconVideo },
    { id: 'settings' as const, label: 'Settings', icon: IconSettings },
  ];

  return (
    <AppShell
      styles={{
        main: {
          background: theme.colorScheme === 'dark' ? theme.colors.dark[8] : theme.colors.gray[0],
        },
      }}
      navbarOffsetBreakpoint="sm"
      asideOffsetBreakpoint="sm"
      navbar={
        <Navbar p="md" hiddenBreakpoint="sm" hidden={!opened} width={{ sm: 200, lg: sidebarCollapsed ? 80 : 250 }}>
          <Navbar.Section grow>
            <Stack spacing="xs">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = activeTab === item.id;
                return (
                  <ActionIcon
                    key={item.id}
                    variant={isActive ? 'filled' : 'subtle'}
                    color={isActive ? 'yted' : 'gray'}
                    onClick={() => {
                      setActiveTab(item.id);
                      setOpened(false);
                    }}
                    size="xl"
                    sx={{
                      width: sidebarCollapsed ? rem(50) : '100%',
                      justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                      padding: sidebarCollapsed ? 0 : rem(12),
                    }}
                    title={item.label}
                  >
                    <Group spacing="sm" noWrap>
                      <Icon size={20} />
                      {!sidebarCollapsed && (
                        <Text size="sm" fw={500}>
                          {item.label}
                        </Text>
                      )}
                    </Group>
                  </ActionIcon>
                );
              })}
            </Stack>
          </Navbar.Section>
        </Navbar>
      }
      footer={
        <Footer height={40} p="xs">
          <Group position="apart">
            <Text size="xs" c="dimmed">
              YTed v1.0.0
            </Text>
            <Text size="xs" c="dimmed">
              Ready
            </Text>
          </Group>
        </Footer>
      }
      header={
        <Header height={{ base: 50, md: 60 }} p="sm">
          <div style={{ display: 'flex', alignItems: 'center', height: '100%' }}>
            <MediaQuery largerThan="sm" styles={{ display: 'none' }}>
              <Burger
                opened={opened}
                onClick={() => setOpened((o) => !o)}
                size="sm"
                color={theme.colors.gray[6]}
                mr="xl"
              />
            </MediaQuery>
            
            <Group position="apart" style={{ flex: 1 }}>
              <Group spacing="sm">
                <img
                  src="/logo.svg"
                  alt="YTed"
                  style={{ height: 32, width: 32 }}
                />
                <Text fw={700} size="lg">YTed</Text>
              </Group>
              
              <Group spacing="xs">
                <Tooltip label="Toggle sidebar">
                  <ActionIcon variant="subtle" onClick={toggleSidebar}>
                    <IconMenu2 size={20} />
                  </ActionIcon>
                </Tooltip>
                <Tooltip label="Toggle theme">
                  <ActionIcon variant="subtle">
                    {theme.colorScheme === 'dark' ? <IconSun size={20} /> : <IconMoon size={20} />}
                  </ActionIcon>
                </Tooltip>
              </Group>
            </Group>
          </div>
        </Header>
      }
    >
      <div style={{ padding: theme.spacing.md }}>
        {activeTab === 'downloads' && <DownloadPage />}
        {activeTab === 'library' && <LibraryPage />}
        {activeTab === 'settings' && <SettingsPage />}
      </div>
    </AppShell>
  );
}

export default App;
