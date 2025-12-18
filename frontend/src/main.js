import './style.css';
import { EditorView, basicSetup } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { keymap } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view';
import { Evaluate, GetVersion, OpenFileDialog, SaveFileDialog, ReadFile, WriteFile, AddRecentFile, GetLastFile, AutoSave, AdjustReferences } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';

let editor;
let currentFile = '';
let debounceTimer = null;
let autosaveTimer = null;
let previousText = '';
let previousLineCount = 0;
let isUpdatingEditor = false; // Flag to prevent re-entry during programmatic updates
const AUTOSAVE_DELAY = 2000; // 2 seconds after last change

// Dark theme
const darkTheme = EditorView.theme({
    '&': {
        backgroundColor: '#0f172a',
        color: '#f8fafc',
    },
    '.cm-content': {
        caretColor: '#6366f1',
    },
    '.cm-cursor': {
        borderLeftColor: '#6366f1',
    },
    '&.cm-focused .cm-selectionBackground, .cm-selectionBackground': {
        backgroundColor: 'rgba(99, 102, 241, 0.3)',
    },
    '.cm-gutters': {
        backgroundColor: '#1e293b',
        color: '#64748b',
        borderRight: '1px solid #334155',
    },
    '.cm-activeLineGutter': {
        backgroundColor: '#334155',
        color: '#f8fafc',
    },
    '.cm-activeLine': {
        backgroundColor: 'rgba(51, 65, 85, 0.5)',
    },
}, { dark: true });

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
            keymap.of([...defaultKeymap, ...historyKeymap]),
            darkTheme,
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
            // Save scroll position and cursor
            const scrollTop = editor.scrollDOM.scrollTop;
            const scrollLeft = editor.scrollDOM.scrollLeft;
            const cursorPos = editor.state.selection.main.head;
            
            editor.dispatch({
                changes: { from: 0, to: editor.state.doc.length, insert: newText },
                selection: { anchor: Math.min(cursorPos, newText.length) },
            });
            
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

// Insert snippet at cursor
function insertSnippet(snippet) {
    const pos = editor.state.selection.main.head;
    editor.dispatch({
        changes: { from: pos, insert: snippet },
        selection: { anchor: pos + snippet.length },
    });
    evaluateContent();
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

// Set up menu event listeners
function setupMenuEvents() {
    EventsOn('menu:new', newFile);
    EventsOn('menu:open', openFile);
    EventsOn('menu:save', saveFile);
    EventsOn('menu:saveAs', saveFileAs);
    EventsOn('menu:openRecent', openFilePath);
    EventsOn('menu:cut', () => document.execCommand('cut'));
    EventsOn('menu:copy', () => document.execCommand('copy'));
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
