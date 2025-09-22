const { defineConfig } = require('@vscode/test-cli');

module.exports = defineConfig({
  files: 'dist/test/**/*.test.js',
  version: 'insiders',
  workspaceFolder: '.',
  mocha: {
    ui: 'tdd',
    timeout: 20000,
    reporter: 'spec'
  }
});
