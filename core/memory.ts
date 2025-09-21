/**
 * ContextMemory v2 - Core Memory Model
 * 
 * Simple, kubernetes-style approach: name + labels + content
 * Minimal metadata, maximum utility
 */

export interface Memory {
    /** Unique identifier */
    id: string;
    
    /** Human-readable name (AI-assisted default) */
    name: string;
    
    /** Content of the memory (session context, chat, notes) */
    content: string;
    
    /** Labels for organization and search (kubernetes-style) */
    labels: Record<string, string>;
    
    /** Core timestamps */
    createdAt: string;
    updatedAt: string;
    
    /** Optional metadata (extensible) */
    metadata?: Record<string, any>;
}

export interface CreateMemoryRequest {
    /** Memory name (will be AI-suggested if not provided) */
    name?: string;
    
    /** Memory content */
    content: string;
    
    /** Labels (will be AI-suggested if not provided) */
    labels?: Record<string, string>;
    
    /** Optional metadata */
    metadata?: Record<string, any>;
}

export interface UpdateMemoryRequest {
    /** Memory ID */
    id: string;
    
    /** Updated name (optional) */
    name?: string;
    
    /** Updated content (optional) */
    content?: string;
    
    /** Updated labels (optional) */
    labels?: Record<string, string>;
    
    /** Updated metadata (optional) */
    metadata?: Record<string, any>;
}

export interface SearchMemoryRequest {
    /** Text search across name and content */
    query?: string;
    
    /** Label selector (kubernetes-style) */
    labelSelector?: Record<string, string>;
    
    /** Limit results */
    limit?: number;
    
    /** Sort order */
    sortBy?: 'createdAt' | 'updatedAt' | 'name';
    sortOrder?: 'asc' | 'desc';
}

export interface SearchMemoryResponse {
    memories: Memory[];
    total: number;
}

/**
 * Core Memory Operations Interface
 * 
 * This interface defines the fundamental CRUD operations
 * that any storage backend must implement
 */
export interface MemoryStorage {
    /** Create a new memory */
    create(request: CreateMemoryRequest): Promise<Memory>;
    
    /** Get memory by ID */
    get(id: string): Promise<Memory | null>;
    
    /** Update existing memory */
    update(request: UpdateMemoryRequest): Promise<Memory>;
    
    /** Delete memory by ID */
    delete(id: string): Promise<void>;
    
    /** Search memories */
    search(request: SearchMemoryRequest): Promise<SearchMemoryResponse>;
    
    /** List all memories (convenience method) */
    list(): Promise<Memory[]>;
    
    /** Health check */
    health(): Promise<boolean>;
}

/**
 * AI Assistant Interface
 * 
 * For generating smart defaults for names and labels
 */
export interface MemoryAssistant {
    /** Suggest a name for memory content */
    suggestName(content: string): Promise<string>;
    
    /** Suggest labels for memory content */
    suggestLabels(content: string, name?: string): Promise<Record<string, string>>;
    
    /** Extract session context from content */
    extractSessionContext(content: string): Promise<{
        name: string;
        labels: Record<string, string>;
        summary: string;
    }>;
}

/**
 * Memory Service
 * 
 * High-level service combining storage and AI assistance
 */
export class MemoryService {
    constructor(
        private storage: MemoryStorage,
        private assistant?: MemoryAssistant
    ) {}
    
    async create(request: CreateMemoryRequest): Promise<Memory> {
        // AI-assist name if not provided
        if (!request.name && this.assistant) {
            request.name = await this.assistant.suggestName(request.content);
        }
        
        // AI-assist labels if not provided
        if (!request.labels && this.assistant) {
            request.labels = await this.assistant.suggestLabels(request.content, request.name);
        }
        
        // Fallback defaults
        if (!request.name) {
            request.name = `Memory ${new Date().toISOString().split('T')[0]}`;
        }
        if (!request.labels) {
            request.labels = { type: 'manual' };
        }
        
        return this.storage.create(request);
    }
    
    async get(id: string): Promise<Memory | null> {
        return this.storage.get(id);
    }
    
    async update(request: UpdateMemoryRequest): Promise<Memory> {
        return this.storage.update(request);
    }
    
    async delete(id: string): Promise<void> {
        return this.storage.delete(id);
    }
    
    async search(request: SearchMemoryRequest): Promise<SearchMemoryResponse> {
        return this.storage.search(request);
    }
    
    async list(): Promise<Memory[]> {
        return this.storage.list();
    }
    
    async health(): Promise<boolean> {
        return this.storage.health();
    }
}

/**
 * Utility functions
 */
export class MemoryUtils {
    /** Generate unique memory ID */
    static generateId(): string {
        const timestamp = Date.now().toString(36);
        const random = Math.random().toString(36).substring(2, 8);
        return `mem_${timestamp}_${random}`;
    }
    
    /** Validate memory name */
    static validateName(name: string): boolean {
        return name.length > 0 && name.length <= 200;
    }
    
    /** Validate labels (kubernetes-style) */
    static validateLabels(labels: Record<string, string>): boolean {
        for (const [key, value] of Object.entries(labels)) {
            if (!key.match(/^[a-z0-9\-_\.\/]+$/i) || key.length > 63) return false;
            if (typeof value !== 'string' || value.length > 63) return false;
        }
        return true;
    }
    
    /** Format memory for display */
    static formatMemory(memory: Memory): string {
        const labels = Object.entries(memory.labels)
            .map(([k, v]) => `${k}=${v}`)
            .join(', ');
        
        return `${memory.name} [${labels}] (${memory.id})`;
    }
}
