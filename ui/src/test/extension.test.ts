import * as assert from 'assert';
import * as vscode from 'vscode';

suite('Extension Test Suite', () => {
	vscode.window.showInformationMessage('Start all tests.');

	test('Extension should be present', () => {
		assert.ok(vscode.extensions.getExtension('cloudygreybeard.contextmemory'));
	});

	test('Should activate extension', async () => {
		const extension = vscode.extensions.getExtension('cloudygreybeard.contextmemory');
		if (extension) {
			await extension.activate();
			assert.ok(extension.isActive);
		}
	});
});
