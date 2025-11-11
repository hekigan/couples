import { test, expect } from '@playwright/test';
import { createAnonymousUser, generateRoomName, getRoomIdFromURL } from '../../helpers/test.helper';

test.describe('Room Creation - Smoke Test', () => {
  test('anonymous user can create a room', async ({ page }) => {
    // Step 1: Create anonymous user
    const username = await createAnonymousUser(page);
    console.log(`Created anonymous user: ${username}`);

    // Step 2: Navigate to create room page
    await page.getByTestId('create-room-link').click();
    await expect(page).toHaveURL('/game/create-room');

    // Step 3: Fill in room form
    const roomName = generateRoomName();
    await page.getByTestId('room-name-input').fill(roomName);

    // Step 4: Submit form
    await page.getByTestId('create-room-submit').click();

    // Step 5: Verify redirect to room lobby
    await expect(page).toHaveURL(/\/game\/room\/[a-f0-9-]+/);

    // Step 6: Verify room ID is displayed
    const roomId = getRoomIdFromURL(page);
    await expect(page.getByTestId('room-id-input')).toHaveValue(roomId);
    console.log(`Room created successfully: ${roomId}`);

    // Step 7: Verify user is shown as owner (username may be truncated)
    await expect(page.getByTestId('owner-name')).toContainText(username.substring(0, 15));

    // Step 8: Verify room status badge exists
    await expect(page.getByTestId('room-status-badge')).toBeVisible();

    console.log('✓ Room creation smoke test passed');
  });

  test('room lobby shows category selection', async ({ page }) => {
    // Create user and room
    const username = await createAnonymousUser(page);
    await page.getByTestId('create-room-link').click();

    const roomName = generateRoomName();
    await page.getByTestId('room-name-input').fill(roomName);
    await page.getByTestId('create-room-submit').click();

    // Wait for room lobby to load
    await expect(page).toHaveURL(/\/game\/room\//);

    // Verify categories section exists (may be hidden initially, shown when guest joins)
    const categoriesSection = page.getByTestId('categories-section');
    await expect(categoriesSection).toBeAttached();

    console.log('✓ Category selection test passed - categories section exists');
  });
});
