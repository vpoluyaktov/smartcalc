// @ts-check
const { test, expect } = require('@playwright/test');
const {
  clearEditor,
  typeInEditor,
  pressEnter,
  waitForEvaluation,
  getLineText,
  getLineCount,
  waitForEditorReady,
  getEditorContent,
  goToLine,
} = require('./helpers');

test.describe('JWT Decoder', () => {
  // This test suite verifies the JWT decoder functionality.
  // JWT tokens contain base64url-encoded data which includes characters like '-' and '_'
  // that could be mistakenly interpreted as operators.

  // A valid JWT token for testing (header.payload.signature)
  // Header: {"alg":"HS256","typ":"JWT"}
  // Payload: {"sub":"1234567890","name":"John Doe","email":"john@example.com","role":"admin","iat":1516239022,"exp":1893456000}
  const TEST_JWT_TOKEN = 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huQGV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE4OTM0NTYwMDB9.tjYf8RqHCqXhGt-FY1RqCVLgZLgpG1xYn8OU_U0RQWM';

  test.beforeEach(async ({ page }) => {
    await page.goto('/');
    await waitForEditorReady(page);
    await clearEditor(page);
  });

  test('should decode JWT token and show header/payload', async ({ page }) => {
    // Verifies that JWT tokens are decoded correctly and display
    // the header, payload, and signature information.
    await typeInEditor(page, `jwt decode ${TEST_JWT_TOKEN} =`);
    await pressEnter(page);
    await waitForEvaluation(page, 500);

    // Check that the output contains expected JWT parts
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThan(1); // Multi-line output

    // Get all lines and check for JWT output markers
    let hasHeader = false;
    let hasPayload = false;
    let hasSignature = false;
    let hasJohnDoe = false;

    for (let i = 1; i <= lineCount; i++) {
      const line = await getLineText(page, i);
      if (line.includes('Header:')) hasHeader = true;
      if (line.includes('Payload:')) hasPayload = true;
      if (line.includes('Signature:')) hasSignature = true;
      if (line.includes('John Doe')) hasJohnDoe = true;
    }

    expect(hasHeader).toBe(true);
    expect(hasPayload).toBe(true);
    expect(hasSignature).toBe(true);
    expect(hasJohnDoe).toBe(true);
  });

  test('should preserve JWT output when clicking elsewhere in editor', async ({ page }) => {
    // This test verifies the fix for the bug where JWT output would disappear
    // and show "= ERR" when clicking elsewhere in the editor.
    //
    // Root cause: The formatExpression function was corrupting JWT tokens
    // by adding spaces around '-' characters (valid in base64url encoding).
    // When re-evaluating after clicking elsewhere, the corrupted token
    // would fail to decode.

    // Step 1: Type and evaluate the JWT expression
    await typeInEditor(page, `jwt decode ${TEST_JWT_TOKEN} =`);
    await pressEnter(page);
    await waitForEvaluation(page, 500);

    // Verify initial evaluation succeeded
    const initialLineCount = await getLineCount(page);
    expect(initialLineCount).toBeGreaterThan(5); // JWT output has many lines

    // Check that we have valid output (not ERR)
    const line1 = await getLineText(page, 1);
    expect(line1).not.toContain('ERR');
    expect(line1).toContain('jwt decode');

    // Step 2: Click on a different line (simulating user clicking elsewhere)
    // This triggers re-evaluation which was causing the bug
    await goToLine(page, 3); // Go to an output line
    await waitForEvaluation(page, 500);

    // Step 3: Verify the JWT output is still valid after clicking elsewhere
    const afterClickLineCount = await getLineCount(page);
    expect(afterClickLineCount).toBeGreaterThan(5); // Should still have multi-line output

    // The first line should still contain the JWT expression without ERR
    const line1AfterClick = await getLineText(page, 1);
    expect(line1AfterClick).not.toContain('ERR');
    expect(line1AfterClick).toContain('jwt decode');

    // Verify the output still contains expected content
    let hasHeaderAfterClick = false;
    let hasPayloadAfterClick = false;

    for (let i = 1; i <= afterClickLineCount; i++) {
      const line = await getLineText(page, i);
      if (line.includes('Header:')) hasHeaderAfterClick = true;
      if (line.includes('Payload:')) hasPayloadAfterClick = true;
    }

    expect(hasHeaderAfterClick).toBe(true);
    expect(hasPayloadAfterClick).toBe(true);
  });

  test('should handle JWT token with special base64url characters', async ({ page }) => {
    // JWT tokens use base64url encoding which includes '-' and '_' characters.
    // These characters should NOT be interpreted as operators.
    // This test uses a token that contains many '-' characters in the signature.

    // Token with signature containing multiple '-' characters
    // Header: {"alg":"RS256","typ":"JWT","kid":"test-key-id"}
    // Payload: {"sub":"user-123","aud":["api-v1","api-v2"],"iss":"https://auth.example.com/"}
    const TOKEN_WITH_DASHES = 'eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6InRlc3Qta2V5LWlkIn0.eyJzdWIiOiJ1c2VyLTEyMyIsImF1ZCI6WyJhcGktdjEiLCJhcGktdjIiXSwiaXNzIjoiaHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tLyJ9.test-signature-with-dashes';

    await typeInEditor(page, `jwt decode ${TOKEN_WITH_DASHES} =`);
    await pressEnter(page);
    await waitForEvaluation(page, 500);

    // Should decode successfully, not show ERR
    const line1 = await getLineText(page, 1);
    expect(line1).not.toContain('ERR');

    // Should show the decoded content
    let hasRS256 = false;
    let hasTestKeyId = false;
    const lineCount = await getLineCount(page);

    for (let i = 1; i <= lineCount; i++) {
      const line = await getLineText(page, i);
      if (line.includes('RS256')) hasRS256 = true;
      if (line.includes('test-key-id')) hasTestKeyId = true;
    }

    expect(hasRS256).toBe(true);
    expect(hasTestKeyId).toBe(true);
  });

  test('should handle raw JWT token without prefix', async ({ page }) => {
    // JWT tokens can be decoded without the "jwt decode" prefix
    await typeInEditor(page, `${TEST_JWT_TOKEN} =`);
    await pressEnter(page);
    await waitForEvaluation(page, 500);

    // Should decode successfully
    const lineCount = await getLineCount(page);
    expect(lineCount).toBeGreaterThan(1);

    let hasHeader = false;
    for (let i = 1; i <= lineCount; i++) {
      const line = await getLineText(page, i);
      if (line.includes('Header:')) hasHeader = true;
    }

    expect(hasHeader).toBe(true);
  });
});
