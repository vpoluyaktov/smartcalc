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
  // This test suite verifies that line references (\n syntax) are automatically
  // adjusted when lines are inserted or deleted. This is a core feature that
  // maintains referential integrity in spreadsheet-like calculations.
  //
  // When a line is inserted, all references pointing to lines at or after the
  // insertion point must be incremented. When a line is deleted, references
  // must be decremented accordingly.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should adjust references when inserting a line before them', async ({ page }) => {
    // Verifies that inserting a line before a reference causes the reference
    // to be incremented. If \1 points to line 1, and we insert a line before
    // line 1, then \1 should become \2.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2
    // Action: Insert empty line before line 1
    // Expected: Line 3 now has \2 (was \1)
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
    // Verifies that inserting a line in the middle of the document adjusts
    // only the references that point to lines at or after the insertion point.
    //
    // Setup: Line 1 = comment, Line 2 = 2+2=4, Line 3 = \2 * 4
    // Action: Insert empty line after line 1
    // Expected: Line 4 now has \3 (was \2, because line 2 moved to line 3)
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
    // Verifies that a chain of references is correctly adjusted when a line
    // is inserted. All references in the chain must be updated.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2 = 200, Line 3 = \2 + 50 = 250
    // Action: Insert line at beginning
    // Expected: Line 3 = \2, Line 4 = \3 (both shifted by 1)
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
    // Verifies that deleting a line before a reference causes the reference
    // to be decremented. If \2 points to line 2, and we delete line 1,
    // then \2 should become \1.
    //
    // Setup: Line 1 = empty, Line 2 = 100, Line 3 = \2 * 2
    // Action: Delete line 1
    // Expected: Line 2 now has \1 (was \2)
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
    // Reproduces a real user scenario with comments, empty lines, and chained
    // references. Verifies that inserting an empty line correctly shifts all
    // subsequent references.
    //
    // Setup:
    // - Lines 1-3: comments
    // - Lines 4-5: empty
    // - Line 6: 2 + 2 = 4
    // - Line 7: empty
    // - Line 8: \6 x 4 = 16
    // - Line 9: empty
    // - Line 10: \8 / 2 = 8
    //
    // Action: Insert line at line 5
    // Expected: \6 becomes \7, \8 becomes \9
    
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
    // Verifies that references pointing to lines BEFORE the insertion point
    // are NOT adjusted. Only references to lines at or after the insertion
    // point should be incremented.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2, Line 3 = 50
    // Action: Insert line after line 2
    // Expected: Line 2 still has \1 (unchanged, points before insertion)
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
    // Verifies that inserting MULTIPLE lines at once correctly shifts all
    // references by the number of lines inserted.
    //
    // Setup with empty lines between values:
    // - Line 1: 100 =
    // - Line 2: (empty)
    // - Line 3: 50 =
    // - Line 4: (empty)
    // - Line 5: \1 + \3 = 150
    //
    // Action: Insert 3 empty lines before line 1
    // Expected: \1 becomes \4, \3 becomes \6 (both shifted by 3)
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await pressEnter(page); // Empty line 2
    
    await typeInEditor(page, '50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await pressEnter(page); // Empty line 4
    
    await typeInEditor(page, '\\1 + \\3 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify initial state
    let line5 = await getLineText(page, 5);
    expect(line5).toContain('\\1');
    expect(line5).toContain('\\3');
    expect(line5).toContain('150');
    
    // Insert 3 empty lines before line 1 using multiple Enter presses
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await pressEnter(page);
    await pressEnter(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Now refs should be shifted by 3:
    // Line 8 should have \4 + \6 (was \1 + \3)
    const line8 = await getLineText(page, 8);
    expect(line8).toContain('\\4');
    expect(line8).toContain('\\6');
  });

  test('should adjust references when inserting empty lines between refs', async ({ page }) => {
    // Verifies reference adjustment when inserting lines between existing
    // references. References pointing to lines before the insertion stay
    // unchanged, while references pointing to lines after are incremented.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2, Line 3 = \2 + 10
    // Action: Insert 2 empty lines after line 1
    // Expected: Line 4 keeps \1, Line 5 has \4 (was \2)
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
    // Verifies that inserting lines AFTER all references does not affect
    // any existing references. This is an edge case where no adjustment
    // should occur.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2 = 200
    // Action: Insert 2 empty lines after line 2
    // Expected: Line 2 still has \1 and result 200 (unchanged)
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
    // Verifies that deleting MULTIPLE lines correctly decrements references
    // by the number of lines deleted.
    //
    // Setup:
    // - Lines 1-3: empty
    // - Line 4: 100 =
    // - Line 5: \4 * 2 = 200
    //
    // Action: Delete lines 1, 2, 3
    // Expected: Line 2 now has \1 (was \4, shifted by -3)
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
    // Verifies reference adjustment when deleting lines between references.
    // References pointing to lines after the deleted section are decremented.
    //
    // Setup:
    // - Lines 1-2: empty
    // - Line 3: 100 =
    // - Line 4: \3 * 2 = 200
    //
    // Action: Delete lines 1 and 2
    // Expected: Line 2 now has \1 (was \3, shifted by -2)
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
    // Verifies a complex scenario with multiple references and chained
    // dependencies. Tests that all references in a chain are correctly
    // adjusted when lines are inserted.
    //
    // Setup:
    // - Line 1: 10 =
    // - Line 2: 20 =
    // - Line 3: \1 + \2 = 30
    // - Line 4: \3 * 2 = 60
    //
    // Action: Insert 2 empty lines at the beginning
    // Expected: Line 5 = \3 + \4, Line 6 = \5 (all shifted by 2)
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

  test('should immediately show correct results after inserting lines before refs (no ERR)', async ({ page }) => {
    // This test catches a bug where references would show ERR temporarily
    // after inserting lines, because the old results remained but the
    // reference numbers were adjusted to point to non-existent lines.
    //
    // Setup matching the user's scenario:
    // Line 1-3: comments
    // Line 4: empty
    // Line 5: empty (cursor here)
    // Line 6: 2 + 2 = 4
    // Line 7: empty
    // Line 8: \6 x 4 = 16
    // Line 9: empty
    // Line 10: \8 / 2 = 8
    await typeInEditor(page, '# Comment 1');
    await pressEnter(page);
    await typeInEditor(page, '# Comment 2');
    await pressEnter(page);
    await typeInEditor(page, '# Comment 3');
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
    
    // Verify initial state - all results should be correct
    let line6 = await getLineText(page, 6);
    let line8 = await getLineText(page, 8);
    let line10 = await getLineText(page, 10);
    expect(line6).toContain('2 + 2 = 4');
    expect(line8).toContain('\\6');
    expect(line8).toContain('16');
    expect(line10).toContain('\\8');
    expect(line10).toContain('8');
    
    // Go to line 5 (empty line) and press Enter twice
    await goToLine(page, 5);
    await pressEnter(page);
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // After inserting 2 lines at line 5:
    // - Line 8 should now have 2 + 2 = 4
    // - Line 10 should have \8 x 4 = 16 (NOT ERR!)
    // - Line 12 should have \10 / 2 = 8 (NOT ERR!)
    const newLine8 = await getLineText(page, 8);
    const newLine10 = await getLineText(page, 10);
    const newLine12 = await getLineText(page, 12);
    
    expect(newLine8).toContain('2 + 2 = 4');
    expect(newLine10).toContain('\\8');
    expect(newLine10).toContain('16');
    expect(newLine10).not.toContain('ERR');
    expect(newLine12).toContain('\\10');
    expect(newLine12).toContain('8');
    expect(newLine12).not.toContain('ERR');
  });
});
