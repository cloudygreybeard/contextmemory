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


    // Create Memory from Chat command - Direct Cursor AI Pane integration
    const createMemoryFromChat = vscode.commands.registerCommand('contextmemory.createMemoryFromChat', async () => {
        try {
            // Show progress indicator
            await vscode.window.withProgress({
                location: vscode.ProgressLocation.Notification,
                title: "Capturing Cursor chat...",
                cancellable: false
            }, async (progress) => {
                progress.report({ increment: 20, message: "Accessing Cursor AI pane data..." });
                
                // Use the CLI command to import the latest chat directly from Cursor
                const result = await cmctlService.importCursorChat();
                
                progress.report({ increment: 80, message: "Creating memory..." });
                
                if (result.success) {
                    vscode.window.showInformationMessage(
                        `Chat memory "${result.memoryName}" captured successfully from Cursor AI pane!`
                    );
                    memoryTreeProvider.refresh();
                } else {
                    throw new Error(result.error || 'Failed to import chat');
                }
            });
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to capture chat from Cursor: ${error.message}`);
            
            // Provide helpful guidance
            const action = await vscode.window.showWarningMessage(
                'Could not access Cursor AI pane data. Make sure you have an active chat conversation.',
                'Try Again',
                'Open Troubleshooting'
            );
            
            if (action === 'Try Again') {
                vscode.commands.executeCommand('contextmemory.createMemoryFromChat');
            } else if (action === 'Open Troubleshooting') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory/docs/troubleshooting.md'));
            }
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
            const versionCheck = await cmctlService.checkVersionCompatibility();
            const info = await cmctlService.getStorageInfo();
            
            const compatibilityStatus = versionCheck.compatible ? 'Compatible' : 'Version Mismatch';
            const message = 
                `ContextMemory Health Check\n` +
                `CLI Version: ${versionCheck.cliVersion}\n` +
                `Extension Version: ${versionCheck.extensionVersion}\n` +
                `Compatibility: ${compatibilityStatus}\n` +
                `Storage: ${info.storageDir}\n` +
                `Memories: ${info.memoriesCount}\n` +
                `Size: ${info.totalSize}`;
                
            if (versionCheck.compatible) {
                vscode.window.showInformationMessage(message);
            } else {
                vscode.window.showWarningMessage(
                    `${message}\n\nWarning: ${versionCheck.reason}`,
                    'Update CLI',
                    'Check Docs'
                ).then(selection => {
                    if (selection === 'Update CLI') {
                        vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory#installation'));
                    } else if (selection === 'Check Docs') {
                        vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory'));
                    }
                });
            }
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
                `DANGER: This will delete ALL ${count} memories permanently! This cannot be undone!`,
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
                    label: 'CLI Path',
                    description: config.cliPath || 'cmctl',
                    detail: 'Path to the cmctl CLI binary',
                    configKey: 'cliPath',
                    type: 'string'
                },
                {
                    label: 'Storage Directory',
                    description: config.storageDir || '~/.contextmemory',
                    detail: 'Directory where memories are stored',
                    configKey: 'storageDir',
                    type: 'string'
                },
                {
                    label: 'Provider',
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
                    label: 'Default Labels',
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

    // Add essential commands to subscriptions - focused on chat capture
    context.subscriptions.push(
        createMemoryFromChat,
        refreshMemories,
        openMemory,
        health
    );
}

/**
 * Chat-specific memory creation with enhanced AI naming
 */
async function promptForChatMemoryCreation(
    content: string,
    defaultLabels: Record<string, string> = {}
): Promise<CreateMemoryRequest | undefined> {
    // Use enhanced chat-specific naming
    const suggestedName = await suggestChatNameFromContent(content);
    
    const name = await vscode.window.showInputBox({
        prompt: 'Chat Memory Name (AI suggested based on conversation)',
        placeHolder: 'AI analyzed the conversation to suggest this name...',
        value: suggestedName,
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

    // Get configuration for auto-labeling
    const config = vscode.workspace.getConfiguration('contextmemory');
    const defaultLabelsFromConfig = config.get<string[]>('defaultLabels', []);
    
    let labels = { ...defaultLabels };
    
    // Add default labels from config
    defaultLabelsFromConfig.forEach(label => {
        const [key, value] = label.split('=');
        if (key && value) {
            labels[key.trim()] = value.trim();
        }
    });

    // For chat, we focus on AI-suggested labels rather than user input
    const labelString = Object.entries(labels)
        .map(([key, value]) => `${key}=${value}`)
        .join(',');
    
    // Show labels for user review/override
    const finalLabels = await vscode.window.showInputBox({
        prompt: 'Labels (AI suggested based on chat analysis)',
        placeHolder: 'AI detected: programming language, activity, topic...',
        value: labelString,
        validateInput: (value) => {
            if (value && value.includes('=') && !value.match(/^[\w-]+=[\w-]+(,[\w-]+=[\w-]+)*$/)) {
                return 'Invalid format. Use: key1=value1,key2=value2';
            }
            return null;
        }
    });

    if (finalLabels === undefined) {
        return undefined;
    }

    // Parse final labels
    const parsedLabels: Record<string, string> = {};
    if (finalLabels && finalLabels.trim()) {
        const pairs = finalLabels.split(',');
        pairs.forEach(pair => {
            const [key, value] = pair.split('=');
            if (key && value) {
                parsedLabels[key.trim()] = value.trim();
            }
        });
    }

    return {
        name: name.trim(),
        content: content,
        labels: parsedLabels
    };
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
 * Enhanced chat-specific naming that understands conversation patterns
 */
async function suggestChatNameFromContent(content: string): Promise<string> {
    // Strategy 1: Look for explicit topics or subjects
    const topicName = extractChatTopic(content);
    if (topicName) {
        return topicName;
    }
    
    // Strategy 2: Extract from main technical concepts discussed
    const conceptName = extractTechnicalConcepts(content);
    if (conceptName) {
        return conceptName;
    }
    
    // Strategy 3: Use chat session pattern
    const sessionName = generateSessionName(content);
    return sessionName;
}

/**
 * Detect if content appears to be a chat conversation
 */
function isChatContent(content: string): boolean {
    const chatIndicators = [
        /\*\*User\*\*:/i,
        /\*\*Assistant\*\*:/i,
        /^User:/m,
        /^Assistant:/m,
        /^Human:/m,
        /^AI:/m,
        /^You:/m,
        /^Me:/m,
        /@\w+:/, // Mentions
        />\s*[A-Z]/, // Quote blocks
        /^[A-Z][a-z]+:\s/m // Name: pattern
    ];
    
    const indicatorCount = chatIndicators.reduce((count, pattern) => {
        return count + (pattern.test(content) ? 1 : 0);
    }, 0);
    
    // Also check for conversation-like patterns
    const messageCount = (content.match(/^[A-Za-z]+:\s/gm) || []).length;
    const hasCodeBlocks = /```/.test(content);
    const hasQuestions = (content.match(/\?/g) || []).length;
    
    return indicatorCount >= 2 || messageCount >= 3 || (hasCodeBlocks && hasQuestions >= 2);
}

/**
 * Extract metadata from chat content for better labeling
 */
function extractChatMetadata(content: string): Record<string, string> {
    const metadata: Record<string, string> = {};
    
    // Extract date if present
    const dateMatch = content.match(/\*\*Date\*\*:\s*([^\n]+)/i) || 
                     content.match(/Date:\s*([^\n]+)/i);
    if (dateMatch) {
        metadata.date = dateMatch[1].trim();
    } else {
        metadata.date = new Date().toISOString().split('T')[0];
    }
    
    // Extract topic/subject
    const topicMatch = content.match(/\*\*Topic\*\*:\s*([^\n]+)/i) ||
                      content.match(/\*\*Subject\*\*:\s*([^\n]+)/i) ||
                      content.match(/Topic:\s*([^\n]+)/i);
    if (topicMatch) {
        metadata.topic = topicMatch[1].trim().toLowerCase().replace(/\s+/g, '-');
    }
    
    // Detect programming languages mentioned
    const languagePatterns = /\b(javascript|typescript|python|java|go|rust|cpp|c\+\+|html|css|sql|bash|shell)\b/gi;
    const languages = content.match(languagePatterns);
    if (languages && languages.length > 0) {
        const uniqueLanguages = [...new Set(languages.map(l => l.toLowerCase()))];
        metadata.language = uniqueLanguages[0]; // Primary language
        if (uniqueLanguages.length > 1) {
            metadata.languages = uniqueLanguages.join(',');
        }
    }
    
    // Detect if it's about debugging, implementing, reviewing, etc.
    const activityPatterns = [
        { pattern: /\b(debug|debugging|bug|error|issue|problem)\b/gi, label: 'debugging' },
        { pattern: /\b(implement|implementation|create|build|develop)\b/gi, label: 'implementation' },
        { pattern: /\b(review|code review|feedback|suggestions)\b/gi, label: 'review' },
        { pattern: /\b(refactor|refactoring|improve|optimization)\b/gi, label: 'refactoring' },
        { pattern: /\b(test|testing|unit test|integration)\b/gi, label: 'testing' }
    ];
    
    for (const { pattern, label } of activityPatterns) {
        if (pattern.test(content)) {
            metadata.activity = label;
            break;
        }
    }
    
    return metadata;
}

/**
 * Extract chat topic from content
 */
function extractChatTopic(content: string): string | null {
    // Look for explicit topic declarations
    const topicPatterns = [
        /\*\*Topic\*\*:\s*([^\n]+)/i,
        /\*\*Subject\*\*:\s*([^\n]+)/i,
        /^#\s*(.+)$/m, // First markdown header
        /Topic:\s*([^\n]+)/i,
        /Subject:\s*([^\n]+)/i
    ];
    
    for (const pattern of topicPatterns) {
        const match = content.match(pattern);
        if (match && match[1]) {
            return cleanTitle(match[1]);
        }
    }
    
    return null;
}

/**
 * Extract technical concepts discussed in chat
 */
function extractTechnicalConcepts(content: string): string | null {
    const technicalTerms = [
        // Programming concepts
        'authentication', 'authorization', 'api', 'database', 'frontend', 'backend',
        'microservices', 'docker', 'kubernetes', 'deployment', 'testing', 'debugging',
        'performance', 'optimization', 'security', 'encryption', 'validation',
        'refactoring', 'architecture', 'design pattern', 'algorithm', 'data structure',
        
        // Frameworks and tools
        'react', 'vue', 'angular', 'nodejs', 'express', 'fastapi', 'django', 'flask',
        'spring', 'laravel', 'rails', 'nextjs', 'nuxt', 'gatsby', 'svelte',
        
        // Technologies
        'graphql', 'rest', 'websocket', 'redis', 'mongodb', 'postgresql', 'mysql',
        'elasticsearch', 'kafka', 'rabbitmq', 'aws', 'azure', 'gcp', 'firebase'
    ];
    
    const foundTerms = technicalTerms.filter(term => 
        new RegExp(`\\b${term}\\b`, 'gi').test(content)
    );
    
    if (foundTerms.length > 0) {
        // Use the most mentioned term
        const termCounts = foundTerms.map(term => ({
            term,
            count: (content.match(new RegExp(`\\b${term}\\b`, 'gi')) || []).length
        }));
        
        const topTerm = termCounts.sort((a, b) => b.count - a.count)[0];
        return titleCase(`${topTerm.term} discussion`);
    }
    
    return null;
}

/**
 * Generate session-based name for chat
 */
function generateSessionName(content: string): string {
    const today = new Date().toISOString().split('T')[0];
    
    // Try to extract key action words
    const actionWords = content.match(/\b(help|implement|fix|debug|create|build|review|explain|understand)\b/gi);
    if (actionWords && actionWords.length > 0) {
        const primaryAction = actionWords[0].toLowerCase();
        return titleCase(`${primaryAction} session ${today}`);
    }
    
    return `Development Session ${today}`;
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

