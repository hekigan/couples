# E2E Testing with Playwright

**Status:** ✅ **Infrastructure Complete - Tests Need Debugging**
**Date:** 2025-11-10
**Framework:** Playwright 1.40+
**Scope:** Quick-start smoke tests for critical game flows

---

## Overview

This project uses **Playwright** for end-to-end testing of the Couple Card Game application. The focus is on testing the **HTMX-driven frontend** with real-time **Server-Sent Events (SSE)** integration.

### What's Tested
- Anonymous user creation and authentication
- Room creation and joining
- Two-player game synchronization via SSE
- HTMX partial page updates
- Real-time multi-player interactions

### What's NOT Tested (Yet)
- OAuth login flows (Google/Facebook/GitHub)
- Friend system workflows
- Mobile responsive layouts
- Error scenarios and edge cases
- Performance benchmarks

---

## Quick Start

### 1. Setup (One-Time)

```bash
# Install Playwright and browsers
make test-e2e-setup
```

This will:
- Install npm dependencies (`@playwright/test`)
- Download Chromium browser (~220MB)

### 2. Run Tests

```bash
# Run all E2E tests (headless)
make test-e2e

# Run tests with visible browser (debugging)
make test-e2e-headed

# Open Playwright UI (interactive test runner)
make test-e2e-ui

# Debug mode with step-by-step execution
make test-e2e-debug

# View test report
make test-e2e-report
```

### 3. Prerequisites

**Before running tests, ensure:**
- ✅ Server is running: `make run` (or tests will auto-start it)
- ✅ Database is accessible (uses production DB, not test DB)
- ✅ Port 8188 is available

---

## Project Structure

```
tests/
├── e2e/                           # E2E test files
│   └── game/
│       ├── room-creation.spec.ts  # Smoke test: create room
│       └── two-player-game.spec.ts # Complex: full game flow
├── helpers/                       # Test utilities
│   ├── sse.helper.ts             # SSE event testing
│   └── test.helper.ts            # General test helpers
└── fixtures/                      # Test data (future)
    └── test-users.ts             # User creation utilities

playwright.config.ts               # Playwright configuration
package.json                       # npm dependencies
```

---

## Test Files

### 1. Room Creation Smoke Test
**File:** `tests/e2e/game/room-creation.spec.ts`
**Duration:** ~10-15 seconds
**Tests:**
- Anonymous user can create a room
- Room lobby shows category selection

**What It Does:**
```typescript
1. Create anonymous user
2. Navigate to /game/create-room
3. Fill in room name
4. Submit form
5. Verify redirect to /game/room/{id}
6. Verify room name displayed
7. Verify category checkboxes exist
```

### 2. Two-Player Game Flow
**File:** `tests/e2e/game/two-player-game.spec.ts`
**Duration:** ~20-30 seconds
**Tests:**
- Two players can create, join, and start a game
- SSE events synchronize both players
- Real-time updates work correctly

**What It Does:**
```typescript
1. Create two browser contexts (Player 1 & Player 2)
2. Player 1: Create anonymous account + room
3. Player 1: Select categories
4. Player 2: Create anonymous account + join room
5. Verify Player 1 sees Player 2 joined (SSE event)
6. Player 2: Click "Ready" button
7. Player 1: Start game
8. Both players: Verify redirect to /game/play/{id}
9. Verify game content loads
10. Verify turn indicators visible
```

---

## Helper Functions

### SSE Helper (`tests/helpers/sse.helper.ts`)

```typescript
import { SSEHelper } from '../helpers/sse.helper';

// Create SSE helper
const sse = new SSEHelper(page);

// Wait for SSE connection
await sse.waitForSSEConnection(roomId);

// Wait for specific SSE event
const received = await sse.waitForSSEEvent('player_joined', 5000);

// Wait for HTMX swap after SSE event
await sse.waitForHTMXSwap('#player-info', 'Player2 joined');

// Enable logging (debugging)
sse.enableSSELogging();
```

### Test Helper (`tests/helpers/test.helper.ts`)

```typescript
import { createAnonymousUser, generateRoomName, getRoomIdFromURL } from '../helpers/test.helper';

// Create anonymous user
const username = await createAnonymousUser(page);

// Generate unique room name
const roomName = generateRoomName();

// Extract room ID from URL
const roomId = getRoomIdFromURL(page); // from /game/room/abc-123

// Wait for HTMX request
await waitForHTMXRequest(page, '#game-content');

// Wait for text to appear
await waitForText(page, '.status', 'Game Started');

// Take screenshot (debugging)
await takeScreenshot(page, 'error-state');
```

