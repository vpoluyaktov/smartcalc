import './style.css';
import { EditorView, basicSetup } from 'codemirror';
import { EditorState, RangeSetBuilder } from '@codemirror/state';
import { keymap, Decoration, ViewPlugin } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view';
import { Evaluate, GetVersion, OpenFileDialog, SaveFileDialog, ReadFile, WriteFile, AddRecentFile, GetLastFile, AutoSave, AdjustReferences, CopyWithResolvedRefs } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

let editor;
let currentFile = '';
let debounceTimer = null;
let autosaveTimer = null;
let previousText = '';
let previousLineCount = 0;
let isUpdatingEditor = false; // Flag to prevent re-entry during programmatic updates
const AUTOSAVE_DELAY = 2000; // 2 seconds after last change

// Dark theme (Tokyo Night inspired)
const darkTheme = EditorView.theme({
    '&': {
        backgroundColor: '#1a1b26',
        color: '#c0caf5',
    },
    '.cm-content': {
        caretColor: '#7aa2f7',
    },
    '.cm-cursor': {
        borderLeftColor: '#7aa2f7',
    },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
        backgroundColor: 'rgba(122, 162, 247, 0.3)',
    },
    '.cm-gutters': {
        backgroundColor: '#16161e',
        color: '#565f89',
        borderRight: '1px solid #3b4261',
    },
    '.cm-activeLineGutter': {
        backgroundColor: '#292e42',
        color: '#7aa2f7',
    },
    '.cm-activeLine': {
        backgroundColor: 'rgba(41, 46, 66, 0.5)',
    },
}, { dark: true });

// Light theme
const lightTheme = EditorView.theme({
    '&': {
        backgroundColor: '#f8f9fa',
        color: '#343a40',
    },
    '.cm-content': {
        caretColor: '#4285f4',
    },
    '.cm-cursor': {
        borderLeftColor: '#4285f4',
    },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
        backgroundColor: 'rgba(66, 133, 244, 0.25)',
    },
    '.cm-gutters': {
        backgroundColor: '#f1f3f4',
        color: '#868e96',
        borderRight: '1px solid #dee2e6',
    },
    '.cm-activeLineGutter': {
        backgroundColor: '#e9ecef',
        color: '#495057',
    },
    '.cm-activeLine': {
        backgroundColor: 'rgba(0, 0, 0, 0.04)',
    },
}, { dark: false });

