/**
 * WebAssembly service for embedded CLI functionality
 * Provides direct access to Go CLI functions without subprocess overhead
 */

export interface WASMChatData {
    id: string;
    title: string;
    content: string;
    timestamp: number;
    messages: WASMMessage[];
}

export interface WASMMessage {
    role: 'user' | 'assistant';
    content: string;
    timestamp: number;
}

export interface WASMMemory {
    id: string;
    name: string;
    content: string;
    labels: Record<string, string>;
    created: string;
}

/**
 * WebAssembly interface to the Go CLI
 */
export class CMCtlWASMService {
    private wasmModule: WebAssembly.Module | null = null;
    private wasmInstance: WebAssembly.Instance | null = null;
    private go: any = null;
    private isInitialized = false;

    constructor() {
        // Go WASM runtime will be loaded dynamically
    }

    /**
     * Initialize the WebAssembly module
     */
    async initialize(): Promise<void> {
        if (this.isInitialized) {
            return;
        }

        try {
            // Load Go WASM runtime
            const wasmExecPath = require.resolve('./wasm_exec.js');
            const { Go } = await import(wasmExecPath);
            this.go = new Go();

            // Load and instantiate WASM module
            const wasmPath = require.resolve('./cmctl.wasm');
            const wasmBinary = await this.loadWASMBinary(wasmPath);
            this.wasmModule = await WebAssembly.compile(wasmBinary);
            this.wasmInstance = await WebAssembly.instantiate(this.wasmModule, this.go.importObject);

            // Start Go program
            this.go.run(this.wasmInstance);
            this.isInitialized = true;

            console.log('[WASM] ContextMemory CLI initialized successfully');
        } catch (error) {
            console.error('[WASM] Failed to initialize CLI:', error);
            throw new Error(`WebAssembly initialization failed: ${error}`);
        }
    }

    /**
     * Load WASM binary file
     */
    private async loadWASMBinary(path: string): Promise<ArrayBuffer> {
        // In VS Code extension context, we need to read from file system
        const fs = await import('fs');
        const buffer = fs.readFileSync(path);
        return buffer.buffer.slice(buffer.byteOffset, buffer.byteOffset + buffer.byteLength);
    }

    /**
     * Check if WASM service is available
     */
    isAvailable(): boolean {
        return this.isInitialized && 
               typeof (globalThis as any).listCursorChats === 'function';
    }

    /**
     * Import chat from Cursor's AI pane
     */
    async importCursorChat(options: { latest?: boolean; tabId?: string } = {}): Promise<WASMMemory> {
        if (!this.isAvailable()) {
            throw new Error('WASM service not initialized or chat functions not available');
        }

        try {
            const result = (globalThis as any).importCursorChat(JSON.stringify(options));
            const parsedResult = JSON.parse(result);
            
            if (parsedResult.error) {
                throw new Error(parsedResult.error);
            }

            return parsedResult as WASMMemory;
        } catch (error) {
            console.error('[WASM] Failed to import cursor chat:', error);
            throw error;
        }
    }

    /**
     * List available Cursor chats
     */
    async listCursorChats(): Promise<WASMChatData[]> {
        if (!this.isAvailable()) {
            throw new Error('WASM service not initialized or chat functions not available');
        }

        try {
            const result = (globalThis as any).listCursorChats();
            const parsedResult = JSON.parse(result);
            
            if (parsedResult.error) {
                throw new Error(parsedResult.error);
            }

            return parsedResult.chats as WASMChatData[];
        } catch (error) {
            console.error('[WASM] Failed to list cursor chats:', error);
            throw error;
        }
    }

    /**
     * Get health status
     */
    async getHealth(): Promise<{ status: string; version: string }> {
        if (!this.isAvailable()) {
            throw new Error('WASM service not initialized');
        }

        try {
            const result = (globalThis as any).getHealth();
            return JSON.parse(result);
        } catch (error) {
            console.error('[WASM] Failed to get health status:', error);
            throw error;
        }
    }

    /**
     * Fallback: Check if subprocess CLI is available
     */
    hasSubprocessFallback(): boolean {
        // This would check if the external cmctl binary is available
        // Implementation depends on how we want to handle fallbacks
        return false;
    }
}
