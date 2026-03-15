import { createTheme, MantineColorsTuple } from '@mantine/core';

// YTed primary red color scale
const ytedRed: MantineColorsTuple = [
  '#ffe5e5',
  '#ffc2c2',
  '#ff9e9e',
  '#ff7a7a',
  '#ff5555',
  '#ff3333',
  '#ff0000', // Primary
  '#e60000',
  '#cc0000',
  '#b30000',
];

export const theme = createTheme({
  primaryColor: 'yted',
  colors: {
    yted: ytedRed,
  },
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  fontFamilyMonospace: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
  headings: {
    fontFamily: 'inherit',
    fontWeight: '700',
  },
  radius: {
    xs: '4px',
    sm: '8px',
    md: '12px',
    lg: '16px',
    xl: '24px',
  },
  shadows: {
    xs: '0 1px 2px rgba(0, 0, 0, 0.1)',
    sm: '0 2px 4px rgba(0, 0, 0, 0.1)',
    md: '0 4px 8px rgba(0, 0, 0, 0.12)',
    lg: '0 8px 16px rgba(0, 0, 0, 0.14)',
    xl: '0 12px 24px rgba(0, 0, 0, 0.16)',
  },
  other: {
    appName: 'YTed',
  },
});

export type YTedTheme = typeof theme;
