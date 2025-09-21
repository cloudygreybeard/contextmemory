import * as vscode from 'vscode';
import { MemoryTreeDataProvider } from './providers/memoryTreeProvider';
import { CMCtlService } from './services/cmctlService';
import { registerCommands } from './commands';

/**
 * Extension activation function
 */
export async function activate(context: vscode.ExtensionContext) {
    console.log('ContextMemory v0.6.0 extension is now active!');

    // Initialize services
    const cmctlService = new CMCtlService();
    
    // Check CLI availability
    try {
        await cmctlService.checkHealth();
        console.log('✅ cmctl CLI is available and healthy');
    } catch (error) {
        vscode.window.showWarningMessage(
            'ContextMemory: cmctl CLI not found or unhealthy. Please ensure cmctl is installed and in PATH.',
            'Install CLI',
            'Check Docs'
        ).then(selection => {
            if (selection === 'Install CLI') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory#installation'));
            } else if (selection === 'Check Docs') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory'));
            }
        });
    }

    // Initialize tree data provider
    const memoryTreeProvider = new MemoryTreeDataProvider(cmctlService);
    
    // Register tree view
    const memoryTreeView = vscode.window.createTreeView('contextmemory.memories', {
        treeDataProvider: memoryTreeProvider,
        showCollapseAll: true,
        canSelectMany: false
    });

    // Register commands
    registerCommands(context, cmctlService, memoryTreeProvider);

    // Add to subscriptions for proper cleanup
    context.subscriptions.push(
        memoryTreeView,
        cmctlService
    );

    // Show welcome message for first-time users
    const hasShownWelcome = context.globalState.get('contextmemory.hasShownWelcome', false);
    if (!hasShownWelcome) {
        vscode.window.showInformationMessage(
            'Welcome to ContextMemory v0.6.0! Transform your coding conversations into searchable memories.',
            'Get Started',
            'View Documentation'
        ).then(selection => {
            if (selection === 'Get Started') {
                vscode.commands.executeCommand('contextmemory.createMemory');
            } else if (selection === 'View Documentation') {
                vscode.env.openExternal(vscode.Uri.parse('https://github.com/cloudygreybeard/contextmemory'));
            }
        });
        context.globalState.update('contextmemory.hasShownWelcome', true);
    }
}

/**
 * Extension deactivation function
 */
export function deactivate() {
    console.log('ContextMemory extension deactivated');
}
