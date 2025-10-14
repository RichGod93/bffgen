/**
 * Platform Detection Utilities
 * Maps Node.js platform/arch to Go binary names
 */

const os = require("os");

const PLATFORM_MAP = {
  darwin: {
    x64: "darwin-amd64",
    arm64: "darwin-arm64",
  },
  linux: {
    x64: "linux-amd64",
    arm64: "linux-arm64",
  },
  win32: {
    x64: "windows-amd64",
  },
};

const SUPPORTED_PLATFORMS = [
  "darwin-x64",
  "darwin-arm64",
  "linux-x64",
  "linux-arm64",
  "win32-x64",
];

/**
 * Get current platform and architecture
 */
function getPlatform() {
  const platform = os.platform();
  const arch = os.arch();
  return { platform, arch };
}

/**
 * Map Node.js platform/arch to Go binary name
 */
function getBinaryName() {
  const { platform, arch } = getPlatform();

  if (!PLATFORM_MAP[platform] || !PLATFORM_MAP[platform][arch]) {
    throw new Error(
      `Unsupported platform: ${platform}-${arch}\n` +
        `Supported platforms: ${SUPPORTED_PLATFORMS.join(", ")}\n` +
        `Please install manually from: https://github.com/RichGod93/bffgen/releases`
    );
  }

  const goPlatform = PLATFORM_MAP[platform][arch];
  const ext = platform === "win32" ? ".exe" : "";

  return `bffgen-${goPlatform}${ext}`;
}

/**
 * Get download URL for current platform
 */
function getDownloadUrl(version) {
  const binaryName = getBinaryName();
  return `https://github.com/RichGod93/bffgen/releases/download/v${version}/${binaryName}`;
}

/**
 * Get checksums URL for version
 */
function getChecksumsUrl(version) {
  return `https://github.com/RichGod93/bffgen/releases/download/v${version}/checksums.txt`;
}

/**
 * Check if platform is supported
 */
function isSupported() {
  try {
    getBinaryName();
    return true;
  } catch (error) {
    return false;
  }
}

module.exports = {
  getPlatform,
  getBinaryName,
  getDownloadUrl,
  getChecksumsUrl,
  isSupported,
  SUPPORTED_PLATFORMS,
};
