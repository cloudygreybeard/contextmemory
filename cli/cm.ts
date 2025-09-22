#!/usr/bin/env node
/**
 * ContextMemory v2 CLI
 * 
 * Simple command-line interface for memory management
 * Usage: cm <command> [options]
 */

import * as fs from 'fs';
import * as path from 'path';
import { MemoryService } from '../core/memory';
import { FileBasedStorage } from '../storage/file-storage';
import { AIAssistantFactory } from '../core/ai-assistant';

// Initialize storage, AI assistant, and service
const storage = new FileBasedStorage();
const aiAssistant = AIAssistantFactory.createDefault();
const memoryService = new MemoryService(storage, aiAssistant);

async function main() {
    const args = process.argv.slice(2);
    const command = args[0];
    
    try {
        switch (command) {
            case 'create':
                await createMemory(args.slice(1));
                break;
            case 'get':
                await getMemory(args.slice(1));
                break;
            case 'update':
                await updateMemory(args.slice(1));
                break;
            case 'delete':
                await deleteMemory(args.slice(1));
                break;
            case 'search':
                await searchMemories(args.slice(1));
                break;
            case 'list':
                await listMemories();
                break;
            case 'health':
                await healthCheck();
                break;
            case 'info':
                await storageInfo();
                break;
            default:
                showHelp();
        }
    } catch (error) {
        console.error('Error:', error instanceof Error ? error.message : error);
        process.exit(1);
    }
}

async function createMemory(args: string[]) {
    const name = getArg(args, '--name', '-n');
    const content = getArg(args, '--content', '-c') || readStdin();
    const labelsStr = getArg(args, '--labels', '-l');
    
    if (!content) {
        throw new Error('Content is required (use --content or pipe from stdin)');
    }
    
    const labels: Record<string, string> = {};
    if (labelsStr) {
        labelsStr.split(',').forEach(pair => {
            const [key, value] = pair.split('=');
            if (key && value) {
                labels[key.trim()] = value.trim();
            }
        });
    }
    
    const memory = await memoryService.create({
        name,
        content,
        labels: Object.keys(labels).length > 0 ? labels : undefined
    });
    
    console.log('Memory created with AI assistance:');
    console.log(`   ID: ${memory.id}`);
    console.log(`   Name: ${memory.name} ${name ? '(manual)' : '(AI-suggested)'}`);
    console.log(`   Labels: ${formatLabels(memory.labels)} ${labelsStr ? '(manual)' : '(AI-suggested)'}`);
    console.log(`   Created: ${memory.createdAt}`);
}

async function getMemory(args: string[]) {
    const id = args[0];
    if (!id) {
        throw new Error('Memory ID is required');
    }
    
    const memory = await memoryService.get(id);
    if (!memory) {
        throw new Error(`Memory not found: ${id}`);
    }
    
    console.log('📄 Memory:');
    console.log(`   ID: ${memory.id}`);
    console.log(`   Name: ${memory.name}`);
    console.log(`   Labels: ${formatLabels(memory.labels)}`);
    console.log(`   Created: ${memory.createdAt}`);
    console.log(`   Updated: ${memory.updatedAt}`);
    console.log('\nContent:');
    console.log(memory.content);
}

async function updateMemory(args: string[]) {
    const id = args[0];
    if (!id) {
        throw new Error('Memory ID is required');
    }
    
    const name = getArg(args, '--name', '-n');
    const content = getArg(args, '--content', '-c');
    const labelsStr = getArg(args, '--labels', '-l');
    
    const updateRequest: any = { id };
    if (name) updateRequest.name = name;
    if (content) updateRequest.content = content;
    if (labelsStr) {
        const labels: Record<string, string> = {};
        labelsStr.split(',').forEach(pair => {
            const [key, value] = pair.split('=');
            if (key && value) {
                labels[key.trim()] = value.trim();
            }
        });
        updateRequest.labels = labels;
    }
    
    const memory = await memoryService.update(updateRequest);
    
    console.log('Memory updated:');
    console.log(`   ID: ${memory.id}`);
    console.log(`   Name: ${memory.name}`);
    console.log(`   Labels: ${formatLabels(memory.labels)}`);
    console.log(`   Updated: ${memory.updatedAt}`);
}

