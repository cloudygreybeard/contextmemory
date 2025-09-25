import * as vscode from 'vscode';
import * as cp from 'child_process';
import * as fs from 'fs/promises';
import * as path from 'path';
import * as os from 'os';
import { promisify } from 'util';
import { Memory, CreateMemoryRequest, SearchMemoryRequest, SearchMemoryResponse, StorageInfo, CMCtlConfig } from '../types/memory';

const execAsync = promisify(cp.exec);

/**
 * Service class for interacting with the cmctl CLI
 */
export class CMCtlService implements vscode.Disposable {
    private config: CMCtlConfig;
    private outputChannel: vscode.OutputChannel;
    private static readonly EXTENSION_VERSION = '0.7.0';

    constructor() {
        this.outputChannel = vscode.window.createOutputChannel('ContextMemory');
        this.config = this.loadConfig();
        
        // Listen for configuration changes
        vscode.workspace.onDidChangeConfiguration(event => {
            if (event.affectsConfiguration('contextmemory')) {
                this.config = this.loadConfig();
                this.outputChannel.appendLine('Configuration updated');
            }
        });
    }

    private loadConfig(): CMCtlConfig {
        const config = vscode.workspace.getConfiguration('contextmemory');
        return {
            cliPath: config.get<string>('cliPath', 'cmctl'),
            storageDir: config.get<string>('storageDir', ''),
            provider: config.get<string>('provider', 'file'),
            verbosity: config.get<number>('verbosity', 1),
            showMemoryIds: config.get<boolean>('showMemoryIds', false)
        };
    }

    private buildCommand(subcommand: string, args: string[] = []): string {
        const baseArgs: string[] = [];
        
        // Add global flags
        if (this.config.storageDir) {
            baseArgs.push('--storage-dir', this.config.storageDir);
        }
        
        if (this.config.provider !== 'file') {
            baseArgs.push('--provider', this.config.provider);
        }
        
        if (this.config.verbosity !== 1) {
            baseArgs.push('-v', this.config.verbosity.toString());
        }

        return `${this.config.cliPath} ${baseArgs.join(' ')} ${subcommand} ${args.join(' ')}`.trim();
    }

    private async executeCommand(command: string): Promise<string> {
        this.outputChannel.appendLine(`Executing: ${command}`);
        
        try {
            const { stdout, stderr } = await execAsync(command);
            
            if (stderr) {
                this.outputChannel.appendLine(`Warning: ${stderr}`);
            }
            
            this.outputChannel.appendLine(`Output: ${stdout}`);
            return stdout.trim();
        } catch (error: any) {
            const errorMessage = error.message || error.toString();
            this.outputChannel.appendLine(`Error: ${errorMessage}`);
            throw new Error(`cmctl command failed: ${errorMessage}`);
        }
    }

    /**
     * Check if cmctl CLI is available and healthy
     */
    async checkHealth(): Promise<void> {
        const command = this.buildCommand('health');
        await this.executeCommand(command);
    }

    /**
     * Get CLI version
     */
    async getVersion(): Promise<string> {
        const command = `${this.config.cliPath} --version`;
        const output = await this.executeCommand(command);
        return output.replace('cmctl version ', '');
    }

    /**
     * Parse semantic version string into components
     */
    private parseVersion(version: string): { major: number; minor: number; patch: number } {
        const cleanVersion = version.replace(/^v/, '').trim();
        const parts = cleanVersion.split('.');
        
        if (parts.length !== 3) {
            throw new Error(`Invalid version format: ${version}`);
        }

        return {
            major: parseInt(parts[0], 10),
            minor: parseInt(parts[1], 10),
            patch: parseInt(parts[2], 10)
        };
    }

    /**
     * Check if CLI version is compatible with extension version
     * Policy: Minor version compatibility (0.6.x extension works with 0.6.y CLI)
     */
    async checkVersionCompatibility(): Promise<{ compatible: boolean; reason?: string; cliVersion: string; extensionVersion: string }> {
        try {
            const cliVersion = await this.getVersion();
            const extensionVersion = CMCtlService.EXTENSION_VERSION;

            const cli = this.parseVersion(cliVersion);
            const ext = this.parseVersion(extensionVersion);

            const compatible = cli.major === ext.major && cli.minor === ext.minor;

            return {
                compatible,
                cliVersion,
                extensionVersion,
                reason: compatible ? undefined : 
                    `Extension v${extensionVersion} requires CLI v${ext.major}.${ext.minor}.x, but found v${cliVersion}`
            };
        } catch (error: any) {
            return {
                compatible: false,
                cliVersion: 'unknown',
                extensionVersion: CMCtlService.EXTENSION_VERSION,
                reason: `Failed to check CLI version: ${error.message}`
            };
        }
    }

