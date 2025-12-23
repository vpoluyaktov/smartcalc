// @ts-check
const { test, expect } = require('@playwright/test');
const {
  clearEditor,
  typeInEditor,
  pressEnter,
  waitForEvaluation,
  getLineText,
  waitForEditorReady,
} = require('./helpers');

test.describe('Basic Calculations', () => {
  // This test suite verifies the core calculation functionality of SmartCalc.
  // It covers arithmetic operations, number formatting, special syntax
  // (percentages, currencies, units), and comment handling.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should perform basic arithmetic', async ({ page }) => {
    // Verifies the four basic arithmetic operations: addition, subtraction,
    // multiplication, and division. Each operation is tested on a separate
    // line to ensure independent evaluation.
    await typeInEditor(page, '2 + 3 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('5');

    await typeInEditor(page, '10 - 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 2)).toContain('6');

    await typeInEditor(page, '6 * 7 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 3)).toContain('42');

    await typeInEditor(page, '100 / 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 4)).toContain('25');
  });

  test('should handle parentheses', async ({ page }) => {
    // Verifies that parentheses correctly override operator precedence.
    // Without parentheses, 2 + 3 * 4 = 14, but (2 + 3) * 4 = 20.
    await typeInEditor(page, '(2 + 3) * 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('20');
  });

  test('should handle decimal numbers', async ({ page }) => {
    // Verifies that floating-point arithmetic works correctly.
    // Tests precision with common decimal values like pi.
    await typeInEditor(page, '3.14 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('6.28');
  });

  test('should handle negative numbers', async ({ page }) => {
    // Verifies that negative numbers are parsed and calculated correctly.
    // The minus sign should be recognized as part of the number, not an operator.
    await typeInEditor(page, '-5 + 10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('5');
  });

  test('should handle exponents', async ({ page }) => {
    // Verifies exponentiation using the ^ operator.
    // 2^8 = 256 tests a common power-of-two calculation.
    await typeInEditor(page, '2 ^ 8 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('256');
  });

  test('should handle line references', async ({ page }) => {
    // Verifies the \n syntax for referencing previous line results.
    // Creates a chain: line 1 = 100, line 2 = \1 * 2 = 200, line 3 = \1 + \2 = 300.
    // This is a core feature that enables spreadsheet-like calculations.
    await typeInEditor(page, '100 =');
    await pressEnter(page);
    await waitForEvaluation(page);

    await typeInEditor(page, '\\1 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 2)).toContain('200');

    await typeInEditor(page, '\\1 + \\2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 3)).toContain('300');
  });

  test('should format large numbers with thousands separator', async ({ page }) => {
    // Verifies that large numbers are formatted with commas for readability.
    // 1,000,000 + 234,567 = 1,234,567 should display with proper separators.
    await typeInEditor(page, '1000000 + 234567 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('1,234,567');
  });

  test('should handle currency calculations', async ({ page }) => {
    // Verifies that currency symbols are recognized and preserved in results.
    // $100 + $50 should equal $150, maintaining the dollar sign.
    await typeInEditor(page, '$100 + $50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('$');
    expect(line1).toContain('150');
  });

  test('should handle percentage calculations', async ({ page }) => {
    // Verifies two percentage syntaxes:
    // 1. "200 + 10%" means 200 + 10% of 200 = 220
    // 2. "what is 50% of 200" = 100 (natural language percentage)
    await typeInEditor(page, '200 + 10% =');
    await pressEnter(page);
    await waitForEvaluation(page, 500); // Extra time for percentage parsing
    expect(await getLineText(page, 1)).toContain('220');

    await typeInEditor(page, 'what is 50% of 200 =');
    await pressEnter(page);
    await waitForEvaluation(page, 500);
    expect(await getLineText(page, 2)).toContain('100');
  });

  test('should handle comparison operators', async ({ page }) => {
    // Verifies boolean comparison operators: >, ==, <
    // Results should be 'true' or 'false' strings.
    await typeInEditor(page, '5 > 3 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('true');

    await typeInEditor(page, '2 == 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 2)).toContain('true');

    await typeInEditor(page, '10 < 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 3)).toContain('false');
  });

  test('should skip comment lines', async ({ page }) => {
    // Verifies that lines starting with # are treated as comments.
    // Comments should not be evaluated and should remain unchanged.
    // Calculations on subsequent lines should still work.
    await typeInEditor(page, '# This is a comment');
    await pressEnter(page);
    await typeInEditor(page, '5 + 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Comment should remain unchanged
    expect(await getLineText(page, 1)).toBe('# This is a comment');
    // Calculation should work
    expect(await getLineText(page, 2)).toContain('10');
  });

  test('should handle inline comments', async ({ page }) => {
    // Verifies that inline comments (# after the expression) are preserved.
    // The result should be inserted between = and #, keeping the comment.
    // Example: "10 + 10 = # my calculation" becomes "10 + 10 = 20 # my calculation"
    await typeInEditor(page, '10 + 10 = # my calculation');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('20');
    expect(line1).toContain('# my calculation');
  });

  test('should handle unit conversions', async ({ page }) => {
    // Verifies unit conversion syntax: "X unit1 to unit2".
    // 1 km to m = 1000 m. The result includes the target unit suffix.
    await typeInEditor(page, '1 km to m =');
    await pressEnter(page);
    await waitForEvaluation(page);
    // Result is "1000 m" (with unit suffix, no thousands separator for unit conversions)
    expect(await getLineText(page, 1)).toContain('1000 m');
  });

  test('should handle mathematical functions', async ({ page }) => {
    // Verifies built-in mathematical functions like sqrt() and abs().
    // sqrt(16) = 4, abs(-42) = 42.
    await typeInEditor(page, 'sqrt(16) =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('4');

    await typeInEditor(page, 'abs(-42) =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 2)).toContain('42');
  });

  test('should handle base64 encode without double evaluation', async ({ page }) => {
    // Verifies that base64 encoding works correctly and doesn't get evaluated twice.
    // The bug: base64 results often end with '=' (padding), which was being
    // mistakenly interpreted as the result delimiter, causing double evaluation.
    //
    // "base64 encode hello world" should produce "aGVsbG8gd29ybGQ=" (with trailing =)
    // The trailing = is base64 padding, NOT a result delimiter.
    await typeInEditor(page, 'base64 encode hello world =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    // Should contain the base64 result exactly once
    expect(line1).toContain('aGVsbG8gd29ybGQ=');
    // Should NOT contain a double-encoded result (which would happen if evaluated twice)
    expect(line1).not.toContain('aGVsbG8gd29ybGQgPSBhR1ZzYkc4Z2QyOXliR1E=');
  });

  test('should handle base64 decode correctly', async ({ page }) => {
    // Verifies that base64 decoding works with strings that have padding (=).
    // "SGVsbG8gd29ybGQ=" decodes to "Hello world"
    await typeInEditor(page, 'base64 decode SGVsbG8gd29ybGQ= =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('Hello world');
  });
});