---

## Configuration

### Playwright Config (`playwright.config.ts`)

```typescript
{
  testDir: './tests/e2e',
  fullyParallel: false,        // Sequential execution
  workers: 1,                   // Single worker (DB isolation)
  timeout: 30000,               // 30s per test
  retries: 0,                   // No retries (for now)

  use: {
    baseURL: 'http://localhost:8188',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    trace: 'on-first-retry'
  },

  projects: [
    { name: 'chromium' }        // Chromium only (quick start)
  ],

  webServer: {
    command: 'make run',        // Auto-start server
    url: 'http://localhost:8188/health',
    reuseExistingServer: true
  }
}
```

---

## Writing Tests

### Basic Test Template

```typescript
import { test, expect } from '@playwright/test';
import { createAnonymousUser } from '../../helpers/test.helper';

test.describe('My Feature', () => {
  test('should do something', async ({ page }) => {
    // Setup
    const username = await createAnonymousUser(page);

    // Action
    await page.click('a[href="/some-page"]');

    // Assertion
    await expect(page).toHaveURL('/some-page');
    await expect(page.locator('h1')).toContainText('Expected Text');
  });
});
```

### Two-Player Test Template

```typescript
import { test, Browser } from '@playwright/test';

test('two players test', async ({ browser }: { browser: Browser }) => {
  const context1 = await browser.newContext();
  const context2 = await browser.newContext();

  const player1 = await context1.newPage();
  const player2 = await context2.newPage();

  try {
    // Player 1 actions
    await createAnonymousUser(player1, 'Player1');

    // Player 2 actions
    await createAnonymousUser(player2, 'Player2');

    // Verify interaction between players

  } finally {
    await context1.close();
    await context2.close();
  }
});
```

### Testing SSE Events

```typescript
import { SSEHelper } from '../../helpers/sse.helper';

test('SSE real-time update', async ({ page }) => {
  const sse = new SSEHelper(page);

  // Navigate to page with SSE
  await page.goto('/game/room/abc-123');

  // Wait for SSE connection
  await sse.waitForSSEConnection('abc-123');

  // Trigger action that causes SSE event
  await page.click('#some-action-button');

  // Wait for SSE event
  const received = await sse.waitForSSEEvent('some_event', 10000);
  expect(received).toBe(true);

  // Verify HTMX updated the DOM
  await sse.waitForHTMXSwap('#target-element', 'Updated Text');
});
```

---

## Debugging Tests

### 1. View Screenshots and Videos

After test failure:
```bash
ls test-results/
# Screenshots: test-failed-1.png
# Videos: video.webm
```

### 2. Use Playwright UI

```bash
make test-e2e-ui
```

This opens an interactive UI where you can:
- See test execution step-by-step
- Time-travel through test execution
- Inspect DOM at each step
- Re-run specific tests

### 3. Debug Mode

```bash
make test-e2e-debug
```

This opens the Playwright Inspector where you can:
- Step through test line-by-line
- Pause execution
- Inspect page state
- Run commands in console

### 4. Add Console Logs

```typescript
test('my test', async ({ page }) => {
  console.log('Current URL:', page.url());

  const element = await page.locator('h1').textContent();
  console.log('H1 text:', element);

  // Enable SSE logging
  const sse = new SSEHelper(page);
  sse.enableSSELogging(); // Logs all SSE events to console
});
```

### 5. Take Screenshots Manually

```typescript
import { takeScreenshot } from '../../helpers/test.helper';

test('my test', async ({ page }) => {
  // ... test code ...

  // Take screenshot at specific point
  await takeScreenshot(page, 'before-action');

  await page.click('#action-button');

  await takeScreenshot(page, 'after-action');
});
```

---

## Known Issues & Workarounds

### Issue 1: Tests Failing - Anonymous User Creation

**Status:** 🚧 **IN PROGRESS**

**Problem:** The `createAnonymousUser()` helper is not correctly handling the anonymous login flow.

**Current Behavior:**
- Tests fail with timeout when trying to create anonymous user
- Helper expects `/setup-username` redirect but gets `/` instead

**Workaround:**
1. Manually inspect the actual anonymous user creation flow
2. Update `tests/helpers/test.helper.ts` to match the real flow
3. Check if authentication requires clicking a "Guest" button on home page
4. Verify session cookies are being set correctly

