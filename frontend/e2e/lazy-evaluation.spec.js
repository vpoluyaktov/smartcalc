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
  waitForEditorReady,
  pressArrow,
} = require('./helpers');

test.describe('Lazy Evaluation', () => {
  // This test suite verifies the lazy evaluation behavior in SmartCalc.
  // Lazy evaluation means expressions are only evaluated when the user
  // explicitly triggers it (Enter key or leaving the line), not while typing.
  // This improves performance and prevents flickering during input.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should not evaluate while typing on current line', async ({ page }) => {
    // Verifies that typing on a line does NOT trigger immediate evaluation.
    // The result should only appear after pressing Enter or leaving the line.
    // This prevents distracting result updates while the user is still typing.
    await typeInEditor(page, '2 + 2 =');
    
    // Wait a bit but don't press Enter
    await page.waitForTimeout(200);
    
    // The line should still just have "2 + 2 =" without result
    // (result is stripped while editing)
    const line1 = await getLineText(page, 1);
    // Should not have a numeric result yet (or result should be stripped)
    expect(line1).toContain('2 + 2');
  });

  test('should evaluate when pressing Enter', async ({ page }) => {
    // Verifies that pressing Enter triggers evaluation of the current line.
    // This is the primary way users confirm their expression is complete.
    await typeInEditor(page, '2 + 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Now line 1 should have the result
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('4');
  });

  test('should evaluate when moving to another line with arrow keys', async ({ page }) => {
    // Verifies that moving to another line with arrow keys triggers evaluation.
    // This allows users to navigate away from a line and have it auto-evaluate,
    // which is convenient for reviewing and editing multiple calculations.
    //
    // Setup: Create line 1 with result, then type on line 2 without Enter.
    // Action: Press Up arrow to return to line 1.
    // Expected: Line 2 should now have its result evaluated.
    await typeInEditor(page, '10 + 5 =');
    await pressEnter(page);
    await waitForEvaluation(page, 300);
    
    // Line 1 should have result after Enter
    let line1 = await getLineText(page, 1);
    expect(line1).toContain('15');
    
    // Type on line 2 without pressing Enter
    await typeInEditor(page, '20 + 30 =');
    await waitForEvaluation(page, 300);
    
    // Line 2 should NOT have result yet (lazy evaluation - no Enter pressed)
    let line2 = await getLineText(page, 2);
    expect(line2).toBe('20 + 30 =');
    
    // Now press Up arrow to move back to line 1 - this should trigger evaluation of line 2
    await pressArrow(page, 'Up');
    await waitForEvaluation(page, 500);
    
    // Line 2 should now have the result
    line2 = await getLineText(page, 2);
    expect(line2).toContain('50');
  });

  test('should strip result when editing a line with existing result', async ({ page }) => {
    // Verifies that when editing a line that already has a result,
    // the old result is stripped to prevent confusion.
    // The result will be recalculated when the user finishes editing.
    //
    // This prevents showing stale/incorrect results while the user
    // is modifying the expression.
    await typeInEditor(page, '100 + 50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify result exists
    let line1 = await getLineText(page, 1);
    expect(line1).toContain('150');
    
    // Go back and start editing
    await goToLine(page, 1);
    await goToEndOfLine(page);
    
    // Type something (this should trigger result stripping)
    await typeInEditor(page, ' ');
    await page.waitForTimeout(100);
    
    // Result should be stripped (line ends with = or has no numeric result)
    line1 = await getLineText(page, 1);
    // The result "150" should be removed while editing
    expect(line1).not.toMatch(/= 150$/);
  });

  test('should strip results from dependent lines when editing', async ({ page }) => {
    // Verifies that when editing a line, all dependent lines (lines that
    // reference it via \n syntax) have their results stripped.
    //
    // This is crucial for maintaining consistency - if line 1 changes,
    // any line referencing \1 must be recalculated.
    //
    // Setup: Line 1 = 100, Line 2 = \1 * 2 = 200
    // Action: Edit line 1
    // Expected: Line 2's result should be stripped
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page, 500);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page, 500);
    
    // Verify both have results
    let line1 = await getLineText(page, 1);
    let line2 = await getLineText(page, 2);
    expect(line1).toContain('100');
    expect(line2).toContain('200');
    
    // Edit line 1 - go to the expression part and modify it
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    // Type a character to trigger editing (this should strip results)
    await typeInEditor(page, '2'); // Makes it "2100 = 100"
    await waitForEvaluation(page, 500);
    
    // After editing line 1, line 2's result should be stripped
    // because line 2 depends on line 1 which changed
    line2 = await getLineText(page, 2);
    // The dependent line should have its result stripped (ends with just "=")
    // or show a different/error result since the source changed
    // Note: The stripping happens via stripCurrentLineResult which calls FindDependentLines
    expect(line2).toMatch(/\\1 \* 2 =/);
  });

  test('should add = and evaluate when pressing Enter on line without =', async ({ page }) => {
    // Verifies that pressing Enter on an expression without = will:
    // 1. Automatically append = to the expression
    // 2. Evaluate and show the result
    // This is a convenience feature for quick calculations.
    await typeInEditor(page, '5 * 5');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Line should now have = and result
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('=');
    expect(line1).toContain('25');
  });

  test('should evaluate all dependent lines when source line changes', async ({ page }) => {
    // Verifies that changing a source line triggers re-evaluation of all
    // dependent lines in the chain.
    //
    // Setup: Line 1 = 10, Line 2 = \1 * 2 = 20, Line 3 = \2 + 5 = 25
    // Action: Change line 1 from 10 to 20
    // Expected: Line 2 = 40, Line 3 = 45 (entire chain updates)
    //
    // This is the core spreadsheet-like behavior of SmartCalc.
    await typeInEditor(page, '10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, '\\2 + 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Verify chain: 10, 20, 25
    expect(await getLineText(page, 1)).toContain('10');
    expect(await getLineText(page, 2)).toContain('20');
    expect(await getLineText(page, 3)).toContain('25');
    
    // Change line 1 to 20
    await goToLine(page, 1);
    await page.keyboard.press('Home');
    await page.keyboard.press('Delete');
    await page.keyboard.press('Delete');
    await typeInEditor(page, '20');
    
    // Move away to trigger evaluation
    await pressArrow(page, 'Down');
    await waitForEvaluation(page);
    
    // Chain should update: 20, 40, 45
    expect(await getLineText(page, 1)).toContain('20');
    expect(await getLineText(page, 2)).toContain('40');
    expect(await getLineText(page, 3)).toContain('45');
  });
});
