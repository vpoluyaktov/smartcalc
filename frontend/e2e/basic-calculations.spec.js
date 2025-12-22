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
  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should perform basic arithmetic', async ({ page }) => {
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
    await typeInEditor(page, '(2 + 3) * 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('20');
  });

  test('should handle decimal numbers', async ({ page }) => {
    await typeInEditor(page, '3.14 * 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('6.28');
  });

  test('should handle negative numbers', async ({ page }) => {
    await typeInEditor(page, '-5 + 10 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('5');
  });

  test('should handle exponents', async ({ page }) => {
    await typeInEditor(page, '2 ^ 8 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('256');
  });

  test('should handle line references', async ({ page }) => {
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
    await typeInEditor(page, '1000000 + 234567 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('1,234,567');
  });

  test('should handle currency calculations', async ({ page }) => {
    await typeInEditor(page, '$100 + $50 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('$');
    expect(line1).toContain('150');
  });

  test('should handle percentage calculations', async ({ page }) => {
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
    await typeInEditor(page, '10 + 10 = # my calculation');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('20');
    expect(line1).toContain('# my calculation');
  });

  test('should handle unit conversions', async ({ page }) => {
    await typeInEditor(page, '1 km to m =');
    await pressEnter(page);
    await waitForEvaluation(page);
    // Result is "1000 m" (with unit suffix, no thousands separator for unit conversions)
    expect(await getLineText(page, 1)).toContain('1000 m');
  });

  test('should handle mathematical functions', async ({ page }) => {
    await typeInEditor(page, 'sqrt(16) =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 1)).toContain('4');

    await typeInEditor(page, 'abs(-42) =');
    await pressEnter(page);
    await waitForEvaluation(page);
    expect(await getLineText(page, 2)).toContain('42');
  });
});
