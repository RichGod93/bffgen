/**
 * bffgen Programmatic API
 * Use bffgen from JavaScript/TypeScript code
 */

const { execSync: nodeExecSync, spawn } = require("child_process");
const path = require("path");
const { getBinaryName } = require("../scripts/platform");

const BINARY_PATH = path.join(__dirname, "..", "bin", getBinaryName());

/**
 * Execute bffgen command synchronously
 * @param {string[]} args - Command arguments
 * @param {Object} options - Execution options
 * @returns {Buffer} - Command output
 */
function execBffgenSync(args, options = {}) {
  const command = `"${BINARY_PATH}" ${args.join(" ")}`;
  return nodeExecSync(command, {
    encoding: "utf8",
    ...options,
  });
}

/**
 * Execute bffgen command asynchronously
 * @param {string[]} args - Command arguments
 * @param {Object} options - Execution options
 * @returns {Promise<string>} - Command output
 */
function exec(args, options = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(BINARY_PATH, args, {
      ...options,
      cwd: options.cwd || process.cwd(),
    });

    let stdout = "";
    let stderr = "";

    if (child.stdout) {
      child.stdout.on("data", (data) => {
        stdout += data.toString();
      });
    }

    if (child.stderr) {
      child.stderr.on("data", (data) => {
        stderr += data.toString();
      });
    }

    child.on("error", reject);

    child.on("exit", (code) => {
      if (code !== 0) {
        const error = new Error(`bffgen exited with code ${code}`);
        error.code = code;
        error.stderr = stderr;
        error.stdout = stdout;
        reject(error);
      } else {
        resolve(stdout);
      }
    });
  });
}

/**
 * Initialize a new BFF project
 * @param {Object} options - Project options
 * @param {string} options.name - Project name
 * @param {string} options.lang - Language/runtime (go, nodejs-express, nodejs-fastify)
 * @param {string} options.framework - Framework (chi, echo, fiber, express, fastify)
 * @param {boolean} options.skipTests - Skip test generation
 * @param {boolean} options.skipDocs - Skip documentation generation
 * @returns {Promise<string>}
 */
async function init(options) {
  const args = ["init", options.name];

  if (options.lang) {
    args.push("--lang", options.lang);
  }

  if (options.framework) {
    args.push("--framework", options.framework);
  }

  if (options.skipTests) {
    args.push("--skip-tests");
  }

  if (options.skipDocs) {
    args.push("--skip-docs");
  }

  return exec(args, { stdio: "inherit" });
}

/**
 * Generate code from configuration
 * @param {Object} options - Generation options
 * @returns {Promise<string>}
 */
async function generate(options = {}) {
  const args = ["generate"];

  if (options.check) {
    args.push("--check");
  }

  if (options.dryRun) {
    args.push("--dry-run");
  }

  return exec(args, { stdio: "inherit" });
}

/**
 * Get bffgen version
 * @returns {string}
 */
function getVersion() {
  try {
    const output = execBffgenSync(["version"]);
    const match = output.match(/bffgen version (v?[\d.]+)/);
    return match ? match[1] : require("../package.json").version;
  } catch (error) {
    return require("../package.json").version;
  }
}

module.exports = {
  exec,
  execSync: execBffgenSync,
  init,
  generate,
  getVersion,
  version: require("../package.json").version,
};
