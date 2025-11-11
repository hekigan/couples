import { test, expect, Browser } from '@playwright/test';
import { createAnonymousUser, generateRoomName, getRoomIdFromURL, waitForText } from '../../helpers/test.helper';
import { SSEHelper } from '../../helpers/sse.helper';

test.describe('Two-Player Game Flow', () => {
  test('two players can create, join, and play a game', async ({ browser }: { browser: Browser }) => {
    // ===== SETUP: Create two browser contexts (two different users) =====
    const context1 = await browser.newContext();
    const context2 = await browser.newContext();

    const player1Page = await context1.newPage();
    const player2Page = await context2.newPage();

    const sse1 = new SSEHelper(player1Page);
    const sse2 = new SSEHelper(player2Page);

    try {
      // ===== STEP 1: Player 1 creates anonymous account and room =====
      console.log('Step 1: Player 1 creates account and room');
      const player1Name = await createAnonymousUser(player1Page, `Player1_${Date.now()}`);
      await player1Page.getByTestId('create-room-link').click();
      await expect(player1Page).toHaveURL('/game/create-room');

      const roomName = generateRoomName();
      await player1Page.getByTestId('room-name-input').fill(roomName);
      await player1Page.getByTestId('create-room-submit').click();

      await expect(player1Page).toHaveURL(/\/game\/room\//);
      const roomId = getRoomIdFromURL(player1Page);
      console.log(`Player 1 created room: ${roomId}`);

      // Verify Player 1 is shown as owner (username may be truncated)
      await expect(player1Page.getByTestId('owner-name')).toContainText(player1Name.substring(0, 15));
      console.log('Player 1 on room page');

      // ===== STEP 2: Player 1 waits for Player 2 (categories will be shown after) =====
      console.log('Step 2: Player 1 waiting for Player 2...');

      // ===== STEP 3: Player 2 creates account and joins room =====
      console.log('Step 3: Player 2 joins room');
      const player2Name = await createAnonymousUser(player2Page, `Player2_${Date.now()}`);

      // Navigate to the same room
      await player2Page.goto(`/game/room/${roomId}`);
      await expect(player2Page).toHaveURL(/\/game\/room\//);

      // Wait for the page to load (not networkidle because SSE keeps connection open)
      await player2Page.waitForLoadState('domcontentloaded');
      // Verify Player 2 sees the room
      await expect(player2Page.getByTestId('room-id-input')).toBeVisible();
      console.log('Player 2 page loaded');

      // ===== STEP 4: Verify Player 1 sees Player 2 joined (SSE event) =====
      console.log('Step 4: Verify Player 1 sees Player 2 joined');
      await expect(player1Page.getByTestId('guest-name')).toContainText(player2Name.substring(0, 15), { timeout: 10000 });
      console.log('Player 1 sees Player 2 joined');

      // ===== STEP 5: Player 2 clicks "Ready" button (if visible) =====
      console.log('Step 5: Check if Player 2 needs to click Ready');
      const readyButton = player2Page.locator('button').filter({ hasText: /ready|prêt/i });
      if (await readyButton.count() > 0) {
        await readyButton.click();
        await player2Page.waitForTimeout(500);
        console.log('Player 2 clicked Ready');
      }

      // ===== STEP 6: Player 1 starts the game =====
      console.log('Step 6: Player 1 starts game');
      const startButton = player1Page.locator('button').filter({ hasText: /start|démarrer/i });

      // Wait up to 10 seconds for button to be enabled
      await startButton.waitFor({ state: 'visible', timeout: 10000 });
      await startButton.click();

      // ===== STEP 7: Both players should be redirected to play page =====
      console.log('Step 7: Verify both players redirected to play page');
      await expect(player1Page).toHaveURL(/\/game\/play\//, { timeout: 10000 });
      await expect(player2Page).toHaveURL(/\/game\/play\//, { timeout: 10000 });

      console.log('✓ Both players on play page');

      // ===== STEP 8: Verify game content loads =====
      console.log('Step 8: Verify game content loads');
      await expect(player1Page.getByTestId('game-content')).toBeVisible({ timeout: 10000 });
      await expect(player2Page.getByTestId('game-content')).toBeVisible({ timeout: 10000 });

      console.log('✓ Game content visible for both players');

      // ===== STEP 9: Check game header and progress =====
      console.log('Step 9: Verify game header and progress');
      await expect(player1Page.getByTestId('game-header')).toBeVisible({ timeout: 5000 });
      await expect(player2Page.getByTestId('game-header')).toBeVisible({ timeout: 5000 });

      await expect(player1Page.getByTestId('game-progress')).toBeVisible();
      await expect(player2Page.getByTestId('game-progress')).toBeVisible();

      console.log('✓ Game UI elements visible');

      // ===== SUCCESS =====
      console.log('✅ Two-player game flow test PASSED');
      console.log(`   - Player 1: ${player1Name}`);
      console.log(`   - Player 2: ${player2Name}`);
      console.log(`   - Room ID: ${roomId}`);

    } catch (error) {
      // Take screenshots on failure
      await player1Page.screenshot({ path: `test-results/player1-error-${Date.now()}.png` });
      await player2Page.screenshot({ path: `test-results/player2-error-${Date.now()}.png` });
      throw error;
    } finally {
      // Cleanup
      await context1.close();
      await context2.close();
    }
  });
});
