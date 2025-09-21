/**
 * Core types for ContextMemory extension
 */

export interface Memory {
    id: string;
    name: string;
    content: string;
    labels: Record<string, string>;
    created: string;
    updated: string;
}

export interface CreateMemoryRequest {
    name: string;
    content: string;
    labels: Record<string, string>;
}

export interface SearchMemoryRequest {
    query?: string;
    labels?: Record<string, string>;
    limit?: number;
}

export interface SearchMemoryResponse {
    memories: Memory[];
    total: number;
}

export interface StorageInfo {
    storageDir: string;
    memoriesCount: number;
    totalSize: string;
}

export interface CMCtlConfig {
    cliPath: string;
    storageDir?: string;
    provider: string;
    verbosity: number;
}

export interface MemoryTreeItem {
    id: string;
    name: string;
    type: 'memory' | 'category' | 'label';
    memory?: Memory;
    children?: MemoryTreeItem[];
}

export enum MemoryCategory {
    Recent = 'recent',
    ByType = 'byType',
    ByLanguage = 'byLanguage',
    ByProject = 'byProject'
}
