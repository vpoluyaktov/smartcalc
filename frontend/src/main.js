import './style.css';
import { EditorView, basicSetup } from 'codemirror';
import { EditorState } from '@codemirror/state';
import { keymap } from '@codemirror/view';
import { defaultKeymap, history, historyKeymap } from '@codemirror/commands';
import { lineNumbers, highlightActiveLineGutter, highlightActiveLine } from '@codemirror/view';
import { Evaluate, GetVersion, OpenFileDialog, SaveFileDialog, ReadFile, WriteFile } from '../wailsjs/go/main/App';

let editor;
let currentFile = '';
let debounceTimer = null;

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
    if (debounceTimer) {
        clearTimeout(debounceTimer);
    }
    debounceTimer = setTimeout(() => {
        evaluateContent();
    }, 150);
}

// Evaluate content and update display
async function evaluateContent() {
    const text = editor.state.doc.toString();
    try {
        const results = await Evaluate(text);
        
        // Build new content from results
        const newLines = results.map(r => r.output);
        const newText = newLines.join('\n');
        
        // Only update if different (to avoid cursor jump)
        if (newText !== text) {
            const cursorPos = editor.state.selection.main.head;
            editor.dispatch({
                changes: { from: 0, to: editor.state.doc.length, insert: newText },
                selection: { anchor: Math.min(cursorPos, newText.length) },
            });
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
            const content = await ReadFile(path);
            editor.dispatch({
                changes: { from: 0, to: editor.state.doc.length, insert: content },
            });
            currentFile = path;
            updateFileName();
            evaluateContent();
        }
    } catch (err) {
        console.error('Open error:', err);
    }
}

async function saveFile() {
    if (currentFile) {
        try {
            await WriteFile(currentFile, editor.state.doc.toString());
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

// Initialize on load
document.addEventListener('DOMContentLoaded', initEditor);
