const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const TEST_ROOT = path.join(__dirname, 'unit');
const tests = new Map(); // id -> test
const completed = new Set(); // ids

// Parse command line arguments for specific test files
const specificFiles = process.argv.slice(2).map(f => 
  f.endsWith('.test.js') ? f : (f.endsWith('.js') ? f : f + '.test.js')
);

function loadTests() {
  tests.clear();
  completed.clear();

  function walk(dir) {
    for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
      const fullPath = path.join(dir, entry.name);
      if (entry.isDirectory()) {
        walk(fullPath);
        continue;
      }
      if (!entry.name.endsWith('.test.js')) continue;

      // If specific files are requested, only load those
      if (specificFiles.length > 0 && !specificFiles.some(f => entry.name === f || entry.name.endsWith(f))) {
        continue;
      }

      delete require.cache[require.resolve(fullPath)];
      const mod = require(fullPath);
      if (!Array.isArray(mod.tests)) continue;

      for (const test of mod.tests) {
        test._logs = [];
        tests.set(test.id, test);
      }
    }
  }

  walk(TEST_ROOT);

  if (tests.size === 0) {
    console.error('❌ No tests found');
    if (specificFiles.length > 0) {
      console.error('   Requested files:', specificFiles);
    }
    process.exit(1);
  }
}

loadTests();

const oslPath = path.join(__dirname, '..', 'osl');
const failed = [];
let passed = 0;
const start = performance.now();

for (const test of tests.values()) {
  // Write test code to temp file
  const tempDir = fs.mkdtempSync('osl-test-');
  const testFile = path.join(tempDir, 'test.osl');
  fs.writeFileSync(testFile, test.code);

  try {
    // Run the test
    const output = execSync(`"${oslPath}" run "${testFile}"`, {
      cwd: __dirname,
      encoding: 'utf-8',
      stdio: ['pipe', 'pipe', 'pipe'],
      timeout: 30000
    });

    // Normalize output
    const logs = output
      .trim()
      .split('\n')
      .map(line => {
        try {
          // Try to parse as JSON for numbers/booleans
          return JSON.parse(line);
        } catch {
          return line;
        }
      });

    test._logs = logs;
    completed.add(test.id);

    // Check if pass
    const isPass = JSON.stringify(test._logs) === JSON.stringify(test.expect);

    if (isPass) {
      console.log(`✅ [${test.name}]`);
      passed++;
    } else {
      console.log(`❌ [${test.name}]`);
      console.log('   Expected:', test.expect);
      console.log('   Received:', test._logs);
      failed.push(test.name);
    }
  } catch (error) {
    const errorMsg = error.stderr || error.message || String(error);
    test._logs = ['Error:', errorMsg];
    completed.add(test.id);
    console.log(`❌ [${test.name}]`);
    console.log('   Error:', errorMsg.split('\n').slice(0, 3).join('\n   '));
    failed.push(test.name);
  } finally {
    // Cleanup temp directory
    try {
      fs.rmSync(tempDir, { recursive: true, force: true });
    } catch (e) {
      // Ignore cleanup errors
    }
  }
}

console.log(`\n🌈 Test Results (${tests.size} tests)`);
console.log(` Passed: ${passed}`);
console.log(` Failed: ${failed.length}`);
console.log(` Total: ${tests.size}`);

const duration = Math.round(performance.now() - start);
console.log(` Time: ${duration}ms`);

process.exit(failed.length ? 1 : 0);