// Detect system theme preference
function getSystemTheme() {
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

// Get the appropriate theme based on system preference
function getCurrentTheme() {
    return getSystemTheme() === 'dark' ? darkTheme : lightTheme;
}

// Syntax highlighting decorations
const resultMark = Decoration.mark({ class: 'cm-result' });
const errorMark = Decoration.mark({ class: 'cm-error' });
const referenceMark = Decoration.mark({ class: 'cm-reference' });
const numberMark = Decoration.mark({ class: 'cm-number' });
const operatorMark = Decoration.mark({ class: 'cm-operator' });
const currencyMark = Decoration.mark({ class: 'cm-currency' });
const percentMark = Decoration.mark({ class: 'cm-percent' });
const functionMark = Decoration.mark({ class: 'cm-function' });
const keywordMark = Decoration.mark({ class: 'cm-keyword' });
const datetimeMark = Decoration.mark({ class: 'cm-datetime' });
const networkMark = Decoration.mark({ class: 'cm-network' });
const commentMark = Decoration.mark({ class: 'cm-comment' });
const outputPrefixMark = Decoration.mark({ class: 'cm-output-prefix' });

function buildDecorations(view) {
    const builder = new RangeSetBuilder();
    const doc = view.state.doc;
    
    for (let i = 1; i <= doc.lines; i++) {
        const line = doc.line(i);
        const text = line.text;
        const from = line.from;
        
        // Output lines starting with "> "
        if (text.startsWith('> ')) {
            builder.add(from, from + 2, outputPrefixMark);
            // Check for errors in output
            if (text.includes('ERR:') || text.includes('error')) {
                builder.add(from + 2, line.to, errorMark);
            } else {
                builder.add(from + 2, line.to, resultMark);
            }
            continue;
        }
        
        // Comment lines starting with # or //
        if (text.trim().startsWith('#') || text.trim().startsWith('//')) {
            builder.add(from, line.to, commentMark);
            continue;
        }
        
        // Process tokens in the line
        let pos = 0;
        while (pos < text.length) {
            const remaining = text.slice(pos);
            let matched = false;
            
            // Line references \1, \2, etc.
            const refMatch = remaining.match(/^\\[0-9]+/);
            if (refMatch) {
                builder.add(from + pos, from + pos + refMatch[0].length, referenceMark);
                pos += refMatch[0].length;
                matched = true;
                continue;
            }
            
            // Currency with amount $1,234.56
            const currencyMatch = remaining.match(/^\$[\d,]+\.?\d*/);
            if (currencyMatch) {
                builder.add(from + pos, from + pos + currencyMatch[0].length, currencyMark);
                pos += currencyMatch[0].length;
                matched = true;
                continue;
            }
            
            // IP addresses and CIDR notation
            const ipMatch = remaining.match(/^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}(\/\d{1,2})?/);
            if (ipMatch) {
                builder.add(from + pos, from + pos + ipMatch[0].length, networkMark);
                pos += ipMatch[0].length;
                matched = true;
                continue;
            }
            
            // Time patterns like 14:30, 6:00 am
            const timeMatch = remaining.match(/^\d{1,2}:\d{2}(:\d{2})?(\s*(am|pm|AM|PM))?/);
            if (timeMatch) {
                builder.add(from + pos, from + pos + timeMatch[0].length, datetimeMark);
                pos += timeMatch[0].length;
                matched = true;
                continue;
            }
            
            // Numbers (including decimals and negative)
            const numMatch = remaining.match(/^-?\d+\.?\d*/);
            if (numMatch) {
                builder.add(from + pos, from + pos + numMatch[0].length, numberMark);
                pos += numMatch[0].length;
                matched = true;
                continue;
            }
            
            // Percentage
            if (remaining[0] === '%') {
                builder.add(from + pos, from + pos + 1, percentMark);
                pos += 1;
                matched = true;
                continue;
            }
            
            // Operators
            if (['+', '-', '*', '/', '^', '×', '÷', '='].includes(remaining[0])) {
                builder.add(from + pos, from + pos + 1, operatorMark);
                pos += 1;
                matched = true;
                continue;
            }
            
            // Functions
            const funcMatch = remaining.match(/^(sin|cos|tan|sqrt|abs|log|ln|exp|floor|ceil|round|min|max)\s*\(/i);
            if (funcMatch) {
                builder.add(from + pos, from + pos + funcMatch[1].length, functionMark);
                pos += funcMatch[1].length;
                matched = true;
                continue;
            }
            
            // Keywords
            const kwMatch = remaining.match(/^(now|today|yesterday|tomorrow|in|to|till|from|split|subnets?|hosts?|mask|wildcard|how\s+many|is)\b/i);
            if (kwMatch) {
                builder.add(from + pos, from + pos + kwMatch[0].length, keywordMark);
                pos += kwMatch[0].length;
                matched = true;
                continue;
            }
            
            // Date keywords
            const dateKwMatch = remaining.match(/^(days?|weeks?|months?|years?|hours?|minutes?|seconds?|Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec|January|February|March|April|June|July|August|September|October|November|December|Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday)\b/i);
            if (dateKwMatch) {
                builder.add(from + pos, from + pos + dateKwMatch[0].length, datetimeMark);
                pos += dateKwMatch[0].length;
                matched = true;
                continue;
            }
            
            if (!matched) {
                pos++;
            }
        }
    }
    
    return builder.finish();
}

const syntaxHighlighter = ViewPlugin.fromClass(class {
    constructor(view) {
        this.decorations = buildDecorations(view);
    }
    update(update) {
        if (update.docChanged || update.viewportChanged) {
            this.decorations = buildDecorations(update.view);
        }
    }
}, {
    decorations: v => v.decorations
});

// Custom Enter key handler - auto-append '=' if needed
function handleEnterKey(view) {
    const state = view.state;
    const pos = state.selection.main.head;
    const line = state.doc.lineAt(pos);
    const lineText = line.text;
    const cursorAtEnd = pos === line.to;
    
    // Check conditions:
    // 1. Cursor is at end of line
    // 2. Line is not empty
    // 3. Line is not a comment (starting with #)
    // 4. Line doesn't already end with '='
    if (cursorAtEnd && 
        lineText.trim().length > 0 && 
        !lineText.trim().startsWith('#') && 
        !lineText.trim().startsWith('//') &&
        !lineText.trimEnd().endsWith('=')) {
        
        // Insert ' =' at cursor, then newline
        view.dispatch({
            changes: { from: pos, insert: ' =\n' },
            selection: { anchor: pos + 3 },
        });
        return true;
    }
    
    // Default behavior: just insert newline
    return false;
}

// Custom keymap for Enter key
const customKeymap = keymap.of([
    {
        key: 'Enter',
        run: handleEnterKey,
    },
]);

// Initialize editor
function initEditor() {
    const container = document.getElementById('editor-container');
    
    const state = EditorState.create({
        doc: '',
        extensions: [
            lineNumbers(),
            highlightActiveLineGutter(),
            highlightActiveLine(),
            history(),
            customKeymap,
            keymap.of([...defaultKeymap, ...historyKeymap]),
            getCurrentTheme(),
            syntaxHighlighter,
            EditorView.updateListener.of((update) => {
                if (update.docChanged) {
                    onTextChanged();
                }
            }),
            EditorView.lineWrapping,
        ],
    });

    editor = new EditorView({
        state,
        parent: container,
    });

    // Load version
    GetVersion().then(version => {
        document.getElementById('version').textContent = `Version ${version}`;
    });

    // Set up keyboard shortcuts
    document.addEventListener('keydown', handleKeyboard);
    
    // Listen for system theme changes
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', (e) => {
        // Recreate editor with new theme
        const content = editor.state.doc.toString();
        const cursorPos = editor.state.selection.main.head;
        
        editor.destroy();
        
        const newState = EditorState.create({
            doc: content,
            extensions: [
                lineNumbers(),
                highlightActiveLineGutter(),
                highlightActiveLine(),
                history(),
                customKeymap,
                keymap.of([...defaultKeymap, ...historyKeymap]),
                e.matches ? darkTheme : lightTheme,
                syntaxHighlighter,
                EditorView.updateListener.of((update) => {
                    if (update.docChanged) {
                        onTextChanged();
                    }
                }),
                EditorView.lineWrapping,
            ],
        });
        
        editor = new EditorView({
            state: newState,
            parent: container,
        });
        
        // Restore cursor position
        editor.dispatch({
            selection: { anchor: Math.min(cursorPos, content.length) },
        });
    });
}