**Next Steps:**
- Use `make test-e2e-ui` to visually debug the login flow
- Check the actual HTML structure of the home page and login pages
- Update helper to match the real authentication flow

### Issue 2: Form Fields Not Found

**Problem:** Tests can't find `input[name="room_name"]` on create-room page.

**Possible Causes:**
1. User not properly authenticated (redirected away from form)
2. Form structure different than expected
3. Page loading too slowly

**Debug:**
```typescript
// Add logging to see actual page state
await page.screenshot({ path: 'debug-create-room.png' });
const html = await page.content();
console.log('Page HTML:', html.substring(0, 500));

// Check for form fields
const formFields = await page.locator('input').all();
console.log('Found form fields:', formFields.length);
for (const field of formFields) {
  const name = await field.getAttribute('name');
  console.log('Field name:', name);
}
```

---

## Future Enhancements

### Priority 1 (Near-Term)
- [ ] Fix anonymous user creation helper
- [ ] Add more assertions to existing tests
- [ ] Test actual gameplay (draw question, submit answer, next turn)
- [ ] Add test for game completion flow

### Priority 2 (Medium-Term)
- [ ] OAuth login mocking/testing
- [ ] Friend system E2E tests
- [ ] Error handling tests (invalid inputs, edge cases)
- [ ] Mobile responsive tests (different viewports)
- [ ] Test SSE reconnection logic

### Priority 3 (Long-Term)
- [ ] CI/CD integration (GitHub Actions)
- [ ] Parallel test execution (with DB isolation)
- [ ] Visual regression testing
- [ ] Performance benchmarking
- [ ] Cross-browser testing (Firefox, WebKit)
- [ ] Accessibility testing

---

## Makefile Commands Reference

```bash
# Setup
make test-e2e-setup          # One-time setup (install Playwright)

# Run tests
make test-e2e                # Run all tests (headless)
make test-e2e-headed         # Run with visible browser
make test-e2e-ui             # Interactive UI mode
make test-e2e-debug          # Step-by-step debugging

# Reports
make test-e2e-report         # View HTML test report
```

---

## Additional Resources

### Playwright Documentation
- [Playwright Docs](https://playwright.dev/)
- [Best Practices](https://playwright.dev/docs/best-practices)
- [Debugging Guide](https://playwright.dev/docs/debug)
- [API Reference](https://playwright.dev/docs/api/class-playwright)

### HTMX Testing Resources
- [Testing HTMX Applications](https://htmx.org/essays/testing/)
- [SSE Extension Docs](https://htmx.org/extensions/server-sent-events/)

### Project-Specific
- `HTMX_PROJECT_COMPLETE.md` - HTMX refactoring details
- `HTMX_INTEGRATION_VALIDATION.md` - SSE-HTMX event mapping
- `CLAUDE.md` - Project architecture overview

---

## Troubleshooting

### Tests Won't Run

**Check:** Is the server running?
```bash
curl http://localhost:8188/health
# Should return: {"status":"healthy"}
```

**Fix:** Start the server
```bash
make run
```

### Tests Timeout

**Check:** Increase timeout in `playwright.config.ts`
```typescript
timeout: 60000, // 60 seconds
```

**Or:** Increase timeout for specific test
```typescript
test('my slow test', async ({ page }) => {
  test.setTimeout(60000); // 60 seconds for this test only
});
```

### Can't Find Elements

**Debug:** Print page HTML
```typescript
const html = await page.content();
console.log(html);

// Or take screenshot
await page.screenshot({ path: 'debug.png', fullPage: true });
```

### SSE Events Not Working

**Check:** Is SSE connection established?
```typescript
const sse = new SSEHelper(page);
sse.enableSSELogging(); // See all SSE events in console

await page.goto('/game/room/abc-123');
await sse.waitForSSEConnection('abc-123', 10000);
```

---

## Summary

✅ **Playwright infrastructure is complete and ready to use**
🚧 **Tests need debugging to match actual app behavior**
📝 **Documentation complete with examples and troubleshooting**
🎯 **Next step: Debug and fix `createAnonymousUser()` helper**

The E2E testing foundation is solid. Once the anonymous user creation flow is corrected, the tests should pass and provide a reliable safety net for future development.

---

*Last updated: 2025-11-10*
*Author: Claude Code*
