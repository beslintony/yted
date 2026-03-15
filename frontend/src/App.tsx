import { useState, useEffect } from 'react';
import {
  AppShell,
  Navbar,
  Header,
  Footer,
  Text,
  Burger,
  useMantineColorScheme,
  Group,
  ActionIcon,
  Tooltip,
  Stack,
  Box,
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
  const { colorScheme, toggleColorScheme } = useMantineColorScheme();
  const [opened, setOpened] = useState(false);
  const [activeTab, setActiveTab] = useState<'downloads' | 'library' | 'settings'>('downloads');
  const { sidebarCollapsed, toggleSidebar, theme: userTheme, setTheme } = useSettingsStore();

  // Sync user theme preference with Mantine
  useEffect(() => {
    if (userTheme === 'dark' && colorScheme !== 'dark') {
      toggleColorScheme('dark');
    } else if (userTheme === 'light' && colorScheme !== 'light') {
      toggleColorScheme('light');
    }
  }, [userTheme]);

  const handleThemeToggle = () => {
    const newScheme = colorScheme === 'dark' ? 'light' : 'dark';
    toggleColorScheme(newScheme);
    setTheme(newScheme as any);
  };

  const navItems = [
    { id: 'downloads' as const, label: 'Downloads', icon: IconDownload },
    { id: 'library' as const, label: 'Library', icon: IconVideo },
    { id: 'settings' as const, label: 'Settings', icon: IconSettings },
  ];

  const dark = colorScheme === 'dark';

  return (
    <AppShell
      styles={{
        main: {
          background: dark ? '#1a1b1e' : '#f8f9fa',
        },
      }}
      navbarOffsetBreakpoint="sm"
      navbar={
        <Navbar 
          p="md" 
          hiddenBreakpoint="sm" 
          hidden={!opened} 
          width={{ sm: sidebarCollapsed ? 80 : 200 }}
          sx={{
            background: dark ? '#141517' : '#fff',
            borderRight: `1px solid ${dark ? '#2c2e33' : '#e9ecef'}`,
          }}
        >
          <Navbar.Section grow>
            <Stack spacing="xs">
              {navItems.map((item) => {
                const Icon = item.icon;
                const isActive = activeTab === item.id;
                return (
                  <ActionIcon
                    key={item.id}
                    variant={isActive ? 'filled' : 'subtle'}
                    color={isActive ? 'red' : 'gray'}
                    onClick={() => {
                      setActiveTab(item.id);
                      setOpened(false);
                    }}
                    sx={{
                      width: sidebarCollapsed ? 60 : '100%',
                      height: 44,
                      justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
                      padding: sidebarCollapsed ? 0 : '0 12px',
                      borderRadius: 8,
                      backgroundColor: isActive ? (dark ? '#c92a2a' : '#fa5252') : 'transparent',
                      color: isActive ? '#fff' : (dark ? '#c1c2c5' : '#495057'),
                      '&:hover': {
                        backgroundColor: isActive 
                          ? (dark ? '#c92a2a' : '#fa5252') 
                          : (dark ? '#2c2e33' : '#f1f3f5'),
                      },
                    }}
                    title={item.label}
                  >
                    <Group spacing={12} noWrap>
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
        <Footer 
          height={40} 
          p="xs"
          sx={{
            background: dark ? '#141517' : '#fff',
            borderTop: `1px solid ${dark ? '#2c2e33' : '#e9ecef'}`,
          }}
        >
          <Group position="apart">
            <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
              YTed v1.0.0
            </Text>
            <Text size="xs" c={dark ? 'dimmed' : 'gray.6'}>
              Ready
            </Text>
          </Group>
        </Footer>
      }
      header={
        <Header 
          height={60} 
          p="sm"
          sx={{
            background: dark ? '#141517' : '#fff',
            borderBottom: `1px solid ${dark ? '#2c2e33' : '#e9ecef'}`,
          }}
        >
          <Box sx={{ display: 'flex', alignItems: 'center', height: '100%' }}>
            <Burger
              opened={opened}
              onClick={() => setOpened((o) => !o)}
              size="sm"
              color={dark ? '#c1c2c5' : '#495057'}
              mr="md"
              sx={{ display: 'none', '@media (max-width: 768px)': { display: 'block' } }}
            />
            
            <Group position="apart" sx={{ flex: 1 }}>
              <Group spacing="sm">
                <img
                  src="/logo.svg"
                  alt="YTed"
                  style={{ height: 32, width: 32 }}
                />
                <Text fw={700} size="lg" c={dark ? '#fff' : '#000'}>YTed</Text>
                
                {/* Move sidebar toggle here, near app name */}
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
              
              <Group spacing="xs">
                <Tooltip label={dark ? "Switch to light mode" : "Switch to dark mode"}>
                  <ActionIcon 
                    variant="subtle" 
                    onClick={handleThemeToggle}
                    color="gray"
                    sx={{ color: dark ? '#c1c2c5' : '#495057' }}
                  >
                    {dark ? <IconSun size={20} /> : <IconMoon size={20} />}
                  </ActionIcon>
                </Tooltip>
              </Group>
            </Group>
          </Box>
        </Header>
      }
    >
      <Box sx={{ padding: 'md' }}>
        {activeTab === 'downloads' && <DownloadPage />}
        {activeTab === 'library' && <LibraryPage />}
        {activeTab === 'settings' && <SettingsPage />}
      </Box>
    </AppShell>
  );
}

export default App;
