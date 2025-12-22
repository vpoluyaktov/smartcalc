/**
 * E2E Test Helpers for SmartCalc
 * Provides utilities for interacting with the CodeMirror editor
 */

/**
 * Get the CodeMirror editor element
 * @param {import('@playwright/test').Page} page
 */
async function getEditor(page) {
  return page.locator('.cm-editor');
}

/**
 * Get the CodeMirror content element
 * @param {import('@playwright/test').Page} page
 */
async function getEditorContent(page) {
  return page.locator('.cm-content');
}

/**
 * Get the current text content of the editor
 * @param {import('@playwright/test').Page} page
 * @returns {Promise<string>}
 */
async function getEditorText(page) {
  // Get text from all cm-line elements
  const lines = await page.locator('.cm-line').allTextContents();
  return lines.join('\n');
}

/**
 * Set the editor content by selecting all and typing
 * @param {import('@playwright/test').Page} page
 * @param {string} text
 */
async function setEditorText(page, text) {
  const editor = await getEditorContent(page);
  await editor.click();
  // Select all and replace
  await page.keyboard.press('Control+a');
  await page.keyboard.type(text, { delay: 10 });
}

/**
 * Clear the editor content
 * @param {import('@playwright/test').Page} page
 */
async function clearEditor(page) {
  const editor = await getEditorContent(page);
  await editor.click();
  await page.keyboard.press('Control+a');
  await page.keyboard.press('Backspace');
}

/**
 * Type text at the current cursor position
 * @param {import('@playwright/test').Page} page
 * @param {string} text
 */
async function typeInEditor(page, text) {
  await page.keyboard.type(text, { delay: 5 });
}

/**
 * Press Enter key
 * @param {import('@playwright/test').Page} page
 */
async function pressEnter(page) {
  await page.keyboard.press('Enter');
}

/**
 * Press arrow keys to navigate
 * @param {import('@playwright/test').Page} page
 * @param {'Up' | 'Down' | 'Left' | 'Right'} direction
 * @param {number} times
 */
async function pressArrow(page, direction, times = 1) {
  for (let i = 0; i < times; i++) {
    await page.keyboard.press(`Arrow${direction}`);
  }
}

/**
 * Go to a specific line number
 * @param {import('@playwright/test').Page} page
 * @param {number} lineNumber - 1-based line number
 */
async function goToLine(page, lineNumber) {
  const editor = await getEditorContent(page);
  await editor.click();
  // Go to beginning of document
  await page.keyboard.press('Control+Home');
  // Move down to the target line
  for (let i = 1; i < lineNumber; i++) {
    await page.keyboard.press('ArrowDown');
  }
}

/**
 * Go to end of current line
 * @param {import('@playwright/test').Page} page
 */
async function goToEndOfLine(page) {
  await page.keyboard.press('End');
}

/**
 * Go to beginning of current line
 * @param {import('@playwright/test').Page} page
 */
async function goToBeginningOfLine(page) {
  await page.keyboard.press('Home');
}

/**
 * Wait for evaluation to complete (debounce + processing)
 * @param {import('@playwright/test').Page} page
 * @param {number} ms - milliseconds to wait (default 250ms for debounce + processing)
 */
async function waitForEvaluation(page, ms = 250) {
  await page.waitForTimeout(ms);
}

/**
 * Get text of a specific line
 * @param {import('@playwright/test').Page} page
 * @param {number} lineNumber - 1-based line number
 * @returns {Promise<string>}
 */
async function getLineText(page, lineNumber) {
  const lines = await page.locator('.cm-line').allTextContents();
  if (lineNumber < 1 || lineNumber > lines.length) {
    return '';
  }
  return lines[lineNumber - 1];
}

/**
 * Get the number of lines in the editor
 * @param {import('@playwright/test').Page} page
 * @returns {Promise<number>}
 */
async function getLineCount(page) {
  const lines = await page.locator('.cm-line').count();
  return lines;
}

/**
 * Check if a line contains a specific reference pattern
 * @param {import('@playwright/test').Page} page
 * @param {number} lineNumber
 * @param {string} refPattern - e.g., "\\6" or "\\7"
 * @returns {Promise<boolean>}
 */
async function lineContainsRef(page, lineNumber, refPattern) {
  const lineText = await getLineText(page, lineNumber);
  return lineText.includes(refPattern);
}

/**
 * Wait for the editor to be ready
 * @param {import('@playwright/test').Page} page
 */
async function waitForEditorReady(page) {
  await page.waitForSelector('.cm-editor', { state: 'visible' });
  await page.waitForSelector('.cm-content', { state: 'visible' });
  // Small delay to ensure editor is fully initialized
  await page.waitForTimeout(100);
}

module.exports = {
  getEditor,
  getEditorContent,
  getEditorText,
  setEditorText,
  clearEditor,
  typeInEditor,
  pressEnter,
  pressArrow,
  goToLine,
  goToEndOfLine,
  goToBeginningOfLine,
  waitForEvaluation,
  getLineText,
  getLineCount,
  lineContainsRef,
  waitForEditorReady,
};
