import * as vscode from 'vscode';
import * as cp from 'child_process';
import { promisify } from 'util';
import { Memory, CreateMemoryRequest, SearchMemoryRequest, SearchMemoryResponse, StorageInfo, CMCtlConfig } from '../types/memory';

const execAsync = promisify(cp.exec);

/**
 * Service class for interacting with the cmctl CLI
 */
export class CMCtlService implements vscode.Disposable {
    private config: CMCtlConfig;
    private outputChannel: vscode.OutputChannel;

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
            verbosity: config.get<number>('verbosity', 1)
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
        
        // Pass content through stdin
        const fullCommand = `echo "${request.content.replace(/"/g, '\\"')}" | ${command}`;
        
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
     * List all memories
     */
    async listMemories(): Promise<Memory[]> {
        const command = this.buildCommand('list');
        const output = await this.executeCommand(command);
        
        if (!output || output.includes('No memories found')) {
            return [];
        }
        
        const lines = output.split('\n').slice(1); // Skip header
        const memories: Memory[] = [];
        
        for (const line of lines) {
            if (line.trim()) {
                // Parse table format: NAME    LABELS    AGE
                const parts = line.split(/\s{2,}/); // Split on multiple spaces
                if (parts.length >= 2) {
                    const name = parts[0].trim();
                    const labelsStr = parts[1].trim();
                    
                    // For list, we only have basic info, would need to fetch full details
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
                        id: name.toLowerCase().replace(/\s+/g, '_'), // Generate ID from name for now
                        name,
                        content: '', // Would need separate get call for full content
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
     * Show output channel for debugging
     */
    showOutputChannel(): void {
        this.outputChannel.show();
    }

    dispose(): void {
        this.outputChannel.dispose();
    }
}