// Handle text changes with debounce
function onTextChanged() {
    // Skip if we're programmatically updating the editor
    if (isUpdatingEditor) {
        return;
    }
    
    if (debounceTimer) {
        clearTimeout(debounceTimer);
    }
    debounceTimer = setTimeout(() => {
        checkAndAdjustReferences();
    }, 150);
    
    // Schedule autosave
    scheduleAutosave();
}

// Check if line count changed and adjust references
async function checkAndAdjustReferences() {
    // Prevent re-entry when we dispatch programmatically
    if (isUpdatingEditor) {
        return;
    }
    
    const currentText = editor.state.doc.toString();
    const currentLineCount = currentText.split('\n').length;
    
    // If line count changed, adjust references
    if (previousLineCount > 0 && currentLineCount !== previousLineCount && previousText !== '') {
        try {
            const adjusted = await AdjustReferences(previousText, currentText);
            if (adjusted !== currentText) {
                // Set flag to prevent re-entry
                isUpdatingEditor = true;
                
                // Save cursor and scroll position
                const scrollTop = editor.scrollDOM.scrollTop;
                const scrollLeft = editor.scrollDOM.scrollLeft;
                const cursorPos = editor.state.selection.main.head;
                
                // Update with adjusted text
                editor.dispatch({
                    changes: { from: 0, to: editor.state.doc.length, insert: adjusted },
                    selection: { anchor: Math.min(cursorPos, adjusted.length) },
                });
                
                // Restore scroll
                requestAnimationFrame(() => {
                    editor.scrollDOM.scrollTop = scrollTop;
                    editor.scrollDOM.scrollLeft = scrollLeft;
                });
                
                // Update previous text to adjusted version
                previousText = adjusted;
                previousLineCount = adjusted.split('\n').length;
                
                // Evaluate the adjusted content (flag stays set)
                await evaluateContentInternal();
                
                // Clear flag after everything is done
                isUpdatingEditor = false;
                return;
            }
        } catch (err) {
            console.error('Adjust references error:', err);
            isUpdatingEditor = false;
        }
    }
    
    // Update previous text tracking
    previousText = currentText;
    previousLineCount = currentLineCount;
    
    // Evaluate normally
    evaluateContent();
}

// Schedule autosave after delay
function scheduleAutosave() {
    if (autosaveTimer) {
        clearTimeout(autosaveTimer);
    }
    autosaveTimer = setTimeout(() => {
        performAutosave();
    }, AUTOSAVE_DELAY);
}