    /**
     * Perform comprehensive health and compatibility check
     */
    async checkHealthAndCompatibility(): Promise<void> {
        // First check basic health
        await this.checkHealth();
        
        // Then check version compatibility
        const versionCheck = await this.checkVersionCompatibility();
        
        if (!versionCheck.compatible) {
            throw new Error(versionCheck.reason || 'Version compatibility check failed');
        }
    }

    /**
     * Get storage information
     */
    async getStorageInfo(): Promise<StorageInfo> {
        const command = this.buildCommand('info');
        const output = await this.executeCommand(command);
        
        // Parse the output (simplified parser)
        const lines = output.split('\n');
        const storageDir = lines.find(line => line.includes('Storage Directory:'))?.split(':')[1]?.trim() || '';
        const memoriesCount = parseInt(lines.find(line => line.includes('Total Memories:'))?.split(':')[1]?.trim() || '0');
        const totalSize = lines.find(line => line.includes('Storage Size:'))?.split(':')[1]?.trim() || '0B';
        
        return {
            storageDir,
            memoriesCount,
            totalSize
        };
    }

    /**
     * Create a new memory
     */
    async createMemory(request: CreateMemoryRequest): Promise<Memory> {
        const args: string[] = [];
        
        args.push('--name', `"${request.name}"`);
        
        // Add labels
        if (Object.keys(request.labels).length > 0) {
            const labelString = Object.entries(request.labels)
                .map(([key, value]) => `${key}=${value}`)
                .join(',');
            args.push('--labels', `"${labelString}"`);
        }

        const command = this.buildCommand('create', args);
        
        // Pass content through stdin, properly escape for shell
        const escapedContent = request.content
            .replace(/\\/g, '\\\\')  // Escape backslashes first
            .replace(/"/g, '\\"')    // Escape quotes
            .replace(/\$/g, '\\$')   // Escape dollar signs
            .replace(/`/g, '\\`');   // Escape backticks
        
        const fullCommand = `echo "${escapedContent}" | ${command}`;
        
        const output = await this.executeCommand(fullCommand);
        
        // Extract memory ID from output (format: memory/mem_id created)
        const match = output.match(/memory\/(\S+)\s+created/);
        if (!match) {
            throw new Error('Failed to extract memory ID from create output');
        }
        
        const memoryId = match[1];
        
        // Fetch the created memory
        return await this.getMemory(memoryId);
    }

    /**
     * Get a specific memory by ID
     */
    async getMemory(id: string): Promise<Memory> {
        const command = this.buildCommand('get', [id]);
        const output = await this.executeCommand(command);
        
        // Parse the output format
        const lines = output.split('\n');
        const name = lines.find(line => line.startsWith('Name:'))?.split('\t')[1] || '';
        const content = lines.slice(lines.findIndex(line => line.includes('Content:')) + 2).join('\n');
        const labelsLine = lines.find(line => line.startsWith('Labels:'))?.split('\t')[1] || '';
        const created = lines.find(line => line.startsWith('Created:'))?.split('\t')[1] || '';
        const updated = lines.find(line => line.startsWith('Updated:'))?.split('\t')[1] || '';
        
        // Parse labels
        const labels: Record<string, string> = {};
        if (labelsLine && labelsLine !== 'none') {
            labelsLine.split(',').forEach(labelPair => {
                const [key, value] = labelPair.split('=');
                if (key && value) {
                    labels[key.trim()] = value.trim();
                }
            });
        }
        
        return {
            id,
            name,
            content,
            labels,
            created,
            updated
        };
    }

    /**
     * List all memories using the unified get command
     */
    async listMemories(): Promise<Memory[]> {
        const args: string[] = ['-o', 'json'];
        
        const command = this.buildCommand('get', args);
        const output = await this.executeCommand(command);
        
        if (!output || output.includes('No resources found')) {
            return [];
        }
        
        try {
            const parsed = JSON.parse(output);
            const memories: Memory[] = [];
            
            if (parsed.items) {
                for (const item of parsed.items) {
                    memories.push({
                        id: item.id,
                        name: item.name,
                        content: item.content || '',
                        labels: item.labels || {},
                        created: item.createdAt || '',
                        updated: item.updatedAt || ''
                    });
                }
            }
            
            return memories;
        } catch (error) {
            console.error('Failed to parse JSON output from cmctl get:', error);
            return [];
        }
    }

    /**
     * Search memories
     */
    async searchMemories(request: SearchMemoryRequest): Promise<SearchMemoryResponse> {
        const args: string[] = [];
        
        if (request.query) {
            args.push('--query', `"${request.query}"`);
        }
        
        if (request.labels && Object.keys(request.labels).length > 0) {
            const labelString = Object.entries(request.labels)
                .map(([key, value]) => `${key}=${value}`)
                .join(',');
            args.push('--labels', `"${labelString}"`);
        }

        const command = this.buildCommand('search', args);
        const output = await this.executeCommand(command);
        
        // Similar parsing to listMemories but for search results
        const memories = await this.parseMemoryList(output);
        
        return {
            memories,
            total: memories.length
        };
    }

    private async parseMemoryList(output: string): Promise<Memory[]> {
        if (!output || output.includes('No memories found')) {
            return [];
        }
        
        const lines = output.split('\n').slice(1); // Skip header
        const memories: Memory[] = [];
        
        for (const line of lines) {
            if (line.trim()) {
                const parts = line.split(/\s{2,}/);
                if (parts.length >= 2) {
                    const name = parts[0].trim();
                    const labelsStr = parts[1].trim();
                    
                    const labels: Record<string, string> = {};
                    if (labelsStr && labelsStr !== '-') {
                        labelsStr.split(',').forEach(labelPair => {
                            const [key, value] = labelPair.split('=');
                            if (key && value) {
                                labels[key.trim()] = value.trim();
                            }
                        });
                    }
                    
                    memories.push({
                        id: name.toLowerCase().replace(/\s+/g, '_'),
                        name,
                        content: '',
                        labels,
                        created: '',
                        updated: ''
                    });
                }
            }
        }
        
        return memories;
    }

    /**
     * Update a memory
     */
    async updateMemory(id: string, updates: Partial<CreateMemoryRequest>): Promise<Memory> {
        try {
            // Get current memory
            const currentMemory = await this.getMemory(id);
            
            // For now, we'll implement update by recreating the memory
            // This is a limitation of the current CLI - we should add proper update support
            const updateRequest: CreateMemoryRequest = {
                name: updates.name || currentMemory.name,
                content: updates.content || currentMemory.content,
                labels: { ...currentMemory.labels, ...(updates.labels || {}) }
            };

            // Delete the old memory
            await this.deleteMemory(id, true);
            
            // Create the new memory with updated content
            const updatedMemory = await this.createMemory(updateRequest);
            
            this.outputChannel.appendLine(`Memory ${id} updated successfully`);
            return updatedMemory;
        } catch (error: any) {
            this.outputChannel.appendLine(`Failed to update memory: ${error.message}`);
            throw new Error(`Update failed: ${error.message}`);
        }
    }

    /**
     * Delete a memory by ID
     */
    async deleteMemory(memoryId: string, force: boolean = false): Promise<void> {
        const args = ['delete', memoryId];
        if (force) {
            args.push('--force');
        }
        const command = this.buildCommand('', args);
        await this.executeCommand(command);
    }

    /**
     * Delete memories by label selector
     */
    async deleteMemoriesByLabels(labelSelector: string, force: boolean = false): Promise<void> {
        const args = ['delete', '--labels', labelSelector];
        if (force) {
            args.push('--force');
        }
        const command = this.buildCommand('', args);
        await this.executeCommand(command);
    }

    /**
     * Delete all memories
     */
    async deleteAllMemories(force: boolean = false): Promise<void> {
        const args = ['delete', '--all'];
        if (force) {
            args.push('--force');
        }
        const command = this.buildCommand('', args);
        await this.executeCommand(command);
    }

    /**
     * Get current configuration
     */
    async getConfig(): Promise<any> {
        // For now, return the VS Code configuration
        // Later we can extend this to read the actual cmctl config file
        const config = vscode.workspace.getConfiguration('contextmemory');
        return {
            cliPath: config.get('cliPath'),
            storageDir: config.get('storageDir'),
            provider: config.get('provider'),
            verbosity: config.get('verbosity'),
            autoSuggestLabels: config.get('autoSuggestLabels'),
            defaultLabels: config.get('defaultLabels'),
            showMemoryIds: config.get('showMemoryIds')
        };
    }

    /**
     * Update configuration
     */
    async updateConfig(configUpdates: any): Promise<void> {
        const config = vscode.workspace.getConfiguration('contextmemory');
        for (const [key, value] of Object.entries(configUpdates)) {
            await config.update(key, value, vscode.ConfigurationTarget.Global);
        }
        this.config = this.loadConfig(); // Reload config
    }

    /**
     * Import chat from Cursor AI pane - Enhanced with export integration
     */
    async importCursorChat(): Promise<{ success: boolean; memoryName?: string; error?: string }> {
        try {
            // First, try to use Cursor's native export functionality
            const exportResult = await this.tryNativeCursorExport();
            if (exportResult.success) {
                return exportResult;
            }
            
            // Fallback to database parsing approach
            this.outputChannel.appendLine('Native export failed, falling back to database parsing...');
            const command = this.buildCommand('import-cursor-chat', ['--latest']);
            const output = await this.executeCommand(command);
            
            // Parse the output to extract memory name and ID
            // Expected format: "Successfully imported chat as memory:\nID: mem_xxx\nName: Chat Name"
            const nameMatch = output.match(/Name:\s*(.+)/);
            const idMatch = output.match(/ID:\s*(mem_[a-f0-9_]+)/);
            
            if (output.includes('Successfully imported chat as memory')) {
                const memoryName = nameMatch ? nameMatch[1].trim() : 'Cursor Chat';
                
                return {
                    success: true,
                    memoryName: memoryName
                };
            } else {
                throw new Error('Unexpected output format from import command');
            }
        } catch (error: any) {
            return {
                success: false,
                error: error.message || 'Failed to import cursor chat'
            };
        }
    }

    /**
     * Try to use Cursor's native export functionality
     */
    private async tryNativeCursorExport(): Promise<{ success: boolean; memoryName?: string; error?: string }> {
        try {
            // Discover available chat/export commands
            const commands = await vscode.commands.getCommands(true);
            const exportCommands = commands.filter(cmd => 
                (cmd.toLowerCase().includes('export') && cmd.toLowerCase().includes('chat')) ||
                cmd.toLowerCase().includes('cursor.export') ||
                cmd.toLowerCase().includes('aichat.export')
            );
            
            this.outputChannel.appendLine(`Found potential export commands: ${exportCommands.join(', ')}`);
            
            // Try known command patterns
            const possibleCommands = [
                'cursor.exportChat',
                'cursor.aichat.export',
                'workbench.panel.aichat.export',
                'aichat.export',
                'ai.exportChat',
                ...exportCommands
            ];
            
            for (const cmdName of possibleCommands) {
                try {
                    this.outputChannel.appendLine(`Trying command: ${cmdName}`);
                    await vscode.commands.executeCommand(cmdName);
                    
                    // If command executed successfully, wait for export file
                    const exportedContent = await this.waitForExportFile();
                    if (exportedContent) {
                        return await this.importFromExportedContent(exportedContent);
                    }
                } catch (cmdError: any) {
                    this.outputChannel.appendLine(`Command ${cmdName} failed: ${cmdError.message}`);
                    continue;
                }
            }
            
            throw new Error('No working export command found');
        } catch (error: any) {
            this.outputChannel.appendLine(`Native export failed: ${error.message}`);
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Wait for and detect exported chat file
     */
    private async waitForExportFile(): Promise<string | null> {
        try {
            // Strategy 1: Monitor Downloads folder for new .md files
            const downloadsPath = path.join(os.homedir(), 'Downloads');
            this.outputChannel.appendLine(`Monitoring Downloads folder: ${downloadsPath}`);
            
            // Check if we have permission to read Downloads
            try {
                await fs.access(downloadsPath, fs.constants.R_OK);
            } catch (permError: any) {
                this.outputChannel.appendLine(`Downloads folder access denied (${permError.code}), skipping auto-detection`);
                throw new Error('Downloads access denied');
            }
            
            // Get initial file list
            const initialFiles = await this.getMarkdownFiles(downloadsPath);
            
            // Wait up to 10 seconds for a new file
            for (let i = 0; i < 20; i++) {
                await new Promise(resolve => setTimeout(resolve, 500));
                
                const currentFiles = await this.getMarkdownFiles(downloadsPath);
                const newFiles = currentFiles.filter(file => !initialFiles.includes(file));
                
                if (newFiles.length > 0) {
                    // Found a new markdown file - likely our export
                    const newestFile = newFiles[0]; // Take the first new file
                    this.outputChannel.appendLine(`Found new export file: ${newestFile}`);
                    
                    // Read and return the content
                    return await fs.readFile(newestFile, 'utf8');
                }
            }
            
            // Strategy 2: Prompt user to select the exported file
            this.outputChannel.appendLine('No new file detected, prompting user to select exported file...');
            
            const selectedFile = await vscode.window.showOpenDialog({
                canSelectFiles: true,
                canSelectFolders: false,
                canSelectMany: false,
                filters: {
                    'Markdown files': ['md']
                },
                title: 'Select the exported chat file'
            });
            
            if (selectedFile && selectedFile[0]) {
                const filePath = selectedFile[0].fsPath;
                this.outputChannel.appendLine(`User selected file: ${filePath}`);
                return await fs.readFile(filePath, 'utf8');
            }
            
            this.outputChannel.appendLine('No file selected, falling back to database parsing');
            return null;
            
        } catch (error: any) {
            this.outputChannel.appendLine(`Error in waitForExportFile: ${error.message}`);
            return null;
        }
    }

    /**
     * Get list of markdown files in directory, sorted by modification time (newest first)
     */
    private async getMarkdownFiles(directory: string): Promise<string[]> {
        try {
            const files = await fs.readdir(directory);
            const mdFiles = files.filter(file => file.endsWith('.md'));
            const fullPaths = mdFiles.map(file => path.join(directory, file));
            
            // Sort by modification time (newest first)
            const fileStats = await Promise.all(
                fullPaths.map(async (filePath) => ({
                    path: filePath,
                    mtime: (await fs.stat(filePath)).mtime
                }))
            );
            
            return fileStats
                .sort((a, b) => b.mtime.getTime() - a.mtime.getTime())
                .map(f => f.path);
        } catch (error: any) {
            this.outputChannel.appendLine(`Error reading directory ${directory}: ${error.message}`);
            return [];
        }
    }

    /**
     * Import memory from exported chat content
     */
    private async importFromExportedContent(content: string): Promise<{ success: boolean; memoryName?: string; error?: string }> {
        try {
            // Create memory from the exported content
            // This would involve parsing the exported markdown and creating a memory
            const lines = content.split('\n');
            const title = lines.find(line => line.startsWith('# '))?.substring(2) || 'Exported Chat';
            
            // Use CLI to create memory from content
            const createResult = await this.createMemoryFromText(title, content);
            
            return {
                success: true,
                memoryName: title
            };
        } catch (error: any) {
            return {
                success: false,
                error: error.message
            };
        }
    }

    /**
     * Create memory from text content
     */
    private async createMemoryFromText(name: string, content: string): Promise<void> {
        // Use --content flag directly instead of non-existent --file flag
        const command = this.buildCommand('create', ['--name', name, '--content', content, '--labels', 'type=chat,source=cursor-ai-pane']);
        await this.executeCommand(command);
    }

    /**
     * Reload a chat memory with specified format
     */
    async reloadChat(memoryId: string, format: string = 'conversational'): Promise<string> {
        const args: string[] = ['reload-chat', memoryId, '--format', format];
        
        const command = this.buildCommand('cmctl', args);
        return await this.executeCommand(command);
    }

    /**
     * Show output channel for debugging
     */
    showOutputChannel(): void {
        this.outputChannel.show();
    }

    dispose(): void {
        this.outputChannel.dispose();
    }
}

