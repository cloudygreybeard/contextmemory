/**
 * ContextMemory v2 - AI Assistant Implementation
 * 
 * Smart defaults for memory names and labels using LLM integration
 * Designed to enhance UX without getting in the user's way
 */

import { MemoryAssistant } from './memory';

/**
 * Configuration for AI Assistant
 */
export interface AIAssistantConfig {
    /** Provider type */
    provider: 'openai' | 'anthropic' | 'local' | 'mock';
    
    /** API configuration */
    apiKey?: string;
    baseUrl?: string;
    model?: string;
    
    /** Behavior settings */
    maxTokens?: number;
    temperature?: number;
    
    /** Fallback settings */
    enableFallbacks?: boolean;
    fallbackProvider?: 'rule-based' | 'template';
}

/**
 * Base AI Assistant Interface
 */
export abstract class BaseAIAssistant implements MemoryAssistant {
    protected config: AIAssistantConfig;
    
    constructor(config: AIAssistantConfig) {
        this.config = {
            maxTokens: 100,
            temperature: 0.3,
            enableFallbacks: true,
            fallbackProvider: 'rule-based',
            ...config
        };
    }
    
    abstract suggestName(content: string): Promise<string>;
    abstract suggestLabels(content: string, name?: string): Promise<Record<string, string>>;
    abstract extractSessionContext(content: string): Promise<{
        name: string;
        labels: Record<string, string>;
        summary: string;
    }>;
    
