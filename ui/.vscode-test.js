const { defineConfig } = require('@vscode/test-cli');

module.exports = defineConfig({
  files: 'dist/test/**/*.test.js',
  version: 'insiders',
  workspaceFolder: '.',
  launchArgs: [
    '--disable-extensions',
    '--disable-workspace-trust',
    '--no-sandbox',
    '--disable-dev-shm-usage'
  ],
  mocha: {
    ui: 'tdd',
    timeout: 30000,
    reporter: 'spec'
  }
});
