#!/usr/bin/env node

/**
 * Post-install script for bffgen npm package
 * Downloads the appropriate binary from GitHub Releases
 */

const https = require("https");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");
const {
  getBinaryName,
  getDownloadUrl,
  getChecksumsUrl,
  isSupported,
  SUPPORTED_PLATFORMS,
} = require("./platform");

const PACKAGE_VERSION = require("../package.json").version;
const BIN_DIR = path.join(__dirname, "..", "bin");
const BINARY_PATH = path.join(BIN_DIR, getBinaryName());

// Colors for output
const colors = {
  reset: "\x1b[0m",
  blue: "\x1b[34m",
  green: "\x1b[32m",
  red: "\x1b[31m",
  yellow: "\x1b[33m",
};

function log(message, color = "reset") {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

/**
 * Check if platform is supported
 */
function checkPlatform() {
  if (!isSupported()) {
    log("❌ Platform not supported", "red");
    log(`\nSupported platforms: ${SUPPORTED_PLATFORMS.join(", ")}`, "yellow");
    log("\n📦 Manual Installation:", "blue");
    log(
      `   Download from: https://github.com/RichGod93/bffgen/releases/v${PACKAGE_VERSION}`
    );
    log("   Extract and add to PATH");
    process.exit(1);
  }
}

/**
 * Download file from URL
 */
function downloadFile(url) {
  return new Promise((resolve, reject) => {
    log(`📥 Downloading from: ${url}`, "blue");

    https
      .get(url, { timeout: 30000 }, (response) => {
        if (response.statusCode === 302 || response.statusCode === 301) {
          // Follow redirect
          https
            .get(
              response.headers.location,
              { timeout: 30000 },
              (redirectResponse) => {
                if (redirectResponse.statusCode !== 200) {
                  reject(
                    new Error(
                      `HTTP ${redirectResponse.statusCode}: ${redirectResponse.statusMessage}`
                    )
                  );
                  return;
                }

                const chunks = [];
                redirectResponse.on("data", (chunk) => chunks.push(chunk));
                redirectResponse.on("end", () =>
                  resolve(Buffer.concat(chunks))
                );
                redirectResponse.on("error", reject);
              }
            )
            .on("error", reject);
        } else if (response.statusCode === 200) {
          const chunks = [];
          response.on("data", (chunk) => chunks.push(chunk));
          response.on("end", () => resolve(Buffer.concat(chunks)));
          response.on("error", reject);
        } else {
          reject(
            new Error(
              `HTTP ${response.statusCode}: ${response.statusMessage}\nURL: ${url}`
            )
          );
        }
      })
      .on("error", reject);
  });
}

/**
 * Verify binary checksum
 */
async function verifyChecksum(binaryBuffer, binaryName) {
  try {
    log("🔐 Verifying checksum...", "blue");

    const checksumsUrl = getChecksumsUrl(PACKAGE_VERSION);
    const checksumsData = await downloadFile(checksumsUrl);
    const checksumsText = checksumsData.toString("utf8");

    // Find checksum for our binary
    const lines = checksumsText.split("\n");
    const checksumLine = lines.find((line) => line.includes(binaryName));

    if (!checksumLine) {
      log("⚠️  Warning: Checksum not found in checksums.txt", "yellow");
      return true; // Continue anyway
    }

    const expectedChecksum = checksumLine.split(/\s+/)[0];

    // Calculate actual checksum
    const hash = crypto.createHash("sha256");
    hash.update(binaryBuffer);
    const actualChecksum = hash.digest("hex");

    if (expectedChecksum !== actualChecksum) {
      log("❌ Checksum verification failed!", "red");
      log(`   Expected: ${expectedChecksum}`, "red");
      log(`   Actual:   ${actualChecksum}`, "red");
      log("\n⚠️  Binary may be corrupted. Try reinstalling:", "yellow");
      log("   npm cache clean --force", "yellow");
      log("   npm install -g bffgen", "yellow");
      return false;
    }

    log("✅ Checksum verified", "green");
    return true;
  } catch (error) {
    log(`⚠️  Warning: Could not verify checksum: ${error.message}`, "yellow");
    return true; // Continue anyway
  }
}

/**
 * Main installation process
 */
async function install() {
  try {
    log(`\n📦 Installing bffgen v${PACKAGE_VERSION}...`, "blue");

    // Check platform support
    checkPlatform();

    const binaryName = getBinaryName();
    log(`   Platform: ${binaryName}`, "blue");

    // Create bin directory
    if (!fs.existsSync(BIN_DIR)) {
      fs.mkdirSync(BIN_DIR, { recursive: true });
    }

    // Download binary
    const downloadUrl = getDownloadUrl(PACKAGE_VERSION);
    const binaryBuffer = await downloadFile(downloadUrl);

    // Verify checksum
    const checksumValid = await verifyChecksum(binaryBuffer, binaryName);
    if (!checksumValid) {
      process.exit(1);
    }

    // Write binary file
    fs.writeFileSync(BINARY_PATH, binaryBuffer);

    // Make executable (Unix only)
    if (process.platform !== "win32") {
      fs.chmodSync(BINARY_PATH, 0o755);
    }

    log("✅ bffgen installed successfully!", "green");
    log("\n📚 Quick Start:", "blue");
    log("   npx bffgen init my-project --lang nodejs-express", "blue");
    log("   npx bffgen --help", "blue");
    log("\n📖 Documentation: https://github.com/RichGod93/bffgen", "blue");
  } catch (error) {
    log("\n❌ Installation failed!", "red");
    log(`   Error: ${error.message}`, "red");
    log("\n📦 Manual Installation:", "yellow");
    log(
      `   1. Download from: https://github.com/RichGod93/bffgen/releases/v${PACKAGE_VERSION}`,
      "yellow"
    );
    log(`   2. Extract the binary for your platform`, "yellow");
    log(`   3. Add to your PATH`, "yellow");
    log("\n💡 Or install via Go:", "yellow");
    log(
      `   go install github.com/RichGod93/bffgen/cmd/bffgen@v${PACKAGE_VERSION}`,
      "yellow"
    );

    process.exit(1);
  }
}

// Run installation
install();
