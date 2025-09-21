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

    // Delete Memory command
    const deleteMemory = vscode.commands.registerCommand('contextmemory.deleteMemory', async (memoryId?: string) => {
        try {
            let targetMemoryId = memoryId;
            
            if (!targetMemoryId) {
                // Get memory ID from user input
                const memories = await cmctlService.listMemories();
                const items = memories.map(memory => ({
                    label: memory.name,
                    description: `${Object.entries(memory.labels).map(([k, v]) => `${k}=${v}`).join(', ')}`,
                    detail: memory.id,
                    memoryId: memory.id
                }));
                
                const selected = await vscode.window.showQuickPick(items, {
                    placeHolder: 'Select memory to delete',
                    matchOnDescription: true,
                    matchOnDetail: true
                });
                
                if (!selected) {
                    return;
                }
                targetMemoryId = selected.memoryId;
            }

            // Confirm deletion
            const confirmAction = await vscode.window.showWarningMessage(
                `Are you sure you want to delete this memory?`,
                { modal: true },
                'Delete'
            );

            if (confirmAction === 'Delete') {
                await cmctlService.deleteMemory(targetMemoryId, true); // force = true to skip CLI confirmation
                vscode.window.showInformationMessage('Memory deleted successfully');
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to delete memory: ${error.message}`);
        }
    });

    // Delete Memories by Labels command
    const deleteMemoriesByLabels = vscode.commands.registerCommand('contextmemory.deleteMemoriesByLabels', async () => {
        try {
            const labelSelector = await vscode.window.showInputBox({
                placeHolder: 'Enter label selector (e.g., type=test,project=demo)',
                prompt: 'Delete all memories matching these labels'
            });

            if (!labelSelector) {
                return;
            }

            // Parse label selector into the correct format
            const labels: Record<string, string> = {};
            labelSelector.split(',').forEach(labelPair => {
                const [key, value] = labelPair.split('=', 2);
                if (key && value) {
                    labels[key.trim()] = value.trim();
                }
            });

            // Preview matching memories
            const searchResp = await cmctlService.searchMemories({ labels, limit: 100 });
            const count = searchResp.memories?.length || 0;

            if (count === 0) {
                vscode.window.showInformationMessage('No memories found matching the label selector');
                return;
            }

            // Confirm deletion
            const confirmAction = await vscode.window.showWarningMessage(
                `This will delete ${count} memories matching "${labelSelector}". This cannot be undone!`,
                { modal: true },
                'Delete All'
            );

            if (confirmAction === 'Delete All') {
                await cmctlService.deleteMemoriesByLabels(labelSelector, true);
                vscode.window.showInformationMessage(`Successfully deleted ${count} memories`);
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to delete memories: ${error.message}`);
        }
    });

    // Delete All Memories command
    const deleteAllMemories = vscode.commands.registerCommand('contextmemory.deleteAllMemories', async () => {
        try {
            const memories = await cmctlService.listMemories();
            const count = memories.length;

            if (count === 0) {
                vscode.window.showInformationMessage('No memories to delete');
                return;
            }

            // Confirm deletion with strong warning
            const confirmAction = await vscode.window.showWarningMessage(
                `⚠️ DANGER: This will delete ALL ${count} memories permanently! This cannot be undone!`,
                { modal: true },
                'I understand, delete everything'
            );

            if (confirmAction === 'I understand, delete everything') {
                await cmctlService.deleteAllMemories(true);
                vscode.window.showInformationMessage(`Successfully deleted all ${count} memories`);
                memoryTreeProvider.refresh();
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to delete all memories: ${error.message}`);
        }
    });

    // Open Configuration command
    const openConfig = vscode.commands.registerCommand('contextmemory.openConfig', async () => {
        try {
            const config = await cmctlService.getConfig();
            
            // Create a configuration editor interface
            const items = [
                {
                    label: '🛠️ CLI Path',
                    description: config.cliPath || 'cmctl',
                    detail: 'Path to the cmctl CLI binary',
                    configKey: 'cliPath',
                    type: 'string'
                },
                {
                    label: '📁 Storage Directory',
                    description: config.storageDir || '~/.contextmemory',
                    detail: 'Directory where memories are stored',
                    configKey: 'storageDir',
                    type: 'string'
                },
                {
                    label: '🔗 Provider',
                    description: config.provider || 'file',
                    detail: 'Storage provider (file, s3, gcs, remote)',
                    configKey: 'provider',
                    type: 'enum',
                    enumValues: ['file', 's3', 'gcs', 'remote']
                },
                {
                    label: '🔊 Verbosity Level',
                    description: `${config.verbosity || 1}`,
                    detail: 'CLI verbosity (0=quiet, 1=normal, 2=verbose)',
                    configKey: 'verbosity',
                    type: 'number',
                    min: 0,
                    max: 2
                },
                {
                    label: '🤖 Auto Suggest Labels',
                    description: config.autoSuggestLabels ? 'Enabled' : 'Disabled',
                    detail: 'Automatically suggest labels when creating memories',
                    configKey: 'autoSuggestLabels',
                    type: 'boolean'
                },
                {
                    label: '🏷️ Default Labels',
                    description: Array.isArray(config.defaultLabels) ? config.defaultLabels.join(', ') : 'None',
                    detail: 'Default labels to apply to new memories',
                    configKey: 'defaultLabels',
                    type: 'array'
                },
                {
                    label: '🆔 Show Memory IDs',
                    description: config.showMemoryIds ? 'Enabled' : 'Disabled',
                    detail: 'Show memory IDs when listing memories',
                    configKey: 'showMemoryIds',
                    type: 'boolean'
                }
            ];

            const selected = await vscode.window.showQuickPick(items, {
                placeHolder: 'Select a setting to modify',
                matchOnDescription: true,
                matchOnDetail: true
            });

            if (selected) {
                await editConfigValue(selected, cmctlService);
            }
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to open configuration: ${error.message}`);
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
        health,
        deleteMemory,
        deleteMemoriesByLabels,
        deleteAllMemories,
        openConfig
    );
}

/**
 * Prompt user for memory creation details
 */
async function promptForMemoryCreation(
    initialContent: string = '',
    defaultLabels: Record<string, string> = {}
): Promise<CreateMemoryRequest | undefined> {
    // STEP 1: Get content first (this is the core of our design)
    let content = initialContent;
    if (!content) {
        content = await vscode.window.showInputBox({
            prompt: 'Enter memory content',
            placeHolder: 'Paste your code, notes, or chat content here...',
            validateInput: (value) => {
                if (!value || value.trim().length === 0) {
                    return 'Memory content is required';
                }
                return null;
            }
        }) || '';
    }

    // Final validation - ensure content is not empty
    if (!content || content.trim().length === 0) {
        return undefined;
    }

    // STEP 2: AI suggests name based on content
    const suggestedName = await suggestNameFromContent(content);
    
    // STEP 3: Show AI suggestion, allow user override
    const name = await vscode.window.showInputBox({
        prompt: 'Memory name (AI suggested - edit if needed)',
        placeHolder: 'AI will suggest a name based on your content...',
        value: suggestedName, // Pre-fill with AI suggestion
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

    // STEP 4: AI suggests labels based on content and name  
    if (autoSuggestLabels) {
        labels = { ...labels, ...await suggestLabels(content, name) };
    }

    // STEP 5: Show AI-suggested labels, allow user override
    const labelsString = Object.entries(labels)
        .map(([key, value]) => `${key}=${value}`)
        .join(',');
    
    const finalLabelsString = await vscode.window.showInputBox({
        prompt: 'Labels (AI suggested - edit if needed)',
        placeHolder: 'type=note,language=typescript,source=extension',
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
 * AI-powered name suggestion based on content
 */
async function suggestNameFromContent(content: string): Promise<string> {
    // Strategy 1: Extract from code entities (functions, classes, etc.)
    const codeEntityName = extractCodeEntityName(content);
    if (codeEntityName) {
        return codeEntityName;
    }
    
    // Strategy 2: Extract from headers/titles
    const headerName = extractHeaderName(content);
    if (headerName) {
        return headerName;
    }
    
    // Strategy 3: Extract key concepts
    const conceptualName = extractConceptualName(content);
    if (conceptualName) {
        return conceptualName;
    }
    
    // Strategy 4: Generate from first meaningful line
    const firstLineName = extractFromFirstLine(content);
    if (firstLineName) {
        return firstLineName;
    }
    
    // Fallback: Use timestamp
    return `Memory ${new Date().toISOString().split('T')[0]}`;
}

/**
 * Extract name from code entities (functions, classes, variables)
 */
function extractCodeEntityName(content: string): string | null {
    const codePatterns = [
        // Function definitions
        /(?:function|def|fn)\s+([a-zA-Z_][a-zA-Z0-9_]*)/i,
        // Class definitions  
        /(?:class|struct|interface)\s+([a-zA-Z_][a-zA-Z0-9_]*)/i,
        // Variable/const declarations
        /(?:const|let|var|final)\s+([a-zA-Z_][a-zA-Z0-9_]*)/i,
        // Method definitions
        /([a-zA-Z_][a-zA-Z0-9_]*)\s*\([^)]*\)\s*[{=:]/
    ];
    
    for (const pattern of codePatterns) {
        const match = content.match(pattern);
        if (match && match[1]) {
            return humanizeCodeName(match[1]);
        }
    }
    
    return null;
}

/**
 * Extract name from headers and titles
 */
function extractHeaderName(content: string): string | null {
    const headerPatterns = [
        /^#{1,6}\s+(.+)$/m,        // Markdown headers
        /^(.+)\n[=-]+$/m,         // Underlined headers  
        /^[*-]\s*(.+)$/m,         // Bullet points
        /<h[1-6][^>]*>([^<]+)<\/h[1-6]>/i  // HTML headers
    ];
    
    for (const pattern of headerPatterns) {
        const match = content.match(pattern);
        if (match && match[1]) {
            return cleanTitle(match[1]);
        }
    }
    
    return null;
}

/**
 * Extract conceptual name from key terms
 */
function extractConceptualName(content: string): string | null {
    const words = content.match(/\b[a-zA-Z]{3,}\b/g);
    if (!words) return null;
    
    // Get word frequencies
    const wordCount: Record<string, number> = {};
    words.forEach(word => {
        const lowerWord = word.toLowerCase();
        if (!isStopWord(lowerWord)) {
            wordCount[lowerWord] = (wordCount[lowerWord] || 0) + 1;
        }
    });
    
    // Find most important words
    const sortedWords = Object.entries(wordCount)
        .sort(([,a], [,b]) => b - a)
        .slice(0, 3)
        .map(([word]) => word);
        
    if (sortedWords.length > 0) {
        return titleCase(sortedWords.slice(0, 2).join(' '));
    }
    
    return null;
}

/**
 * Extract from first meaningful line
 */
function extractFromFirstLine(content: string): string | null {
    const lines = content.split('\n').map(line => line.trim()).filter(line => line.length > 0);
    
    for (const line of lines.slice(0, 3)) {
        // Skip obvious noise
        if (line.match(/^[/\-*#+\s]*$/) || line.length < 5) continue;
        
        // Clean and truncate
        const cleaned = cleanTitle(line);
        if (cleaned && cleaned.length > 5) {
            return cleaned.length > 50 ? cleaned.substring(0, 47) + '...' : cleaned;
        }
    }
    
    return null;
}

/**
 * Convert code names to human readable format
 */
function humanizeCodeName(codeName: string): string {
    return codeName
        .replace(/([a-z])([A-Z])/g, '$1 $2')  // camelCase to words
        .replace(/[_-]/g, ' ')                // underscores/hyphens to spaces
        .replace(/\b\w/g, l => l.toUpperCase()) // Title case
        .trim();
}

/**
 * Clean titles and remove noise
 */
function cleanTitle(title: string): string {
    return title
        .replace(/^[#*\-+=\s]+/, '')          // Remove leading symbols
        .replace(/[#*\-+=\s]+$/, '')          // Remove trailing symbols  
        .replace(/\s+/g, ' ')                 // Normalize spaces
        .trim();
}

/**
 * Check if word is a stop word
 */
function isStopWord(word: string): boolean {
    const stopWords = ['the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with', 'by', 'is', 'are', 'was', 'were', 'this', 'that', 'these', 'those'];
    return stopWords.includes(word.toLowerCase());
}

/**
 * Convert to title case
 */
function titleCase(text: string): string {
    return text.replace(/\b\w/g, l => l.toUpperCase());
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

/**
 * Edit a configuration value
 */
async function editConfigValue(configItem: any, cmctlService: any): Promise<void> {
    let newValue: any;

    switch (configItem.type) {
        case 'string':
            newValue = await vscode.window.showInputBox({
                prompt: `Enter new value for ${configItem.label}`,
                value: configItem.description,
                placeHolder: configItem.detail
            });
            break;

        case 'number':
            const numberInput = await vscode.window.showInputBox({
                prompt: `Enter new value for ${configItem.label} (${configItem.min}-${configItem.max})`,
                value: configItem.description,
                validateInput: (value) => {
                    const num = parseInt(value);
                    if (isNaN(num)) {
                        return 'Please enter a valid number';
                    }
                    if (num < (configItem.min || 0) || num > (configItem.max || 10)) {
                        return `Value must be between ${configItem.min || 0} and ${configItem.max || 10}`;
                    }
                    return null;
                }
            });
            if (numberInput !== undefined) {
                newValue = parseInt(numberInput);
            }
            break;

        case 'boolean':
            const booleanChoice = await vscode.window.showQuickPick(['true', 'false'], {
                placeHolder: `Select value for ${configItem.label}`
            });
            if (booleanChoice !== undefined) {
                newValue = booleanChoice === 'true';
            }
            break;

        case 'enum':
            newValue = await vscode.window.showQuickPick(configItem.enumValues || [], {
                placeHolder: `Select value for ${configItem.label}`
            });
            break;

        case 'array':
            const arrayInput = await vscode.window.showInputBox({
                prompt: `Enter comma-separated values for ${configItem.label}`,
                value: Array.isArray(configItem.description) ? configItem.description.join(', ') : configItem.description,
                placeHolder: 'value1, value2, value3'
            });
            if (arrayInput !== undefined) {
                newValue = arrayInput.split(',').map(s => s.trim()).filter(s => s.length > 0);
            }
            break;
    }

    if (newValue !== undefined) {
        try {
            await cmctlService.updateConfig({ [configItem.configKey]: newValue });
            vscode.window.showInformationMessage(`Updated ${configItem.label} to: ${newValue}`);
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to update configuration: ${error.message}`);
        }
    }
}

