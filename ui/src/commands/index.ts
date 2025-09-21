import * as vscode from 'vscode';
import { CMCtlService } from '../services/cmctlService';
import { MemoryTreeDataProvider } from '../providers/memoryTreeProvider';
import { CreateMemoryRequest } from '../types/memory';

/**
 * Register all extension commands
 */
export function registerCommands(
    context: vscode.ExtensionContext,
    cmctlService: CMCtlService,
    memoryTreeProvider: MemoryTreeDataProvider
) {
    // Create Memory command
    const createMemory = vscode.commands.registerCommand('contextmemory.createMemory', async () => {
        try {
            const memory = await promptForMemoryCreation();
            if (memory) {
                const createdMemory = await cmctlService.createMemory(memory);
                vscode.window.showInformationMessage(`Memory "${createdMemory.name}" created successfully`);
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to create memory: ${error.message}`);
        }
    });

    // Create Memory from Selection command
    const createMemoryFromSelection = vscode.commands.registerCommand('contextmemory.createMemoryFromSelection', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor || !editor.selection || editor.selection.isEmpty) {
            vscode.window.showWarningMessage('Please select some text first');
            return;
        }

        const selectedText = editor.document.getText(editor.selection);
        const fileName = editor.document.fileName;
        const language = editor.document.languageId;

        try {
            const memory = await promptForMemoryCreation(selectedText, {
                type: 'code',
                language: language,
                source: 'selection',
                file: fileName.split('/').pop() || ''
            });

            if (memory) {
                const createdMemory = await cmctlService.createMemory(memory);
                vscode.window.showInformationMessage(`Code memory "${createdMemory.name}" created successfully`);
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to create memory from selection: ${error.message}`);
        }
    });

    // Create Memory from Chat command
    const createMemoryFromChat = vscode.commands.registerCommand('contextmemory.createMemoryFromChat', async () => {
        const editor = vscode.window.activeTextEditor;
        if (!editor) {
            vscode.window.showWarningMessage('Please open a chat file first');
            return;
        }

        const content = editor.document.getText();
        const fileName = editor.document.fileName;

        try {
            const memory = await promptForMemoryCreation(content, {
                type: 'chat',
                source: 'chat-export',
                file: fileName.split('/').pop() || ''
            });

            if (memory) {
                const createdMemory = await cmctlService.createMemory(memory);
                vscode.window.showInformationMessage(`Chat memory "${createdMemory.name}" created successfully`);
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to create memory from chat: ${error.message}`);
        }
    });

    // Search Memories command
    const searchMemories = vscode.commands.registerCommand('contextmemory.searchMemories', async () => {
        const query = await vscode.window.showInputBox({
            prompt: 'Enter search query',
            placeHolder: 'Search memories...'
        });

        if (query !== undefined) {
            try {
                const results = await cmctlService.searchMemories({ query });
                await showSearchResults(results.memories);
            } catch (error: any) {
                vscode.window.showErrorMessage(`Search failed: ${error.message}`);
            }
        }
    });

    // List Memories command
    const listMemories = vscode.commands.registerCommand('contextmemory.listMemories', async () => {
        try {
            const memories = await cmctlService.listMemories();
            await showSearchResults(memories);
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to list memories: ${error.message}`);
        }
    });

    // Refresh Memories command
    const refreshMemories = vscode.commands.registerCommand('contextmemory.refreshMemories', () => {
        memoryTreeProvider.refresh();
        vscode.window.showInformationMessage('Memories refreshed');
    });

    // Open Memory command
    const openMemory = vscode.commands.registerCommand('contextmemory.openMemory', async (memoryId: string) => {
        try {
            const memory = await cmctlService.getMemory(memoryId);
            await showMemoryInEditor(memory);
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to open memory: ${error.message}`);
        }
    });

    // Health Check command
    const health = vscode.commands.registerCommand('contextmemory.health', async () => {
        try {
            await cmctlService.checkHealth();
            const version = await cmctlService.getVersion();
            const info = await cmctlService.getStorageInfo();
            
            vscode.window.showInformationMessage(
                `ContextMemory v${version} is healthy\n` +
                `Storage: ${info.storageDir}\n` +
                `Memories: ${info.memoriesCount}\n` +
                `Size: ${info.totalSize}`
            );
        } catch (error: any) {
            vscode.window.showErrorMessage(`Health check failed: ${error.message}`);
            cmctlService.showOutputChannel();
        }
    });

    // Add all commands to subscriptions
    context.subscriptions.push(
        createMemory,
        createMemoryFromSelection,
        createMemoryFromChat,
        searchMemories,
        listMemories,
        refreshMemories,
        openMemory,
        health
    );
}

/**
 * Prompt user for memory creation details
 */
async function promptForMemoryCreation(
    initialContent: string = '',
    defaultLabels: Record<string, string> = {}
): Promise<CreateMemoryRequest | undefined> {
    // Get memory name
    const name = await vscode.window.showInputBox({
        prompt: 'Enter memory name',
        placeHolder: 'My awesome memory...',
        validateInput: (value) => {
            if (!value || value.trim().length === 0) {
                return 'Memory name is required';
            }
            return null;
        }
    });

    if (!name) {
        return undefined;
    }

    // Get content if not provided
    let content = initialContent;
    if (!content) {
        content = await vscode.window.showInputBox({
            prompt: 'Enter memory content',
            placeHolder: 'Content of your memory...'
        }) || '';
    }

    // Get labels
    const config = vscode.workspace.getConfiguration('contextmemory');
    const autoSuggestLabels = config.get<boolean>('autoSuggestLabels', true);
    const defaultLabelsFromConfig = config.get<string[]>('defaultLabels', []);
    
    let labels = { ...defaultLabels };
    
    // Add default labels from config
    defaultLabelsFromConfig.forEach(label => {
        if (label.includes('=')) {
            const [key, value] = label.split('=', 2);
            labels[key.trim()] = value.trim();
        }
    });

    // Auto-suggest labels if enabled
    if (autoSuggestLabels) {
        labels = { ...labels, ...await suggestLabels(content, name) };
    }

    // Allow user to modify labels
    const labelsString = Object.entries(labels)
        .map(([key, value]) => `${key}=${value}`)
        .join(',');
    
    const finalLabelsString = await vscode.window.showInputBox({
        prompt: 'Enter labels (key=value,key2=value2)',
        placeHolder: 'type=note,language=typescript',
        value: labelsString
    });

    if (finalLabelsString !== undefined) {
        labels = {};
        if (finalLabelsString.trim()) {
            finalLabelsString.split(',').forEach(labelPair => {
                const [key, value] = labelPair.split('=', 2);
                if (key && value) {
                    labels[key.trim()] = value.trim();
                }
            });
        }
    }

    return {
        name: name.trim(),
        content: content.trim(),
        labels
    };
}

/**
 * Auto-suggest labels based on content and name
 */
async function suggestLabels(content: string, name: string): Promise<Record<string, string>> {
    const labels: Record<string, string> = {};
    
    // Detect language
    const language = detectLanguage(content);
    if (language) {
        labels.language = language;
    }
    
    // Detect type
    const type = detectType(content, name);
    if (type) {
        labels.type = type;
    }
    
    // Add source
    labels.source = 'extension';
    
    // Add creation date
    labels.created = new Date().toISOString().split('T')[0];
    
    return labels;
}

/**
 * Detect programming language from content
 */
function detectLanguage(content: string): string | null {
    const patterns: Record<string, RegExp[]> = {
        typescript: [/import.*from/, /interface\s+\w+/, /type\s+\w+\s*=/, /async\s+function/],
        javascript: [/console\.log/, /function\s+\w+/, /const\s+\w+\s*=/, /require\(/],
        python: [/def\s+\w+/, /import\s+\w+/, /from\s+\w+\s+import/, /print\(/],
        go: [/func\s+\w+/, /package\s+\w+/, /import\s+/, /fmt\.Print/],
        rust: [/fn\s+\w+/, /let\s+mut/, /use\s+/, /println!/],
        java: [/public\s+class/, /import\s+java/, /System\.out\.print/],
        cpp: [/#include\s*</, /std::/, /cout\s*<</, /namespace\s+/],
        sql: [/SELECT\s+/, /FROM\s+/, /WHERE\s+/, /INSERT\s+INTO/i],
        yaml: [/^---/, /^\s*\w+:\s*/, /^\s*-\s+/],
        json: [/^\s*{/, /"\w+"\s*:/, /^\s*\[/]
    };
    
    for (const [lang, regexes] of Object.entries(patterns)) {
        if (regexes.some(regex => regex.test(content))) {
            return lang;
        }
    }
    
    return null;
}

/**
 * Detect memory type from content and name
 */
function detectType(content: string, name: string): string | null {
    const nameWords = name.toLowerCase();
    const contentWords = content.toLowerCase();
    
    if (nameWords.includes('meeting') || contentWords.includes('meeting')) {
        return 'meeting';
    }
    if (nameWords.includes('bug') || nameWords.includes('issue') || contentWords.includes('error')) {
        return 'bug';
    }
    if (nameWords.includes('feature') || nameWords.includes('enhancement')) {
        return 'feature';
    }
    if (nameWords.includes('chat') || nameWords.includes('conversation')) {
        return 'chat';
    }
    if (nameWords.includes('code') || nameWords.includes('function') || nameWords.includes('snippet')) {
        return 'code';
    }
    if (nameWords.includes('note') || nameWords.includes('reminder')) {
        return 'note';
    }
    
    return 'manual';
}

/**
 * Show search results in a QuickPick
 */
async function showSearchResults(memories: any[]) {
    if (memories.length === 0) {
        vscode.window.showInformationMessage('No memories found');
        return;
    }
    
    const items = memories.map(memory => ({
        label: memory.name,
        description: Object.entries(memory.labels || {})
            .map(([key, value]) => `${key}=${value}`)
            .join(', '),
        detail: memory.content.substring(0, 100) + (memory.content.length > 100 ? '...' : ''),
        memory
    }));
    
    const selected = await vscode.window.showQuickPick(items, {
        placeHolder: `Found ${memories.length} memories`,
        matchOnDescription: true,
        matchOnDetail: true
    });
    
    if (selected) {
        await showMemoryInEditor(selected.memory);
    }
}

/**
 * Show memory content in a new editor
 */
async function showMemoryInEditor(memory: any) {
    const content = [
        `# ${memory.name}`,
        '',
        `**ID:** ${memory.id}`,
        `**Created:** ${memory.created}`,
        `**Updated:** ${memory.updated}`,
        '',
        '**Labels:**',
        ...Object.entries(memory.labels || {}).map(([key, value]) => `- ${key}: ${value}`),
        '',
        '**Content:**',
        '',
        memory.content
    ].join('\n');
    
    const doc = await vscode.workspace.openTextDocument({
        content,
        language: 'markdown'
    });
    
    await vscode.window.showTextDocument(doc);
}
