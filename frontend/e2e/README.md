# E2E Testing with Playwright

This directory contains end-to-end tests for the YTed application using Playwright.

## Setup

```bash
# Install dependencies (already done during npm install)
npx playwright install

# Install browser dependencies (Linux only)
sudo npx playwright install-deps
```

## Running Tests

```bash
# Run all E2E tests
npm run test:e2e

# Run tests with UI mode
npm run test:e2e:ui

# Run tests in headed mode (see browser)
npm run test:e2e:headed

# Run specific test file
npx playwright test e2e/library.spec.ts
```

## Test Structure

- `e2e/*.spec.ts` - Test files for different features
- `e2e/utils/wails-mock.ts` - Utilities for mocking Wails runtime
- `e2e/fixtures/test-data.ts` - Reusable test data

## Mocking Wails

Since this is a Wails application, we mock the Wails runtime for E2E testing:

```typescript
import { createWailsMock, defaultMocks } from './utils/wails-mock';

test.beforeEach(async ({ page }) => {
  await createWailsMock(page, defaultMocks);
  await page.goto('/');
});
```

## Adding Data Test IDs

For reliable E2E tests, add `data-testid` attributes to components:

```tsx
<Button data-testid="add-download-button">Add Download</Button>
```

## CI/CD Integration

Tests run automatically in CI with:
- Chromium and Firefox browsers
- Retries on failure
- Screenshots on failure
- HTML report generation
