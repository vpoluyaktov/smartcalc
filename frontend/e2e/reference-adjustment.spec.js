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
  getEditorContent,
} = require('./helpers');

test.describe('Reference Adjustment', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should adjust references when inserting a line before them', async ({ page }) => {
    // Set up initial content:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state - line 2 should reference \1 and show result
    let line2 = await getLineText(page, 2);
    expect(line2).toContain('\\1');
    expect(line2).toContain('200');
    
    // Go to line 1 and insert a new line before it
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await waitForEvaluation(page);
    
    // Now line 2 should be "100 =" and line 3 should have \2 (adjusted from \1)
    const line3 = await getLineText(page, 3);
    expect(line3).toContain('\\2');
  });

  test('should adjust references when inserting empty line in middle', async ({ page }) => {
    // Set up:
    // Line 1: # comment
    // Line 2: 2 + 2 =
    // Line 3: \2 x 4 =
    await typeInEditor(page, '# comment');
    await pressEnter(page);
    await typeInEditor(page, '2 + 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 x 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify line 3 has \2 and result 16
    let line3 = await getLineText(page, 3);
    expect(line3).toContain('\\2');
    expect(line3).toContain('16');
    
    // Insert empty line after comment (line 1)
    await goToLine(page, 1);
    await goToEndOfLine(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Now line 4 should have \3 (adjusted from \2)
    const line4 = await getLineText(page, 4);
    expect(line4).toContain('\\3');
  });

  test('should adjust chained references when inserting line', async ({ page }) => {
    // Set up chain:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    // Line 3: \2 + 50 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 + 50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify chain works
    let line2 = await getLineText(page, 2);
    let line3 = await getLineText(page, 3);
    expect(line2).toContain('200');
    expect(line3).toContain('250');
    
    // Insert line at beginning
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await waitForEvaluation(page);
    
    // Check references shifted
    const newLine3 = await getLineText(page, 3);
    const newLine4 = await getLineText(page, 4);
    expect(newLine3).toContain('\\2'); // was \1
    expect(newLine4).toContain('\\3'); // was \2
  });

  test('should adjust references when deleting a line before them', async ({ page }) => {
    // Set up:
    // Line 1: (empty)
    // Line 2: 100 =
    // Line 3: \2 * 2 =
    await pressEnter(page);
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify line 3 has \2
    let line3 = await getLineText(page, 3);
    expect(line3).toContain('\\2');
    
    // Delete line 1 (empty line)
    await goToLine(page, 1);
    await page.keyboard.press('Control+Shift+k'); // Delete line
    await waitForEvaluation(page);
    
    // Now line 2 should have \1 (adjusted from \2)
    const newLine2 = await getLineText(page, 2);
    expect(newLine2).toContain('\\1');
  });

  test('should handle user scenario: insert empty line before refs', async ({ page }) => {
    // This is the exact scenario from the user's screenshot
    // Set up:
    // Line 1-3: comments
    // Line 4: empty
    // Line 5: empty (cursor here)
    // Line 6: 2 + 2 = 4
    // Line 7: empty
    // Line 8: \6 x 4 = 16
    // Line 9: empty
    // Line 10: \8 / 2 = 8
    
    await typeInEditor(page, '# Welcome to SmartCalc!');
    await pressEnter(page);
    await typeInEditor(page, '# Check out the Snippets menu');
    await pressEnter(page);
    await typeInEditor(page, '# Type an expression and press Enter');
    await pressEnter(page);
    await pressEnter(page); // Line 4 empty
    await pressEnter(page); // Line 5 empty
    
    await typeInEditor(page, '2 + 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await pressEnter(page); // Line 7 empty
    
    await typeInEditor(page, '\\6 x 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await pressEnter(page); // Line 9 empty
    
    await typeInEditor(page, '\\8 / 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial refs
    let line8 = await getLineText(page, 8);
    let line10 = await getLineText(page, 10);
    expect(line8).toContain('\\6');
    expect(line10).toContain('\\8');
    
    // Go to line 5 (empty) and press Enter to insert new line
    await goToLine(page, 5);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // After inserting, refs should shift:
    // \6 should become \7
    // \8 should become \9
    const newLine9 = await getLineText(page, 9);
    const newLine11 = await getLineText(page, 11);
    expect(newLine9).toContain('\\7');
    expect(newLine11).toContain('\\9');
  });

  test('should not adjust references before insertion point', async ({ page }) => {
    // Set up:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    // Line 3: 50 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Insert line after line 2 (after the reference)
    await goToLine(page, 2);
    await goToEndOfLine(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Line 2 should still have \1 (not adjusted because it's before insertion point)
    const line2 = await getLineText(page, 2);
    expect(line2).toContain('\\1');
  });

  test('should adjust references when inserting multiple empty lines before refs', async ({ page }) => {
    // Set up:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    // Line 3: \2 + 10 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 + 10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line2 = await getLineText(page, 2);
    let line3 = await getLineText(page, 3);
    expect(line2).toContain('\\1');
    expect(line2).toContain('200');
    expect(line3).toContain('\\2');
    expect(line3).toContain('210');
    
    // Insert 3 empty lines before line 1
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await waitForEvaluation(page);
    
    // Now refs should be shifted by 3:
    // Line 5 should have \4 (was \1)
    // Line 6 should have \5 (was \2)
    const line5 = await getLineText(page, 5);
    const line6 = await getLineText(page, 6);
    expect(line5).toContain('\\4');
    expect(line6).toContain('\\5');
  });

  test('should adjust references when inserting empty lines between refs', async ({ page }) => {
    // Set up:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    // Line 3: \2 + 10 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 + 10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line2 = await getLineText(page, 2);
    let line3 = await getLineText(page, 3);
    expect(line2).toContain('\\1');
    expect(line3).toContain('\\2');
    
    // Insert 2 empty lines after line 1 (before the refs)
    await goToLine(page, 1);
    await goToEndOfLine(page);
    await pressEnter(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // After inserting 2 lines after line 1:
    // - Line 1 stays at line 1 (100 =)
    // - Lines 2,3 are empty
    // - Line 4 has \3 (was \1, adjusted because source moved from 1 to 3... wait, source is still at 1)
    // Actually: refs pointing to line 1 stay as \1 because line 1 didn't move
    // But refs on lines that moved need their targets adjusted if targets are after insertion
    // Line 4 (was line 2): \1 stays \1 (target line 1 is before insertion)
    // Line 5 (was line 3): \2 becomes \4 (target line 2 moved to line 4)
    const newLine4 = await getLineText(page, 4);
    const newLine5 = await getLineText(page, 5);
    expect(newLine4).toContain('\\1'); // target line 1 didn't move
    expect(newLine5).toContain('\\4'); // target line 2 moved to line 4
  });

  test('should adjust references when inserting empty lines after all refs', async ({ page }) => {
    // Set up:
    // Line 1: 100 =
    // Line 2: \1 * 2 =
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line2 = await getLineText(page, 2);
    expect(line2).toContain('\\1');
    expect(line2).toContain('200');
    
    // Insert 2 empty lines after line 2 (after all refs)
    await goToLine(page, 2);
    await goToEndOfLine(page);
    await pressEnter(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Line 2 should still have \1 (refs after insertion point are not affected)
    const newLine2 = await getLineText(page, 2);
    expect(newLine2).toContain('\\1');
    expect(newLine2).toContain('200');
  });

  test('should adjust references when deleting multiple empty lines before refs', async ({ page }) => {
    // Set up:
    // Line 1: (empty)
    // Line 2: (empty)
    // Line 3: (empty)
    // Line 4: 100 =
    // Line 5: \4 * 2 =
    await pressEnter(page);
    await pressEnter(page);
    await pressEnter(page);
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\4 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line5 = await getLineText(page, 5);
    expect(line5).toContain('\\4');
    expect(line5).toContain('200');
    
    // Delete lines 1, 2, 3 (three empty lines)
    await goToLine(page, 1);
    await page.keyboard.press('Control+Shift+k'); // Delete line 1
    await page.keyboard.press('Control+Shift+k'); // Delete line 2
    await page.keyboard.press('Control+Shift+k'); // Delete line 3
    await waitForEvaluation(page);
    
    // Now line 2 should have \1 (was \4, shifted by -3)
    const newLine2 = await getLineText(page, 2);
    expect(newLine2).toContain('\\1');
  });

  test('should adjust references when deleting empty lines between refs', async ({ page }) => {
    // Set up:
    // Line 1: (empty)
    // Line 2: (empty)
    // Line 3: 100 =
    // Line 4: \3 * 2 =
    await pressEnter(page);
    await pressEnter(page);
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\3 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line4 = await getLineText(page, 4);
    expect(line4).toContain('\\3');
    expect(line4).toContain('200');
    
    // Delete lines 1 and 2 (empty lines before refs)
    await goToLine(page, 1);
    await page.keyboard.press('Control+Shift+k'); // Delete line 1
    await page.keyboard.press('Control+Shift+k'); // Delete line 2
    await waitForEvaluation(page);
    
    // Line 2 should have \1 (was \3, shifted by -2)
    const newLine2 = await getLineText(page, 2);
    expect(newLine2).toContain('\\1');
  });

  test('should handle complex scenario: multiple refs with inserts and deletes', async ({ page }) => {
    // Set up a complex chain:
    // Line 1: 10 =
    // Line 2: 20 =
    // Line 3: \1 + \2 =
    // Line 4: \3 * 2 =
    await typeInEditor(page, '10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '20 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 + \\2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\3 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line3 = await getLineText(page, 3);
    let line4 = await getLineText(page, 4);
    expect(line3).toContain('\\1');
    expect(line3).toContain('\\2');
    expect(line3).toContain('30'); // 10 + 20
    expect(line4).toContain('\\3');
    expect(line4).toContain('60'); // 30 * 2
    
    // Insert 2 empty lines at the beginning
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await pressEnter(page);
    await page.keyboard.press('ArrowUp');
    await waitForEvaluation(page);
    
    // Now:
    // Line 5 should have \3 + \4 (was \1 + \2)
    // Line 6 should have \5 (was \3)
    const newLine5 = await getLineText(page, 5);
    const newLine6 = await getLineText(page, 6);
    expect(newLine5).toContain('\\3');
    expect(newLine5).toContain('\\4');
    expect(newLine6).toContain('\\5');
  });
});
