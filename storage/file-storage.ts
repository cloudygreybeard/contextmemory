/**
 * ContextMemory v2 - File-Based Storage Backend
 * 
 * Simple, reliable, local file storage for memories
 * No servers, no databases, just files
 */

import * as fs from 'fs';
import * as path from 'path';
import * as os from 'os';
import {
    Memory,
    CreateMemoryRequest,
    UpdateMemoryRequest,
    SearchMemoryRequest,
    SearchMemoryResponse,
    MemoryStorage,
    MemoryUtils
} from '../core/memory';

export interface FileStorageConfig {
    /** Storage directory (default: ~/.contextmemory-v2) */
    storageDir?: string;
    
    /** Enable auto-backup */
    enableBackup?: boolean;
    
    /** Backup directory */
    backupDir?: string;
}

export class FileBasedStorage implements MemoryStorage {
    private storageDir: string;
    private memoriesDir: string;
    private indexFile: string;
    private configFile: string;
    
    constructor(config: FileStorageConfig = {}) {
        this.storageDir = config.storageDir || path.join(os.homedir(), '.contextmemory-v2');
        this.memoriesDir = path.join(this.storageDir, 'memories');
        this.indexFile = path.join(this.storageDir, 'index.json');
        this.configFile = path.join(this.storageDir, 'config.json');
        
        this.initialize();
    }
    
    private initialize(): void {
        try {
            // Create directories
            [this.storageDir, this.memoriesDir].forEach(dir => {
                if (!fs.existsSync(dir)) {
                    fs.mkdirSync(dir, { recursive: true });
                }
            });
            
            // Initialize index if it doesn't exist
            if (!fs.existsSync(this.indexFile)) {
                this.writeIndex({ memories: [], lastUpdated: new Date().toISOString() });
            }
            
            // Initialize config if it doesn't exist
            if (!fs.existsSync(this.configFile)) {
                const defaultConfig = {
                    version: '2.0.0',
                    created: new Date().toISOString(),
                    storage: 'file-based'
                };
                fs.writeFileSync(this.configFile, JSON.stringify(defaultConfig, null, 2));
            }
            
        } catch (error) {
            throw new Error(`Failed to initialize storage: ${error}`);
        }
    }
    
    async create(request: CreateMemoryRequest): Promise<Memory> {
        try {
            const memory: Memory = {
                id: MemoryUtils.generateId(),
                name: request.name || `Memory ${new Date().toISOString().split('T')[0]}`,
                content: request.content,
                labels: request.labels || { type: 'manual' },
                createdAt: new Date().toISOString(),
                updatedAt: new Date().toISOString(),
                metadata: request.metadata
            };
            
            // Validate
            if (!MemoryUtils.validateName(memory.name)) {
                throw new Error('Invalid memory name');
            }
            if (!MemoryUtils.validateLabels(memory.labels)) {
                throw new Error('Invalid labels');
            }
            
            // Write memory file
            const memoryFile = path.join(this.memoriesDir, `${memory.id}.json`);
            fs.writeFileSync(memoryFile, JSON.stringify(memory, null, 2));
            
            // Update index
            await this.updateIndex(memory, 'create');
            
            console.log(`[FileStorage] Created memory: ${memory.id}`);
            return memory;
            
        } catch (error) {
            throw new Error(`Failed to create memory: ${error}`);
        }
    }
    
    async get(id: string): Promise<Memory | null> {
        try {
            const memoryFile = path.join(this.memoriesDir, `${id}.json`);
            
            if (!fs.existsSync(memoryFile)) {
                return null;
            }
            
            const content = fs.readFileSync(memoryFile, 'utf8');
            return JSON.parse(content) as Memory;
            
        } catch (error) {
            console.error(`[FileStorage] Error reading memory ${id}:`, error);
            return null;
        }
    }
    
    async update(request: UpdateMemoryRequest): Promise<Memory> {
        try {
            const existing = await this.get(request.id);
            if (!existing) {
                throw new Error(`Memory not found: ${request.id}`);
            }
            
            const updated: Memory = {
                ...existing,
                ...(request.name && { name: request.name }),
                ...(request.content && { content: request.content }),
                ...(request.labels && { labels: request.labels }),
                ...(request.metadata && { metadata: { ...existing.metadata, ...request.metadata } }),
                updatedAt: new Date().toISOString()
            };
            
            // Validate
            if (!MemoryUtils.validateName(updated.name)) {
                throw new Error('Invalid memory name');
            }
            if (!MemoryUtils.validateLabels(updated.labels)) {
                throw new Error('Invalid labels');
            }
            
            // Write updated memory
            const memoryFile = path.join(this.memoriesDir, `${updated.id}.json`);
            fs.writeFileSync(memoryFile, JSON.stringify(updated, null, 2));
            
            // Update index
            await this.updateIndex(updated, 'update');
            
            console.log(`[FileStorage] Updated memory: ${updated.id}`);
            return updated;
            
        } catch (error) {
            throw new Error(`Failed to update memory: ${error}`);
        }
    }
    
