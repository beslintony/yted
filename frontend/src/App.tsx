import { useState } from 'react';
import {
  AppShell,
  Burger,
  Group,
  Tooltip,
  ActionIcon,
  Stack,
  useMantineColorScheme,
  Drawer,
} from '@mantine/core';
import {
  IconDownload,
  IconVideo,
  IconSettings,
  IconSun,
  IconMoon,
  IconFileText,
} from '@tabler/icons-react';
import { useSettingsStore } from './stores';
import { DownloadPage } from './pages/DownloadPage';
import { LibraryPage } from './pages/LibraryPage';
import { SettingsPage } from './pages/SettingsPage';
import { LoggerViewer } from './components/LoggerViewer';

function App() {
  const { colorScheme, setColorScheme } = useMantineColorScheme();
  const [mobileOpened, setMobileOpened] = useState(false);
  const [activeTab, setActiveTab] = useState<'downloads' | 'library' | 'settings'>('downloads');
  const [loggerOpened, setLoggerOpened] = useState(false);
  const { sidebarCollapsed, toggleSidebar } = useSettingsStore();

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
            {/* Desktop hamburger - toggles sidebar */}
            <Burger
              opened={!sidebarCollapsed}
              onClick={toggleSidebar}
              size="sm"
              color={dark ? '#c1c2c5' : '#495057'}
              aria-label="Toggle sidebar"
            />
            
            {/* Mobile hamburger - toggles mobile menu */}
            <Burger
              opened={mobileOpened}
              onClick={() => setMobileOpened((o) => !o)}
              hiddenFrom="sm"
              size="sm"
              color={dark ? '#c1c2c5' : '#495057'}
              aria-label="Toggle mobile menu"
            />
            
            {/* Logo */}
            <Tooltip label="YTed v1.0.0">
              <img 
                src="/logo.svg" 
                alt="YTed" 
                style={{ height: 36, width: 36, cursor: 'pointer' }}
              />
            </Tooltip>
          </Group>
          
          <Group gap="xs">
            <Tooltip label="View logs">
              <ActionIcon 
                variant="subtle" 
                onClick={() => setLoggerOpened(true)}
                color="gray"
                size="lg"
              >
                <IconFileText size={22} />
              </ActionIcon>
            </Tooltip>
            <Tooltip label={dark ? "Switch to light mode" : "Switch to dark mode"}>
              <ActionIcon 
                variant="subtle" 
                onClick={handleThemeToggle}
                color="gray"
                size="lg"
              >
                {dark ? <IconSun size={22} /> : <IconMoon size={22} />}
              </ActionIcon>
            </Tooltip>
          </Group>
        </Group>
      </AppShell.Header>

      <AppShell.Navbar p="md">
        <Stack gap="xs">
          {navItems.map((item) => {
            const Icon = item.icon;
            const isActive = activeTab === item.id;
            return (
              <Tooltip 
                key={item.id} 
                label={item.label} 
                position="right" 
                disabled={!sidebarCollapsed}
              >
                <ActionIcon
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
                >
                  <Group gap={12} wrap="nowrap">
                    <Icon size={20} />
                    {!sidebarCollapsed && (
                      <span style={{ fontSize: 14, fontWeight: 500 }}>
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
        <Group h="100%" px="md" justify="space-between">
          <span style={{ fontSize: 12, color: dark ? '#909296' : '#868e96' }}>
            YTed v1.0.0
          </span>
          <span style={{ fontSize: 12, color: dark ? '#909296' : '#868e96' }}>
            Ready
          </span>
        </Group>
      </AppShell.Footer>
      
      {/* Logger Drawer */}
      <Drawer
        opened={loggerOpened}
        onClose={() => setLoggerOpened(false)}
        title="Application Logs"
        position="right"
        size="lg"
        padding="md"
      >
        <LoggerViewer />
      </Drawer>
    </AppShell>
  );
}

export default App;
