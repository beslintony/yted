import { MantineThemeOverride } from '@mantine/core';

// YTed theme configuration for Mantine v6
export const theme: MantineThemeOverride = {
  primaryColor: 'red',
  primaryShade: 6,
  colors: {
    red: [
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
    ],
  },
  fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
  fontFamilyMonospace: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, monospace',
  headings: {
    fontFamily: 'inherit',
    fontWeight: 700,
  },
  defaultRadius: 'md',
  shadows: {
    xs: '0 1px 2px rgba(0, 0, 0, 0.1)',
    sm: '0 2px 4px rgba(0, 0, 0, 0.1)',
    md: '0 4px 8px rgba(0, 0, 0, 0.12)',
    lg: '0 8px 16px rgba(0, 0, 0, 0.14)',
    xl: '0 12px 24px rgba(0, 0, 0, 0.16)',
  },
};
