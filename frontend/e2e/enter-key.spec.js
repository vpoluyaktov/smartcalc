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
  // This test suite verifies the Enter key behavior in SmartCalc.
  // The Enter key has context-sensitive behavior: it evaluates expressions,
  // adds = signs when missing, creates new lines, and handles special cases
  // like comments and inline comments.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should add = and evaluate when pressing Enter on expression without =', async ({ page }) => {
    // When pressing Enter on an expression without =, SmartCalc should:
    // 1. Automatically append = to the expression
    // 2. Evaluate and show the result
    // 3. Create a new line below
    // This provides a convenient shortcut for quick calculations.
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
    // When pressing Enter on a line that already has =, SmartCalc should:
    // 1. Evaluate the expression and show the result after =
    // 2. Create a new line below for the next calculation
    await typeInEditor(page, '10 * 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('50');
    
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(2);
  });

  test('should split line when pressing Enter before = sign', async ({ page }) => {
    // When the cursor is in the middle of a line (not at the end),
    // pressing Enter should split the line at the cursor position.
    // This is standard text editor behavior for multi-line editing.
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
    // On an empty line, Enter should simply create a new empty line.
    // This allows users to add spacing between calculations.
    await pressEnter(page);
    await pressEnter(page);
    
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThanOrEqual(2);
  });

  test('should insert newline on comment line', async ({ page }) => {
    // On a comment line (starting with #), Enter should create a new line
    // without attempting to evaluate. Comments are for documentation only.
    await typeInEditor(page, '# This is a comment');
    const initialCount = await getLineCount(page);
    
    await pressEnter(page);
    
    const newCount = await getLineCount(page);
    expect(newCount).toBe(initialCount + 1);
  });

  test('should handle Enter on last line of document', async ({ page }) => {
    // Verifies that Enter works correctly on the last line of the document.
    // This was a bug fix - previously Enter didn't work on the last line
    // because there was no next line to move to.
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
    // When a line has an inline comment (# after =), the result should be
    // inserted between = and #, preserving the user's note.
    // Example: "5 + 5 = # my sum" becomes "5 + 5 = 10 # my sum"
    await typeInEditor(page, '5 + 5 = # my sum');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('10');
    expect(line1).toContain('# my sum');
  });

  test('should handle multiple Enter presses creating multiple lines', async ({ page }) => {
    // Verifies that pressing Enter multiple times creates multiple empty lines.
    // This is useful for organizing calculations into sections.
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
    // Verifies that Enter correctly evaluates currency expressions.
    // The $ symbol should be preserved in the result.
    await typeInEditor(page, '$100 + $50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('$');
    expect(line1).toContain('150');
  });

  test('should handle percentage expressions', async ({ page }) => {
    // Verifies that Enter correctly evaluates percentage expressions.
    // "200 + 10%" means 200 + 10% of 200 = 220.
    await typeInEditor(page, '200 + 10% =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('220');
  });
});