// Perform autosave if we have a current file
async function performAutosave() {
    if (currentFile) {
        try {
            await AutoSave(currentFile, editor.state.doc.toString());
        } catch (err) {
            console.error('Autosave error:', err);
        }
    }
}

// Evaluate content and update display (sets flag to prevent re-entry)
async function evaluateContent() {
    isUpdatingEditor = true;
    try {
        await evaluateContentInternal();
    } finally {
        isUpdatingEditor = false;
    }
}

// Internal evaluate function (does not manage flag)
async function evaluateContentInternal() {
    const text = editor.state.doc.toString();
    try {
        const results = await Evaluate(text);
        
        // Build new content from results
        const newLines = results.map(r => r.output);
        const newText = newLines.join('\n');
        
        // Only update if different (to avoid cursor jump)
        if (newText !== text) {
            // Save scroll position and cursor line/column
            const scrollTop = editor.scrollDOM.scrollTop;
            const scrollLeft = editor.scrollDOM.scrollLeft;
            const cursorPos = editor.state.selection.main.head;
            const cursorLine = editor.state.doc.lineAt(cursorPos);
            const lineNumber = cursorLine.number;
            const columnOffset = cursorPos - cursorLine.from;
            
            editor.dispatch({
                changes: { from: 0, to: editor.state.doc.length, insert: newText },
            });
            
            // Restore cursor position based on line number
            const newDoc = editor.state.doc;
            if (lineNumber <= newDoc.lines) {
                const newLine = newDoc.line(lineNumber);
                const newPos = newLine.from + Math.min(columnOffset, newLine.length);
                editor.dispatch({
                    selection: { anchor: newPos },
                });
            } else {
                // If line doesn't exist, go to end
                editor.dispatch({
                    selection: { anchor: newText.length },
                });
            }
            
            // Restore scroll position after update
            requestAnimationFrame(() => {
                editor.scrollDOM.scrollTop = scrollTop;
                editor.scrollDOM.scrollLeft = scrollLeft;
            });
            
            // Update previous text to the evaluated result
            previousText = newText;
            previousLineCount = newText.split('\n').length;
        }
    } catch (err) {
        console.error('Evaluation error:', err);
    }
}

// Keyboard shortcuts
function handleKeyboard(e) {
    // Ctrl+O - Open
    if (e.ctrlKey && e.key === 'o') {
        e.preventDefault();
        openFile();
    }
    // Ctrl+S - Save
    if (e.ctrlKey && e.key === 's') {
        e.preventDefault();
        saveFile();
    }
    // Ctrl+Shift+S - Save As
    if (e.ctrlKey && e.shiftKey && e.key === 'S') {
        e.preventDefault();
        saveFileAs();
    }
    // Ctrl+N - New
    if (e.ctrlKey && e.key === 'n') {
        e.preventDefault();
        newFile();
    }
    // Ctrl+C - Smart Copy (replace refs with values)
    if (e.ctrlKey && e.key === 'c') {
        e.preventDefault();
        smartCopy();
    }
}

// File operations
async function openFile() {
    try {
        const path = await OpenFileDialog();
        if (path) {
            await openFilePath(path);
        }
    } catch (err) {
        console.error('Open error:', err);
    }
}

async function openFilePath(path) {
    try {
        const content = await ReadFile(path);
        editor.dispatch({
            changes: { from: 0, to: editor.state.doc.length, insert: content },
        });
        currentFile = path;
        updateFileName();
        evaluateContent();
        await AddRecentFile(path);
    } catch (err) {
        console.error('Open error:', err);
    }
}

async function saveFile() {
    if (currentFile) {
        try {
            await WriteFile(currentFile, editor.state.doc.toString());
            await AddRecentFile(currentFile);
        } catch (err) {
            console.error('Save error:', err);
        }
    } else {
        saveFileAs();
    }
}

async function saveFileAs() {
    try {
        const path = await SaveFileDialog();
        if (path) {
            await WriteFile(path, editor.state.doc.toString());
            currentFile = path;
            updateFileName();
            await AddRecentFile(path);
        }
    } catch (err) {
        console.error('Save As error:', err);
    }
}

function newFile() {
    editor.dispatch({
        changes: { from: 0, to: editor.state.doc.length, insert: '' },
    });
    currentFile = '';
    updateFileName();
}

function updateFileName() {
    const name = currentFile ? currentFile.split('/').pop() : 'Untitled';
    document.getElementById('file-name').textContent = name;
}

