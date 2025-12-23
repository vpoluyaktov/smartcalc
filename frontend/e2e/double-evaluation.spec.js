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
} = require('./helpers');

test.describe('Double Evaluation Prevention', () => {
  // This test suite verifies that expressions are not evaluated twice.
  // The bug: some results contain '=' characters (e.g., base64 padding),
  // which could be mistakenly interpreted as the result delimiter,
  // causing the line to be re-evaluated and producing incorrect output.
  //
  // The fix requires that the result delimiter '=' must have a space
  // before it (e.g., "2 + 2 = 4"), while embedded '=' (like base64 padding)
  // does not have a space before it.

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should not double-evaluate base64 encode with single padding', async ({ page }) => {
    // Base64 encoding of "hello world" produces "aGVsbG8gd29ybGQ=" (single =)
    // The trailing = is padding, NOT a result delimiter.
    await typeInEditor(page, 'base64 encode hello world =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('base64 encode hello world = aGVsbG8gd29ybGQ=');
    // Should NOT contain double-encoded result
    expect(line1).not.toContain('aGVsbG8gd29ybGQgPSBhR1ZzYkc4Z2QyOXliR1E');
  });

  test('should not double-evaluate base64 encode with double padding', async ({ page }) => {
    // Base64 encoding of "test" produces "dGVzdA==" (double ==)
    await typeInEditor(page, 'base64 encode test =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('base64 encode test = dGVzdA==');
  });

  test('should not double-evaluate base64 encode without padding', async ({ page }) => {
    // Base64 encoding of "abc" produces "YWJj" (no padding)
    await typeInEditor(page, 'base64 encode abc =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('base64 encode abc = YWJj');
  });

  test('should handle base64 decode with padding in input', async ({ page }) => {
    // Decoding "SGVsbG8gd29ybGQ=" (which contains =) should work correctly
    await typeInEditor(page, 'base64 decode SGVsbG8gd29ybGQ= =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('Hello world');
  });

  test('should preserve base64 result when editing and re-evaluating', async ({ page }) => {
    // Type base64 encode, get result, then go back and trigger re-evaluation
    // The result should remain stable (not double-encoded)
    await typeInEditor(page, 'base64 encode hello =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    let line1 = await getLineText(page, 1);
    expect(line1).toBe('base64 encode hello = aGVsbG8=');
    
    // Go back to line 1 and move away to trigger potential re-evaluation
    await goToLine(page, 1);
    await page.keyboard.press('End');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    // Line 1 should still have the same result (not double-encoded)
    line1 = await getLineText(page, 1);
    expect(line1).toBe('base64 encode hello = aGVsbG8=');
  });

  test('should not confuse comparison operators with result delimiter', async ({ page }) => {
    // Comparison expressions use ==, >=, <=, != which contain =
    // These should not be confused with the result delimiter
    await typeInEditor(page, '5 == 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('true');
    // Should have exactly one result, not re-evaluated
    expect(line1).toBe('5 == 5 = true');
  });

  test('should handle >= comparison correctly', async ({ page }) => {
    await typeInEditor(page, '10 >= 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('10 >= 5 = true');
  });

  test('should handle <= comparison correctly', async ({ page }) => {
    await typeInEditor(page, '3 <= 7 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('3 <= 7 = true');
  });

  test('should handle != comparison correctly', async ({ page }) => {
    await typeInEditor(page, '4 != 5 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('4 != 5 = true');
  });

  test('should handle MD5 hash result correctly', async ({ page }) => {
    // MD5 produces hex output without = but tests the general pattern
    await typeInEditor(page, 'md5 hello =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('md5 hello = ');
    // MD5 of "hello" is 5d41402abc4b2a76b9719d911017c592
    expect(line1).toContain('5d41402abc4b2a76b9719d911017c592');
  });

  test('should handle SHA1 hash result correctly', async ({ page }) => {
    await typeInEditor(page, 'sha1 hello =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('sha1 hello = ');
    // SHA1 of "hello" is aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d
    expect(line1).toContain('aaf4c61ddcc5e8a2dabede0f3b482cd9aea9434d');
  });

  test('should handle SHA256 hash result correctly', async ({ page }) => {
    await typeInEditor(page, 'sha256 hello =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('sha256 hello = ');
    // SHA256 of "hello" is 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
    expect(line1).toContain('2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824');
  });

  test('should handle UUID generation correctly', async ({ page }) => {
    // UUID contains hyphens but no = characters
    await typeInEditor(page, 'uuid =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('uuid = ');
    // UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
    expect(line1).toMatch(/uuid = [0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/);
  });

  test('should handle multiple base64 lines without cross-contamination', async ({ page }) => {
    // Multiple base64 operations should each produce correct results
    await typeInEditor(page, 'base64 encode foo =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, 'base64 encode bar =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    await typeInEditor(page, 'base64 encode baz =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    const line2 = await getLineText(page, 2);
    const line3 = await getLineText(page, 3);
    
    expect(line1).toBe('base64 encode foo = Zm9v');
    expect(line2).toBe('base64 encode bar = YmFy');
    expect(line3).toBe('base64 encode baz = YmF6');
  });

  test('should handle expression with = in string context', async ({ page }) => {
    // Test that expressions like "2+2=" don't get confused
    // when there's no space before =
    await typeInEditor(page, '2 + 2 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toBe('2 + 2 = 4');
    
    // Edit and re-evaluate - should remain stable
    await goToLine(page, 1);
    await page.keyboard.press('ArrowDown');
    await waitForEvaluation(page);
    
    const line1After = await getLineText(page, 1);
    expect(line1After).toBe('2 + 2 = 4');
  });

  test('should handle bitwise operations with hex output', async ({ page }) => {
    // Bitwise operations produce results like "255 (0xFF)"
    await typeInEditor(page, '0xFF and 0x0F =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('0xFF and 0x0F = ');
    expect(line1).toContain('15');
    expect(line1).toContain('0xF');
  });

  test('should handle left shift operation', async ({ page }) => {
    await typeInEditor(page, '1 << 8 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('1 << 8 = ');
    expect(line1).toContain('256');
  });

  test('should handle right shift operation', async ({ page }) => {
    await typeInEditor(page, '256 >> 4 =');
    await pressEnter(page);
    await waitForEvaluation(page);
    
    const line1 = await getLineText(page, 1);
    expect(line1).toContain('256 >> 4 = ');
    expect(line1).toContain('16');
  });
});