async function deleteMemory(args: string[]) {
    const id = args[0];
    if (!id) {
        throw new Error('Memory ID is required');
    }
    
    await memoryService.delete(id);
    console.log(`Memory deleted: ${id}`);
}

async function searchMemories(args: string[]) {
    const query = getArg(args, '--query', '-q');
    const labelsStr = getArg(args, '--labels', '-l');
    const limit = parseInt(getArg(args, '--limit') || '10');
    
    const labelSelector: Record<string, string> = {};
    if (labelsStr) {
        labelsStr.split(',').forEach(pair => {
            const [key, value] = pair.split('=');
            if (key && value) {
                labelSelector[key.trim()] = value.trim();
            }
        });
    }
    
    const result = await memoryService.search({
        query,
        labelSelector: Object.keys(labelSelector).length > 0 ? labelSelector : undefined,
        limit
    });
    
    console.log(`Found ${result.memories.length} of ${result.total} memories:`);
    result.memories.forEach(memory => {
        console.log(`\n📄 ${memory.name} (${memory.id})`);
        console.log(`   Labels: ${formatLabels(memory.labels)}`);
        console.log(`   Updated: ${memory.updatedAt}`);
        console.log(`   Preview: ${memory.content.substring(0, 100)}${memory.content.length > 100 ? '...' : ''}`);
    });
}

async function listMemories() {
    const memories = await memoryService.list();
    
    console.log(`📚 ${memories.length} memories:`);
    memories.forEach(memory => {
        console.log(`\n📄 ${memory.name} (${memory.id})`);
        console.log(`   Labels: ${formatLabels(memory.labels)}`);
        console.log(`   Updated: ${memory.updatedAt}`);
    });
}

async function healthCheck() {
    const isHealthy = await memoryService.health();
    console.log(`Storage health: ${isHealthy ? 'Healthy' : 'Unhealthy'}`);
}

async function storageInfo() {
    const info = (storage as any).getStorageInfo();
    console.log('Storage Information:');
    console.log(`   Location: ${info.storageDir}`);
    console.log(`   Memories: ${info.memoriesCount}`);
    console.log(`   Size: ${(info.totalSize / 1024).toFixed(1)} KB`);
}

// Utility functions
function getArg(args: string[], longFlag: string, shortFlag?: string): string | undefined {
    const flags = [longFlag];
    if (shortFlag) flags.push(shortFlag);
    
    for (const flag of flags) {
        const index = args.indexOf(flag);
        if (index >= 0 && index + 1 < args.length) {
            return args[index + 1];
        }
    }
    return undefined;
}

function readStdin(): string {
    try {
        return fs.readFileSync(0, 'utf8').trim();
    } catch {
        return '';
    }
}

function formatLabels(labels: Record<string, string>): string {
    return Object.entries(labels)
        .map(([k, v]) => `${k}=${v}`)
        .join(', ') || 'none';
}

function showHelp() {
    console.log(`
ContextMemory v2 CLI

USAGE:
    cm <command> [options]

COMMANDS:
    create              Create a new memory
        --name, -n      Memory name
        --content, -c   Memory content (or pipe from stdin)
        --labels, -l    Labels (format: key1=value1,key2=value2)
    
    get <id>            Get memory by ID
    
    update <id>         Update existing memory
        --name, -n      New name
        --content, -c   New content
        --labels, -l    New labels
    
    delete <id>         Delete memory by ID
    
    search              Search memories
        --query, -q     Text search query
        --labels, -l    Label selector (format: key1=value1,key2=value2)
        --limit         Limit results (default: 10)
    
    list                List all memories
    health              Check storage health
    info                Show storage information

EXAMPLES:
    cm create --name "API Notes" --content "REST endpoints..." --labels "type=notes,project=api"
    cm search --query "authentication" --labels "type=session"
    echo "Session context..." | cm create --name "Debug Session"
    cm get mem_abc123_def456
    cm list
`);
}

// Run CLI
if (require.main === module) {
    main().catch(error => {
        console.error('💥 Fatal error:', error);
        process.exit(1);
    });
}
