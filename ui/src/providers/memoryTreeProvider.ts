import * as vscode from 'vscode';
import { Memory, MemoryTreeItem, MemoryCategory } from '../types/memory';
import { CMCtlService } from '../services/cmctlService';

/**
 * Tree data provider for the memories view
 */
export class MemoryTreeDataProvider implements vscode.TreeDataProvider<MemoryTreeItem> {
    private _onDidChangeTreeData: vscode.EventEmitter<MemoryTreeItem | undefined | null | void> = new vscode.EventEmitter<MemoryTreeItem | undefined | null | void>();
    readonly onDidChangeTreeData: vscode.Event<MemoryTreeItem | undefined | null | void> = this._onDidChangeTreeData.event;

    private memories: Memory[] = [];
    
    constructor(private cmctlService: CMCtlService) {
        this.refresh();
    }

    refresh(): void {
        this.loadMemories();
        this._onDidChangeTreeData.fire();
    }

    private async loadMemories(): Promise<void> {
        try {
            this.memories = await this.cmctlService.listMemories();
        } catch (error: any) {
            vscode.window.showErrorMessage(`Failed to load memories: ${error.message}`);
            this.memories = [];
        }
    }

    getTreeItem(element: MemoryTreeItem): vscode.TreeItem {
        const treeItem = new vscode.TreeItem(
            element.name,
            element.children ? vscode.TreeItemCollapsibleState.Expanded : vscode.TreeItemCollapsibleState.None
        );

        if (element.type === 'memory' && element.memory) {
            treeItem.tooltip = this.createMemoryTooltip(element.memory);
            treeItem.description = this.createMemoryDescription(element.memory);
            treeItem.contextValue = 'memory';
            treeItem.command = {
                command: 'contextmemory.openMemory',
                title: 'Open Memory',
                arguments: [element.memory.id]
            };
            
            // Set icon based on memory type
            const typeLabel = element.memory.labels.type || 'note';
            treeItem.iconPath = this.getIconForType(typeLabel);
        } else if (element.type === 'category') {
            treeItem.iconPath = new vscode.ThemeIcon('folder');
            treeItem.contextValue = 'category';
        } else if (element.type === 'label') {
            treeItem.iconPath = new vscode.ThemeIcon('tag');
            treeItem.contextValue = 'label';
        }

        return treeItem;
    }

    async getChildren(element?: MemoryTreeItem): Promise<MemoryTreeItem[]> {
        if (!element) {
            // Root level - show categories
            await this.loadMemories();
            return this.buildRootCategories();
        }

        if (element.children) {
            return element.children;
        }

        return [];
    }

    private buildRootCategories(): MemoryTreeItem[] {
        const categories: MemoryTreeItem[] = [];

        // Recent memories (last 10)
        const recentMemories = this.memories
            .sort((a, b) => new Date(b.updated || b.created).getTime() - new Date(a.updated || a.created).getTime())
            .slice(0, 10);
        
        if (recentMemories.length > 0) {
            categories.push({
                id: 'recent',
                name: `Recent (${recentMemories.length})`,
                type: 'category',
                children: recentMemories.map(memory => this.createMemoryTreeItem(memory))
            });
        }

        // Group by type
        const byType = this.groupMemoriesByLabel('type');
        if (Object.keys(byType).length > 0) {
            categories.push({
                id: 'byType',
                name: 'By Type',
                type: 'category',
                children: Object.entries(byType).map(([type, memories]) => ({
                    id: `type-${type}`,
                    name: `${type} (${memories.length})`,
                    type: 'label' as const,
                    children: memories.map(memory => this.createMemoryTreeItem(memory))
                }))
            });
        }

        // Group by language
        const byLanguage = this.groupMemoriesByLabel('language');
        if (Object.keys(byLanguage).length > 0) {
            categories.push({
                id: 'byLanguage',
                name: 'By Language',
                type: 'category',
                children: Object.entries(byLanguage).map(([language, memories]) => ({
                    id: `language-${language}`,
                    name: `${language} (${memories.length})`,
                    type: 'label' as const,
                    children: memories.map(memory => this.createMemoryTreeItem(memory))
                }))
            });
        }

        // Group by project
        const byProject = this.groupMemoriesByLabel('project');
        if (Object.keys(byProject).length > 0) {
            categories.push({
                id: 'byProject',
                name: 'By Project',
                type: 'category',
                children: Object.entries(byProject).map(([project, memories]) => ({
                    id: `project-${project}`,
                    name: `${project} (${memories.length})`,
                    type: 'label' as const,
                    children: memories.map(memory => this.createMemoryTreeItem(memory))
                }))
            });
        }

        // All memories
        categories.push({
            id: 'all',
            name: `All Memories (${this.memories.length})`,
            type: 'category',
            children: this.memories.map(memory => this.createMemoryTreeItem(memory))
        });

        return categories;
    }

    private groupMemoriesByLabel(labelKey: string): Record<string, Memory[]> {
        const groups: Record<string, Memory[]> = {};
        
        for (const memory of this.memories) {
            const labelValue = memory.labels[labelKey];
            if (labelValue) {
                if (!groups[labelValue]) {
                    groups[labelValue] = [];
                }
                groups[labelValue].push(memory);
            }
        }
        
        return groups;
    }

    private createMemoryTreeItem(memory: Memory): MemoryTreeItem {
        return {
            id: memory.id,
            name: memory.name,
            type: 'memory',
            memory
        };
    }

    private createMemoryTooltip(memory: Memory): string {
        const labels = Object.entries(memory.labels)
            .map(([key, value]) => `${key}=${value}`)
            .join(', ');
        
        return [
            `**${memory.name}**`,
            '',
            `ID: ${memory.id}`,
            `Created: ${memory.created}`,
            `Updated: ${memory.updated}`,
            `Labels: ${labels || 'none'}`,
            '',
            `Content: ${memory.content.substring(0, 200)}${memory.content.length > 200 ? '...' : ''}`
        ].join('\n');
    }

    private createMemoryDescription(memory: Memory): string {
        // Show the most relevant label as description
        const typeLabel = memory.labels.type;
        const languageLabel = memory.labels.language;
        const projectLabel = memory.labels.project;
        
        if (typeLabel && languageLabel) {
            return `${typeLabel} • ${languageLabel}`;
        } else if (typeLabel) {
            return typeLabel;
        } else if (languageLabel) {
            return languageLabel;
        } else if (projectLabel) {
            return projectLabel;
        }
        
        return '';
    }

    private getIconForType(type: string): vscode.ThemeIcon {
        switch (type.toLowerCase()) {
            case 'code':
                return new vscode.ThemeIcon('code');
            case 'session':
            case 'chat':
                return new vscode.ThemeIcon('comment-discussion');
            case 'note':
            case 'manual':
                return new vscode.ThemeIcon('note');
            case 'snippet':
                return new vscode.ThemeIcon('symbol-snippet');
            case 'meeting':
                return new vscode.ThemeIcon('organization');
            case 'bug':
            case 'issue':
                return new vscode.ThemeIcon('bug');
            case 'feature':
                return new vscode.ThemeIcon('lightbulb');
            default:
                return new vscode.ThemeIcon('file');
        }
    }
}