    protected async fallbackName(content: string): Promise<string> {
        // Rule-based fallback for name generation
        const lines = content.split('\n').filter(line => line.trim());
        const firstLine = lines[0]?.trim() || '';
        
        // Extract from first meaningful line
        if (firstLine.length > 0) {
            const cleaned = firstLine
                .replace(/^(#|\/\/|\*|>|-)?\s*/, '') // Remove common prefixes
                .substring(0, 50) // Limit length
                .trim();
            
            if (cleaned.length > 5) {
                return cleaned;
            }
        }
        
        // Fallback to date-based name
        const date = new Date().toISOString().split('T')[0];
        const time = new Date().toTimeString().split(' ')[0].substring(0, 5);
        return `Memory ${date} ${time}`;
    }
    
    protected async fallbackLabels(content: string, name?: string): Promise<Record<string, string>> {
        const labels: Record<string, string> = {};
        
        // Detect content type
        if (this.isSessionContext(content)) {
            labels.type = 'session';
        } else if (this.isCodeSnippet(content)) {
            labels.type = 'code';
        } else if (this.isDocumentation(content)) {
            labels.type = 'docs';
        } else {
            labels.type = 'manual';
        }
        
        // Detect programming language
        const language = this.detectLanguage(content);
        if (language) {
            labels.language = language;
        }
        
        // Add source if name suggests it
        if (name) {
            if (name.toLowerCase().includes('debug')) {
                labels.category = 'debug';
            } else if (name.toLowerCase().includes('api')) {
                labels.category = 'api';
            } else if (name.toLowerCase().includes('error')) {
                labels.category = 'error';
            }
        }
        
        // Add creation context
        labels.created = new Date().toISOString().split('T')[0];
        
        return labels;
    }
    
    private isSessionContext(content: string): boolean {
        const sessionKeywords = ['session', 'context', 'conversation', 'chat', 'discussion'];
        return sessionKeywords.some(keyword => 
            content.toLowerCase().includes(keyword.toLowerCase())
        );
    }
    
    private isCodeSnippet(content: string): boolean {
        const codePatterns = [
            /function\s+\w+\s*\(/,
            /class\s+\w+\s*{/,
            /import\s+.*\s+from/,
            /def\s+\w+\s*\(/,
            /public\s+class\s+\w+/,
            /{[\s\S]*}/,
            /```[\s\S]*```/
        ];
        
        return codePatterns.some(pattern => pattern.test(content));
    }
    
    private isDocumentation(content: string): boolean {
        const docPatterns = [
            /^#+\s+/m, // Markdown headers
            /README/i,
            /documentation/i,
            /\*\*\w+\*\*/,  // Bold text
            /^\s*-\s+/m     // List items
        ];
        
        return docPatterns.some(pattern => pattern.test(content));
    }
    
    protected detectLanguage(content: string): string | null {
        const languagePatterns: Record<string, RegExp[]> = {
            typescript: [/interface\s+\w+/, /type\s+\w+\s*=/, /import.*from/, /\.ts$/],
            javascript: [/function\s+\w+/, /const\s+\w+\s*=/, /require\(/, /\.js$/],
            python: [/def\s+\w+/, /import\s+\w+/, /from\s+\w+\s+import/, /\.py$/],
            java: [/public\s+class/, /public\s+static\s+void\s+main/, /\.java$/],
            go: [/func\s+\w+/, /package\s+main/, /import\s+\(/, /\.go$/],
            rust: [/fn\s+\w+/, /use\s+std::/, /cargo/i, /\.rs$/],
            shell: [/^#!\/bin/, /echo\s+/, /\$\{?\w+\}?/, /\.sh$/],
            sql: [/SELECT\s+.*\s+FROM/i, /CREATE\s+TABLE/i, /INSERT\s+INTO/i]
        };
        
        for (const [language, patterns] of Object.entries(languagePatterns)) {
            if (patterns.some(pattern => pattern.test(content))) {
                return language;
            }
        }
        
        return null;
    }
}

/**
 * Mock AI Assistant (for development and testing)
 */
export class MockAIAssistant extends BaseAIAssistant {
    async suggestName(content: string): Promise<string> {
        // Simulate AI processing delay
        await this.delay(100);
        
        try {
            return await this.extractMeaningfulName(content);
        } catch (error) {
            console.warn('[MockAI] Name suggestion failed, using fallback:', error);
            return this.fallbackName(content);
        }
    }
    
    private async extractMeaningfulName(content: string): Promise<string> {
        const lines = content.split('\n').filter(line => line.trim());
        
        // Strategy 1: Look for function/class/interface definitions
        const codeMatch = this.extractCodeEntityName(lines);
        if (codeMatch) return codeMatch;
        
        // Strategy 2: Look for headers/titles 
        const headerMatch = this.extractHeaderName(lines);
        if (headerMatch) return headerMatch;
        
        // Strategy 3: Extract key concepts from meaningful sentences
        const conceptMatch = this.extractConceptualName(lines);
        if (conceptMatch) return conceptMatch;
        
        // Fallback to improved first-line extraction
        return await this.extractFromFirstLine(lines);
    }
    
    private extractCodeEntityName(lines: string[]): string | null {
        const codePatterns = [
            /(?:function|const|let|var)\s+(\w+)/,
            /class\s+(\w+)/,
            /interface\s+(\w+)/,
            /type\s+(\w+)/,
            /def\s+(\w+)/,  // Python
            /fn\s+(\w+)/,   // Rust
            /func\s+(\w+)/  // Go
        ];
        
        for (const line of lines) {
            for (const pattern of codePatterns) {
                const match = line.match(pattern);
                if (match && match[1]) {
                    return this.humanizeCodeName(match[1]);
                }
            }
        }
        return null;
    }
    
    private extractHeaderName(lines: string[]): string | null {
        for (const line of lines) {
            const trimmed = line.trim();
            
            // Markdown headers
            const headerMatch = trimmed.match(/^#+\s+(.+)$/);
            if (headerMatch) {
                return this.cleanTitle(headerMatch[1]);
            }
            
            // Title-like patterns (all caps, or Title Case)
            if (trimmed.length < 60 && (
                trimmed === trimmed.toUpperCase() || 
                /^[A-Z][a-z]+(\s+[A-Z][a-z]+)*$/.test(trimmed)
            )) {
                return this.cleanTitle(trimmed);
            }
        }
        return null;
    }
    
    private extractConceptualName(lines: string[]): string | null {
        // Look for lines that might contain key concepts
        const meaningfulLines = lines.filter(line => {
            const trimmed = line.trim();
            return trimmed.length > 20 && 
                   trimmed.length < 150 &&
                   !trimmed.startsWith('//') &&
                   !trimmed.startsWith('#') &&
                   !trimmed.startsWith('*') &&
                   !/^[\{\}\[\]\(\)]/.test(trimmed);
        });
        
        for (const line of meaningfulLines) {
            // Extract key nouns and concepts
            const concepts = this.extractKeyConcepts(line);
            if (concepts.length >= 2) {
                return concepts.slice(0, 4).join(' ');
            }
        }
        
        return null;
    }
    
    private extractKeyConcepts(text: string): string[] {
        // Remove common conversational starters
        const cleaned = text
            .replace(/^(today|yesterday|currently|now|i'm|i am|we're|we are|this is|this)\s+/i, '')
            .replace(/^(working on|implementing|creating|building|developing|fixing)\s+/i, '');
        
        // Extract meaningful words (nouns, technical terms)
        const words = cleaned
            .replace(/[^\w\s'-]/g, ' ')  // Keep apostrophes and hyphens
            .replace(/\s+/g, ' ')
            .trim()
            .split(' ')
            .filter(word => word.length > 2)
            .filter(word => !this.isStopWord(word.toLowerCase()));
        
        // Prioritize capitalized words and technical terms
        const prioritized = words
            .map(word => ({
                word: this.titleCase(word),
                score: this.getWordImportance(word)
            }))
            .sort((a, b) => b.score - a.score)
            .map(item => item.word);
        
        return prioritized;
    }
    
    private async extractFromFirstLine(lines: string[]): Promise<string> {
        if (lines.length === 0) {
            return await this.fallbackName('');
        }
        
        const firstLine = lines[0].trim();
        const concepts = this.extractKeyConcepts(firstLine);
        
        if (concepts.length >= 2) {
            return concepts.slice(0, 4).join(' ');
        }
        
        // Last resort: clean up the first line better
        const cleaned = firstLine
            .replace(/^(today|yesterday|currently|now|i'm|i am|we're|we are|this is|this)\s+/i, '')
            .replace(/[^\w\s'-]/g, ' ')
            .replace(/\s+/g, ' ')
            .trim();
        
        if (cleaned.length > 0) {
            const words = cleaned.split(' ').slice(0, 4);
            return words.map(word => this.titleCase(word)).join(' ');
        }
        
        return await this.fallbackName(firstLine);
    }
    
    private humanizeCodeName(name: string): string {
        // Convert camelCase or snake_case to readable names
        return name
            .replace(/([A-Z])/g, ' $1')
            .replace(/_/g, ' ')
            .trim()
            .split(' ')
            .map(word => this.titleCase(word))
            .join(' ');
    }
    
    private cleanTitle(title: string): string {
        return title
            .replace(/[^\w\s'-]/g, ' ')
            .replace(/\s+/g, ' ')
            .trim()
            .split(' ')
            .map(word => this.titleCase(word))
            .join(' ');
    }
    
    private isStopWord(word: string): boolean {
        const stopWords = new Set([
            'the', 'a', 'an', 'and', 'or', 'but', 'in', 'on', 'at', 'to', 'for', 'of', 'with',
            'by', 'from', 'about', 'into', 'through', 'during', 'before', 'after', 'above',
            'below', 'up', 'down', 'out', 'off', 'over', 'under', 'again', 'further', 'then',
            'once', 'here', 'there', 'when', 'where', 'why', 'how', 'all', 'any', 'both',
            'each', 'few', 'more', 'most', 'other', 'some', 'such', 'no', 'nor', 'not',
            'only', 'own', 'same', 'so', 'than', 'too', 'very', 'can', 'will', 'just',
            'should', 'now', 'am', 'is', 'are', 'was', 'were', 'be', 'been', 'being',
            'have', 'has', 'had', 'do', 'does', 'did', 'doing', 'would', 'could', 'should'
        ]);
        return stopWords.has(word);
    }
    
    private getWordImportance(word: string): number {
        let score = 1;
        
        // Technical terms get higher scores
        if (/^[A-Z][a-z]*[A-Z]/.test(word)) score += 2; // camelCase
        if (word.includes('_') || word.includes('-')) score += 1; // snake_case or kebab-case
        if (/^[A-Z]+$/.test(word) && word.length > 1) score += 2; // Acronyms
        if (word.length > 6) score += 1; // Longer words often more specific
        
        // Context-specific terms
        const technicalTerms = ['api', 'function', 'class', 'interface', 'service', 'component',
                              'memory', 'context', 'session', 'implementation', 'assistant'];
        if (technicalTerms.some(term => word.toLowerCase().includes(term))) score += 2;
        
        return score;
    }
    
    private titleCase(word: string): string {
        if (word.length === 0) return word;
        return word.charAt(0).toUpperCase() + word.slice(1).toLowerCase();
    }
    
    async suggestLabels(content: string, name?: string): Promise<Record<string, string>> {
        // Simulate AI processing delay
        await this.delay(150);
        
        try {
            const baseLabels = await this.fallbackLabels(content, name);
            
            // Add AI-enhanced labels
            if (content.length > 1000) {
                baseLabels.size = 'large';
            } else if (content.length > 300) {
                baseLabels.size = 'medium';
            } else {
                baseLabels.size = 'small';
            }
            
            // Detect complexity
            const complexity = this.assessComplexity(content);
            if (complexity !== 'simple') {
                baseLabels.complexity = complexity;
            }
            
            // AI-style confidence scoring
            baseLabels.confidence = this.calculateConfidence(content);
            
            return baseLabels;
        } catch (error) {
            console.warn('[MockAI] Label suggestion failed, using fallback:', error);
            return this.fallbackLabels(content, name);
        }
    }
    
    async extractSessionContext(content: string): Promise<{
        name: string;
        labels: Record<string, string>;
        summary: string;
    }> {
        await this.delay(200);
        
        try {
            const name = await this.suggestName(content);
            const labels = await this.suggestLabels(content, name);
            
            // Generate summary
            const sentences = content.split(/[.!?]+/).filter(s => s.trim().length > 20);
            const keySentences = sentences.slice(0, 3); // Take first 3 meaningful sentences
            const summary = keySentences.join('. ').trim() + (keySentences.length > 0 ? '.' : '');
            
            return {
                name,
                labels: { ...labels, extracted: 'auto' },
                summary: summary || content.substring(0, 200) + '...'
            };
        } catch (error) {
            console.warn('[MockAI] Session context extraction failed:', error);
            throw error;
        }
    }
    
    private assessComplexity(content: string): 'simple' | 'medium' | 'complex' {
        const lines = content.split('\n').length;
        const words = content.split(/\s+/).length;
        const specialChars = (content.match(/[{}()[\];]/g) || []).length;
        
        if (lines > 50 || words > 500 || specialChars > 20) {
            return 'complex';
        } else if (lines > 20 || words > 150 || specialChars > 5) {
            return 'medium';
        }
        return 'simple';
    }
    
    private calculateConfidence(content: string): string {
        // Simple confidence based on content characteristics
        let score = 0.5;
        
        if (content.length > 100) score += 0.1;
        if (content.includes('\n')) score += 0.1;
        const matches = content.match(/\w{3,}/g);
        if (matches && matches.length > 10) score += 0.1;
        if (this.detectLanguage(content)) score += 0.1;
        
        if (score >= 0.8) return 'high';
        if (score >= 0.6) return 'medium';
        return 'low';
    }
    
    private delay(ms: number): Promise<void> {
        return new Promise(resolve => setTimeout(resolve, ms));
    }
}

/**
 * AI Assistant Factory
 */
export class AIAssistantFactory {
    static create(config: AIAssistantConfig): MemoryAssistant {
        switch (config.provider) {
            case 'mock':
                return new MockAIAssistant(config);
            case 'openai':
                // TODO: Implement OpenAI assistant
                console.warn('[AIFactory] OpenAI provider not yet implemented, using mock');
                return new MockAIAssistant(config);
            case 'anthropic':
                // TODO: Implement Anthropic assistant
                console.warn('[AIFactory] Anthropic provider not yet implemented, using mock');
                return new MockAIAssistant(config);
            case 'local':
                // TODO: Implement local LLM assistant
                console.warn('[AIFactory] Local provider not yet implemented, using mock');
                return new MockAIAssistant(config);
            default:
                return new MockAIAssistant(config);
        }
    }
    
    static createDefault(): MemoryAssistant {
        return new MockAIAssistant({
            provider: 'mock',
            enableFallbacks: true
        });
    }
}
