// @ts-check
const { test, expect } = require('@playwright/test');
const {
  clearEditor,
  typeInEditor,
  pressEnter,
  goToLine,
  goToEndOfLine,
  waitForEvaluation,
  getLineText,
  getLineCount,
  waitForEditorReady,
} = require('./helpers');

test.describe('Enter Key Behavior', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should add = and evaluate when pressing Enter on expression without =', async ({ page }) => {
    await typeInEditor(page, '3 + 4');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('=');
    expect(line1).toContain('7');
    
    // Should also create a new line
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(2);
  });

  test('should evaluate and create new line when pressing Enter after =', async ({ page }) => {
    await typeInEditor(page, '10 * 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('50');
    
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(2);
  });

  test('should split line when pressing Enter before = sign', async ({ page }) => {
    await typeInEditor(page, '100 + 200 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Go back to line 1, position cursor before =
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    // Move to position before "+"
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowRight');
    await page.keyboard.press('ArrowRight'); // After "100 "
    
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Line should be split
    const line1 = await getLineText(page, 1);
    const line2 = await getLineText(page, 2);
    
    // First line should have "100 " or similar (before split point)
    expect(line1).toContain('100');
    // Second line should have the rest
    expect(line2).toContain('+');
  });

  test('should insert newline on empty line', async ({ page }) => {
    await pressEnter(page);
    await pressEnter(page);
    
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(2);
  });

  test('should insert newline on comment line', async ({ page }) => {
    await typeInEditor(page, '# This is a comment');
    const initialCount = await getLineCount(page);
    
    await pressEnter(page);
    
    const newCount = await getLineCount(page);
    expect(newCount).toBe(initialCount + 1);
  });

  test('should handle Enter on last line of document', async ({ page }) => {
    await typeInEditor(page, '1 + 1 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '2 + 2 =');
    
    // Get current line count
    const beforeCount = await getLineCount(page);
    
    // Press Enter on last line
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Should have added a new line
    const afterCount = await getLineCount(page);
    expect(afterCount).toBe(beforeCount + 1);
    
    // Last expression should be evaluated
    const line2 = await getLineText(page, 2);
    expect(line2).toContain('4');
  });

  test('should preserve inline comments when evaluating', async ({ page }) => {
    await typeInEditor(page, '5 + 5 = # my sum');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('10');
    expect(line1).toContain('# my sum');
  });

  test('should handle multiple Enter presses creating multiple lines', async ({ page }) => {
    await typeInEditor(page, '1 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await pressEnter(page);
    await pressEnter(page);
    await pressEnter(page);
    
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(4);
  });

  test('should evaluate currency expressions correctly', async ({ page }) => {
    await typeInEditor(page, '$100 + $50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('$');
    expect(line1).toContain('150');
  });

  test('should handle percentage expressions', async ({ page }) => {
    await typeInEditor(page, '200 + 10% =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('220');
  });
});
