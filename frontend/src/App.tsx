import { useState } from 'react';
import {
  AppShell,
  Burger,
  Group,
  Text,
  ActionIcon,
  Tooltip,
  Stack,
  useMantineColorScheme,
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
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const [mobileOpened, setMobileOpened] = useState(false);
  const [activeTab, setActiveTab] = useState<'downloads' | 'library' | 'settings'>('downloads');
  const { sidebarCollapsed, toggleSidebar, setTheme } = useSettingsStore();

  const dark = colorScheme === 'dark';

  const navItems = [
    { id: 'downloads' as const, label: 'Downloads', icon: IconDownload },
    { id: 'library' as const, label: 'Library', icon: IconVideo },
    { id: 'settings' as const, label: 'Settings', icon: IconSettings },
  ];

  const handleThemeToggle = () => {
    const newScheme = dark ? 'light' : 'dark';
    setColorScheme(newScheme);
    setTheme(newScheme as any);
  };

  return (
    <AppShell
      header={{ height: 60 }}
      navbar={{
        width: sidebarCollapsed ? 80 : 200,
        breakpoint: 'sm',
        collapsed: { mobile: !mobileOpened },
      }}
      footer={{ height: 40 }}
      padding="md"
    >
      <AppShell.Header>
        <Group h="100%" px="md" justify="space-between">
          <Group>
            <Burger
              opened={mobileOpened}
              onClick={() => setMobileOpened((o) => !o)}
              hiddenFrom="sm"
              size="sm"
              color={dark ? '#c1c2c5' : '#495057'}
            />
            <img src="/logo.svg" alt="YTed" style={{ height: 32, width: 32 }} />
            <Text fw={700} size="lg" c={dark ? '#fff' : '#000'}>YTed</Text>
            
            <Tooltip label={sidebarCollapsed ? "Expand sidebar" : "Collapse sidebar"}>
              <ActionIcon 
                variant="subtle" 
                onClick={toggleSidebar}
                color="gray"
                ml="sm"
              >
                <IconMenu2 size={18} />
              </ActionIcon>
            </Tooltip>
          </Group>
          
          <Tooltip label={dark ? "Switch to light mode" : "Switch to dark mode"}>
            <ActionIcon 
              variant="subtle" 
              onClick={handleThemeToggle}
              color="gray"
            >
              {dark ? <IconSun size={20} /> : <IconMoon size={20} />}
            </ActionIcon>
          </Tooltip>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="xs">
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
                  setMobileOpened(false);
                }}
                w={sidebarCollapsed ? 60 : '100%'}
                h={44}
                style={{
                  justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                  padding: sidebarCollapsed ? 0 : '0 12px',
                  borderRadius: 8,
                }}
                title={item.label}
              >
                <Group gap={12} wrap="nowrap">
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
      </AppShell.Navbar>

      <AppShell.Main bg={dark ? '#1a1b1e' : '#f8f9fa'}>
        {activeTab === 'downloads' && <DownloadPage />}
        {activeTab === 'library' && <LibraryPage />}
        {activeTab === 'settings' && <SettingsPage />}
      </AppShell.Main>

      <AppShell.Footer>
        <Group h="100%" px="md" justify="space-between">
          <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
            YTed v1.0.0
          </Text>
          <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
            Ready
          </Text>
        </Group>
      </AppShell.Footer>
    </AppShell>
  );
}

export default App;