// Adjust line references in snippet based on insertion line
function adjustSnippetReferences(snippet, insertionLine) {
    // Replace \N references with adjusted line numbers
    // \1 -> \(insertionLine), \2 -> \(insertionLine+1), etc.
    return snippet.replace(/\\(\d+)/g, (match, num) => {
        const originalRef = parseInt(num, 10);
        const adjustedRef = originalRef + insertionLine - 1;
        return '\\' + adjustedRef;
    });
}

// Insert snippet at cursor
async function insertSnippet(snippet) {
    const pos = editor.state.selection.main.head;
    const line = editor.state.doc.lineAt(pos);
    const insertionLine = line.number;
    
    // Adjust line references in the snippet
    const adjustedSnippet = adjustSnippetReferences(snippet, insertionLine);
    
    // Insert the snippet
    editor.dispatch({
        changes: { from: pos, insert: adjustedSnippet },
    });
    
    // Evaluate the content
    await evaluateContent();
    
    // After evaluation, move cursor to end of document (next line after snippet results)
    const docLength = editor.state.doc.length;
    editor.dispatch({
        selection: { anchor: docLength },
    });
    editor.focus();
}

// Show manual dialog
function showManual() {
    const manual = `
# SmartCalc Manual

## Basic Usage
Type mathematical expressions followed by = to calculate results.

## Supported Operations
- **Addition**: +
- **Subtraction**: -
- **Multiplication**: * or x
- **Division**: /
- **Power**: ^
- **Parentheses**: ( )

## Percentages
- 100 + 20% = 120 (adds 20% of 100)
- 100 - 20% = 80 (subtracts 20% of 100)

## Currency
- Prefix numbers with $ for currency formatting
- Supports thousands separators: $1,500.00

## Line References
- Use \\1, \\2, etc. to reference results from previous lines

## Functions
- sin(), cos(), tan()
- sqrt(), abs()
- log(), ln()

## Date/Time
- **Current time**: now, now(), today, today()
- **Time in city**: now in Seattle, now in New York
- **Time conversion**: 6:00 am Seattle in Kiev
- **Date arithmetic**: today() + 30 days, \\1 - 1 week
- **Duration conversion**: 861.5 hours in days
- **Date range**: Dec 6 till March 11

## Network/IP Calculations
- **Subnet info**: 10.100.0.0/24
- **Split by count**: split 10.100.0.0/16 to 4 subnets
- **Host count**: how many hosts in 10.100.0.0/28
- **Subnet mask**: mask for /24, wildcard for /24
- **IP in range**: is 10.100.0.50 in 10.100.0.0/24
`;
    alert(manual);
}

// Show about dialog
function showAbout() {
    GetVersion().then(version => {
        alert(`SmartCalc ${version}

A powerful calculator with support for:
• Multi-line expressions
• Line references
• Currency formatting
• Mathematical functions
• Date/Time calculations
• Network/IP subnet calculations`);
    });
}

// Smart copy - replace line references with actual values
async function smartCopy() {
    const selection = editor.state.selection.main;
    let textToCopy;
    
    if (selection.empty) {
        // No selection - copy entire document
        textToCopy = editor.state.doc.toString();
    } else {
        // Copy selected text
        textToCopy = editor.state.sliceDoc(selection.from, selection.to);
    }
    
    // Replace line references with actual values
    const resolvedText = await CopyWithResolvedRefs(textToCopy);
    
    // Copy to clipboard
    await navigator.clipboard.writeText(resolvedText);
}

// Set up menu event listeners
function setupMenuEvents() {
    EventsOn('menu:new', newFile);
    EventsOn('menu:open', openFile);
    EventsOn('menu:save', saveFile);
    EventsOn('menu:saveAs', saveFileAs);
    EventsOn('menu:openRecent', openFilePath);
    EventsOn('menu:cut', () => document.execCommand('cut'));
    EventsOn('menu:copy', smartCopy);
    EventsOn('menu:paste', () => document.execCommand('paste'));
    EventsOn('menu:snippet', insertSnippet);
    EventsOn('menu:manual', showManual);
    EventsOn('menu:about', showAbout);
}

// Load last file on startup
async function loadLastFile() {
    try {
        const lastFile = await GetLastFile();
        if (lastFile) {
            await openFilePath(lastFile);
        }
    } catch (err) {
        console.error('Load last file error:', err);
    }
}

// Initialize on load
document.addEventListener('DOMContentLoaded', () => {
    initEditor();
    setupMenuEvents();
    loadLastFile();
});
