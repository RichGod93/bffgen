#!/usr/bin/env node

/**
 * bffgen npm binary wrapper
 * Executes the platform-specific downloaded binary
 */

const { spawn } = require("child_process");
const path = require("path");
const fs = require("fs");
const { getBinaryName } = require("../scripts/platform");

// Find binary
const binaryName = getBinaryName();
const binaryPath = path.join(__dirname, binaryName);

// Check if binary exists
if (!fs.existsSync(binaryPath)) {
  console.error("❌ bffgen binary not found!");
  console.error(
    "\nThe binary should have been downloaded during installation."
  );
  console.error("Try reinstalling:");
  console.error("  npm install -g bffgen");
  console.error("\nOr install via Go:");
  console.error("  go install github.com/RichGod93/bffgen/cmd/bffgen@latest");
  process.exit(1);
}

// Execute binary with all arguments
const args = process.argv.slice(2);
const child = spawn(binaryPath, args, {
  stdio: "inherit",
  cwd: process.cwd(),
  env: process.env,
});

// Handle exit
child.on("error", (error) => {
  console.error("Failed to execute bffgen:", error.message);
  process.exit(1);
});

child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
  } else {
    process.exit(code || 0);
  }
});

// Handle signals
process.on("SIGINT", () => {
  child.kill("SIGINT");
});

process.on("SIGTERM", () => {
  child.kill("SIGTERM");
});
