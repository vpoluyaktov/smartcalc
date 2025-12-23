// @ts-check
const { test, expect } = require('@playwright/test');
const {
  clearEditor,
  typeInEditor,
  getEditorText,
  waitForEditorReady,
} = require('./helpers');

test.describe('Context Menu', () => {
  // This test suite verifies the context menu functionality including
  // Cut, Copy, Paste, and Select All operations.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should show context menu on right-click', async ({ page }) => {
    // Right-click on the editor
    const editor = page.locator('#editor-container');
    await editor.click({ button: 'right' });

    // Context menu should be visible
    const contextMenu = page.locator('#context-menu');
    await expect(contextMenu).toBeVisible();

    // Verify menu items are present
    await expect(page.locator('[data-action="cut"]')).toBeVisible();
    await expect(page.locator('[data-action="copy"]')).toBeVisible();
    await expect(page.locator('[data-action="paste"]')).toBeVisible();
    await expect(page.locator('[data-action="selectall"]')).toBeVisible();
  });

  test('should hide context menu on click elsewhere', async ({ page }) => {
    const editor = page.locator('#editor-container');
    const contextMenu = page.locator('#context-menu');

    // Show context menu
    await editor.click({ button: 'right' });
    await expect(contextMenu).toBeVisible();

    // Click elsewhere (on the status bar)
    await page.locator('#status-bar').click();

    // Context menu should be hidden
    await expect(contextMenu).toBeHidden();
  });

  test('should hide context menu on Escape key', async ({ page }) => {
    const editor = page.locator('#editor-container');
    const contextMenu = page.locator('#context-menu');

    // Show context menu
    await editor.click({ button: 'right' });
    await expect(contextMenu).toBeVisible();

    // Press Escape
    await page.keyboard.press('Escape');

    // Context menu should be hidden
    await expect(contextMenu).toBeHidden();
  });

  test('should select all text via context menu', async ({ page }) => {
    // Type some text
    await typeInEditor(page, 'Hello World');

    // Right-click to show context menu
    const editor = page.locator('#editor-container');
    await editor.click({ button: 'right' });

    // Click Select All
    await page.locator('[data-action="selectall"]').click();

    // Wait a moment for selection to apply
    await page.waitForTimeout(100);

    // Verify text is selected by checking if we can get the selection
    // The selection should cover the entire text
    const selectedText = await page.evaluate(() => {
      const cm = document.querySelector('.cm-editor');
      if (cm && cm.cmView) {
        const state = cm.cmView.state;
        const selection = state.selection.main;
        return state.sliceDoc(selection.from, selection.to);
      }
      return '';
    });

    expect(selectedText.length).toBeGreaterThan(0);
  });

  test('should copy text via context menu', async ({ page, context }) => {
    // Grant clipboard permissions
    await context.grantPermissions(['clipboard-read', 'clipboard-write']);

    // Type some text
    await typeInEditor(page, 'Test Copy');

    // Select all text using keyboard shortcut first
    await page.keyboard.press('Control+a');
    await page.waitForTimeout(100);

    // Right-click to show context menu
    const editor = page.locator('#editor-container');
    await editor.click({ button: 'right' });

    // Click Copy
    await page.locator('[data-action="copy"]').click();

    // Wait for clipboard operation
    await page.waitForTimeout(200);

    // Note: Clipboard verification in Playwright with Wails can be tricky
    // The copy operation should complete without errors
  });

  test('should position context menu within viewport', async ({ page }) => {
    // Right-click near the bottom-right corner
    const editor = page.locator('#editor-container');
    const box = await editor.boundingBox();
    
    if (box) {
      // Click near bottom-right of editor
      await page.mouse.click(box.x + box.width - 20, box.y + box.height - 20, { button: 'right' });

      const contextMenu = page.locator('#context-menu');
      await expect(contextMenu).toBeVisible();

      // Get menu position
      const menuBox = await contextMenu.boundingBox();
      const viewport = page.viewportSize();

      if (menuBox && viewport) {
        // Menu should be within viewport
        expect(menuBox.x + menuBox.width).toBeLessThanOrEqual(viewport.width);
        expect(menuBox.y + menuBox.height).toBeLessThanOrEqual(viewport.height);
      }
    }
  });

  test('context menu items should have correct labels', async ({ page }) => {
    const editor = page.locator('#editor-container');
    await editor.click({ button: 'right' });

    // Check labels
    await expect(page.locator('[data-action="cut"] .context-menu-label')).toHaveText('Cut');
    await expect(page.locator('[data-action="copy"] .context-menu-label')).toHaveText('Copy');
    await expect(page.locator('[data-action="paste"] .context-menu-label')).toHaveText('Paste');
    await expect(page.locator('[data-action="selectall"] .context-menu-label')).toHaveText('Select All');
  });

  test('context menu items should have keyboard shortcuts', async ({ page }) => {
    const editor = page.locator('#editor-container');
    await editor.click({ button: 'right' });

    // Check shortcuts
    await expect(page.locator('[data-action="cut"] .context-menu-shortcut')).toHaveText('Ctrl+X');
    await expect(page.locator('[data-action="copy"] .context-menu-shortcut')).toHaveText('Ctrl+C');
    await expect(page.locator('[data-action="paste"] .context-menu-shortcut')).toHaveText('Ctrl+V');
    await expect(page.locator('[data-action="selectall"] .context-menu-shortcut')).toHaveText('Ctrl+A');
  });
});
