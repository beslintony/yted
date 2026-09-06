import { describe, expect, it } from 'vitest';

import {
  ActionsSection,
  CacheSection,
  DownloadsSection,
  FfmpegSection,
  LoggingSection,
  PlayerSection,
  PresetsSection,
  UiSection,
} from './index';

const MEMO = Symbol.for('react.memo');

describe('settings sections memo', () => {
  it.each([
    ['ActionsSection', ActionsSection],
    ['CacheSection', CacheSection],
    ['DownloadsSection', DownloadsSection],
    ['FfmpegSection', FfmpegSection],
    ['LoggingSection', LoggingSection],
    ['PlayerSection', PlayerSection],
    ['PresetsSection', PresetsSection],
    ['UiSection', UiSection],
  ])('%s is wrapped in React.memo', (_name, Section) => {
    expect((Section as unknown as { $$typeof: symbol }).$$typeof).toBe(MEMO);
  });

  // Render-skip behavior (that memo + stable SettingsPage callbacks actually
  // skip commits per keystroke) is covered by render-count assertions in
  // src/pages/SettingsPage.memo.test.tsx, which drives the real SettingsPage
  // with memoized capturing stubs.
});
