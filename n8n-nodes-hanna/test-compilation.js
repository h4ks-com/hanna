// Test script to validate the n8n node compilation
const fs = require('fs');
const path = require('path');

console.log('🔍 Testing n8n Hanna Bot Node Compilation...\n');

// Check if compiled files exist
const distPath = path.join(__dirname, 'dist');
const hannaNodePath = path.join(distPath, 'nodes', 'HannaBot', 'HannaBot.node.js');

if (!fs.existsSync(hannaNodePath)) {
    console.error('❌ Compiled node file not found');
    process.exit(1);
}

// Read the compiled file
const nodeContent = fs.readFileSync(hannaNodePath, 'utf8');

// Check for new operations
const hasListOperation = nodeContent.includes("'List Channels'") && nodeContent.includes("value: 'list'");
const hasWhoisOperation = nodeContent.includes("'Get User Info (WHOIS)'") && nodeContent.includes("value: 'whois'");

// Check for implementation
const hasListCase = nodeContent.includes("case 'list'") && nodeContent.includes("'/api/list'");
const hasWhoisCase = nodeContent.includes("case 'whois'") && nodeContent.includes("'/api/whois'");

// Check for parameter handling
const hasWhoisParam = nodeContent.includes("'whoisNick'");

console.log('✅ Compilation Results:');
console.log(`   📋 List Channels operation: ${hasListOperation ? '✓' : '✗'}`);
console.log(`   👤 WHOIS operation: ${hasWhoisOperation ? '✓' : '✗'}`);
console.log(`   📋 List implementation: ${hasListCase ? '✓' : '✗'}`);
console.log(`   👤 WHOIS implementation: ${hasWhoisCase ? '✓' : '✗'}`);
console.log(`   🔧 WHOIS parameter: ${hasWhoisParam ? '✓' : '✗'}`);

// Check for advanced response handling
const hasAdvancedListHandling = nodeContent.includes('channelCount') && nodeContent.includes('largestChannels');
const hasAdvancedWhoisHandling = nodeContent.includes('userChannels') && nodeContent.includes('isOperator');

console.log('\n✅ Advanced Features:');
console.log(`   📊 List response processing: ${hasAdvancedListHandling ? '✓' : '✗'}`);
console.log(`   🔍 WHOIS response processing: ${hasAdvancedWhoisHandling ? '✓' : '✗'}`);

const allTestsPassed = hasListOperation && hasWhoisOperation && hasListCase && 
                       hasWhoisCase && hasWhoisParam && hasAdvancedListHandling && 
                       hasAdvancedWhoisHandling;

if (allTestsPassed) {
    console.log('\n🎉 All tests passed! The n8n node is ready for deployment.');
} else {
    console.log('\n❌ Some tests failed. Please check the compilation.');
    process.exit(1);
}

// Check package version
const packagePath = path.join(__dirname, 'package.json');
const packageData = JSON.parse(fs.readFileSync(packagePath, 'utf8'));
console.log(`\n📦 Package version: ${packageData.version}`);
console.log(`📝 Package name: ${packageData.name}`);

console.log('\n🚀 Ready for npm publish!');
