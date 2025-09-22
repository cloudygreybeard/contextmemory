#!/usr/bin/env node
// Test script to verify version compatibility logic

function parseVersion(version) {
    const cleanVersion = version.replace(/^v/, '').trim();
    const parts = cleanVersion.split('.');
    
    if (parts.length !== 3) {
        throw new Error(`Invalid version format: ${version}`);
    }

    return {
        major: parseInt(parts[0], 10),
        minor: parseInt(parts[1], 10),
        patch: parseInt(parts[2], 10)
    };
}

function checkCompatibility(cliVersion, extensionVersion) {
    const cli = parseVersion(cliVersion);
    const ext = parseVersion(extensionVersion);

    const compatible = cli.major === ext.major && cli.minor === ext.minor;

    return {
        compatible,
        reason: compatible ? 'Compatible' : 
            `Extension v${extensionVersion} requires CLI v${ext.major}.${ext.minor}.x, but found v${cliVersion}`
    };
}

// Test cases
const testCases = [
    { cli: '0.6.1', ext: '0.6.1', expected: true },  // Same version
    { cli: '0.6.0', ext: '0.6.1', expected: true },  // Same minor, different patch
    { cli: '0.6.2', ext: '0.6.1', expected: true },  // Same minor, different patch
    { cli: '0.5.0', ext: '0.6.1', expected: false }, // Different minor
    { cli: '0.7.0', ext: '0.6.1', expected: false }, // Different minor  
    { cli: '1.0.0', ext: '0.6.1', expected: false }, // Different major
];

console.log('🧪 Version Compatibility Test Results:');
console.log('=====================================');

let allPassed = true;

testCases.forEach((testCase, index) => {
    const result = checkCompatibility(testCase.cli, testCase.ext);
    const passed = result.compatible === testCase.expected;
    allPassed = allPassed && passed;
    
    const status = passed ? '✅ PASS' : '❌ FAIL';
    console.log(`${index + 1}. ${status} CLI v${testCase.cli} ↔ Extension v${testCase.ext}`);
    console.log(`   Expected: ${testCase.expected ? 'Compatible' : 'Incompatible'}`);
    console.log(`   Result: ${result.reason}`);
    console.log('');
});

console.log(`Overall: ${allPassed ? '✅ All tests passed!' : '❌ Some tests failed!'}`);
process.exit(allPassed ? 0 : 1);