    async delete(id: string): Promise<void> {
        try {
            const memoryFile = path.join(this.memoriesDir, `${id}.json`);
            
            if (!fs.existsSync(memoryFile)) {
                throw new Error(`Memory not found: ${id}`);
            }
            
            // Delete file
            fs.unlinkSync(memoryFile);
            
            // Update index
            await this.updateIndex({ id } as Memory, 'delete');
            
            console.log(`[FileStorage] Deleted memory: ${id}`);
            
        } catch (error) {
            throw new Error(`Failed to delete memory: ${error}`);
        }
    }
    
    async search(request: SearchMemoryRequest): Promise<SearchMemoryResponse> {
        try {
            const memories = await this.list();
            let filtered = memories;
            
            // Text search
            if (request.query) {
                const query = request.query.toLowerCase();
                filtered = filtered.filter(memory =>
                    memory.name.toLowerCase().includes(query) ||
                    memory.content.toLowerCase().includes(query)
                );
            }
            
            // Label selector
            if (request.labelSelector) {
                filtered = filtered.filter(memory => {
                    for (const [key, value] of Object.entries(request.labelSelector!)) {
                        if (memory.labels[key] !== value) {
                            return false;
                        }
                    }
                    return true;
                });
            }
            
            // Sort
            const sortBy = request.sortBy || 'updatedAt';
            const sortOrder = request.sortOrder || 'desc';
            
            filtered.sort((a, b) => {
                let aVal = a[sortBy];
                let bVal = b[sortBy];
                
                if (sortBy === 'name') {
                    return sortOrder === 'asc' 
                        ? aVal.localeCompare(bVal)
                        : bVal.localeCompare(aVal);
                } else {
                    const aTime = new Date(aVal).getTime();
                    const bTime = new Date(bVal).getTime();
                    return sortOrder === 'asc' ? aTime - bTime : bTime - aTime;
                }
            });
            
            // Limit
            if (request.limit) {
                filtered = filtered.slice(0, request.limit);
            }
            
            return {
                memories: filtered,
                total: memories.length
            };
            
        } catch (error) {
            throw new Error(`Failed to search memories: ${error}`);
        }
    }
    
    async list(): Promise<Memory[]> {
        try {
            const memoryFiles = fs.readdirSync(this.memoriesDir)
                .filter(file => file.endsWith('.json'))
                .map(file => path.join(this.memoriesDir, file));
            
            const memories: Memory[] = [];
            
            for (const file of memoryFiles) {
                try {
                    const content = fs.readFileSync(file, 'utf8');
                    const memory = JSON.parse(content) as Memory;
                    memories.push(memory);
                } catch (error) {
                    console.warn(`[FileStorage] Skipping corrupted file: ${file}`);
                }
            }
            
            return memories;
            
        } catch (error) {
            throw new Error(`Failed to list memories: ${error}`);
        }
    }
    
    async health(): Promise<boolean> {
        try {
            // Check if storage directory is accessible
            fs.accessSync(this.storageDir, fs.constants.R_OK | fs.constants.W_OK);
            
            // Try to write a test file
            const testFile = path.join(this.storageDir, '.health-check');
            fs.writeFileSync(testFile, 'ok');
            fs.unlinkSync(testFile);
            
            return true;
        } catch (error) {
            console.error('[FileStorage] Health check failed:', error);
            return false;
        }
    }
    
    // Storage info
    getStorageInfo(): { 
        storageDir: string;
        memoriesCount: number;
        totalSize: number;
    } {
        try {
            const files = fs.readdirSync(this.memoriesDir).filter(f => f.endsWith('.json'));
            let totalSize = 0;
            
            files.forEach(file => {
                const filePath = path.join(this.memoriesDir, file);
                totalSize += fs.statSync(filePath).size;
            });
            
            return {
                storageDir: this.storageDir,
                memoriesCount: files.length,
                totalSize
            };
        } catch {
            return {
                storageDir: this.storageDir,
                memoriesCount: 0,
                totalSize: 0
            };
        }
    }
    
    private async updateIndex(memory: Memory, operation: 'create' | 'update' | 'delete'): Promise<void> {
        try {
            const index = this.readIndex();
            
            switch (operation) {
                case 'create':
                    index.memories.push({
                        id: memory.id,
                        name: memory.name,
                        labels: memory.labels,
                        createdAt: memory.createdAt,
                        updatedAt: memory.updatedAt
                    });
                    break;
                    
                case 'update':
                    const updateIdx = index.memories.findIndex((m: any) => m.id === memory.id);
                    if (updateIdx >= 0) {
                        index.memories[updateIdx] = {
                            id: memory.id,
                            name: memory.name,
                            labels: memory.labels,
                            createdAt: memory.createdAt,
                            updatedAt: memory.updatedAt
                        };
                    }
                    break;
                    
                case 'delete':
                    index.memories = index.memories.filter((m: any) => m.id !== memory.id);
                    break;
            }
            
            index.lastUpdated = new Date().toISOString();
            this.writeIndex(index);
            
        } catch (error) {
            console.warn('[FileStorage] Failed to update index:', error);
            // Index update failure is not critical
        }
    }
    
    private readIndex(): any {
        try {
            const content = fs.readFileSync(this.indexFile, 'utf8');
            return JSON.parse(content);
        } catch {
            return { memories: [], lastUpdated: new Date().toISOString() };
        }
    }
    
    private writeIndex(index: any): void {
        fs.writeFileSync(this.indexFile, JSON.stringify(index, null, 2));
    }
}
