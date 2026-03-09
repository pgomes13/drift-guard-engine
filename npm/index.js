"use strict";
// Programmatic API — wraps the drift-guard binary and returns parsed results.

const path = require("path");
const { spawnSync } = require("child_process");

const BIN = path.join(
  __dirname,
  "bin",
  process.platform === "win32" ? "drift-guard.exe" : "drift-guard"
);

/**
 * Run the drift-guard binary and return its stdout.
 * @param {string[]} args
 * @returns {string}
 */
function run(args) {
  const result = spawnSync(BIN, args, { encoding: "utf8" });
  if (result.error) {
    throw new Error(`drift-guard binary error: ${result.error.message}`);
  }
  if (result.status !== 0 && result.status !== 1) {
    // exit 1 means breaking changes found (--fail-on-breaking), not a crash
    throw new Error(`drift-guard exited with code ${result.status}: ${result.stderr}`);
  }
  return result.stdout;
}

/**
 * Diff two OpenAPI 3.x schemas.
 * @param {string} base  Path to the base (old) schema file
 * @param {string} head  Path to the head (new) schema file
 * @returns {DiffResult}
 */
function compareOpenAPI(base, head) {
  return JSON.parse(run(["openapi", "--base", base, "--head", head, "--format", "json"]));
}

/**
 * Diff two GraphQL SDL schemas.
 * @param {string} base
 * @param {string} head
 * @returns {DiffResult}
 */
function compareGraphQL(base, head) {
  return JSON.parse(run(["graphql", "--base", base, "--head", head, "--format", "json"]));
}

/**
 * Diff two Protobuf schemas.
 * @param {string} base
 * @param {string} head
 * @returns {DiffResult}
 */
function compareGRPC(base, head) {
  return JSON.parse(run(["grpc", "--base", base, "--head", head, "--format", "json"]));
}

/**
 * Scan a directory for source references to breaking changes in a diff result.
 * @param {DiffResult} diffResult
 * @param {string} scanDir  Directory to scan (default ".")
 * @param {{ format?: "text"|"json"|"markdown" }} [options]
 * @returns {Hit[]|string}  Array of hits when format="json", otherwise formatted string
 */
function impact(diffResult, scanDir = ".", options = {}) {
  const { format = "json" } = options;
  const fs = require("fs");
  const os = require("os");

  // Write the diff result to a temp file so we can pass --diff
  const tmp = require("path").join(os.tmpdir(), `drift-guard-diff-${Date.now()}.json`);
  try {
    fs.writeFileSync(tmp, JSON.stringify(diffResult));
    const out = run(["impact", "--diff", tmp, "--scan", scanDir, "--format", format]);
    if (format === "json") {
      return JSON.parse(out);
    }
    return out;
  } finally {
    fs.rmSync(tmp, { force: true });
  }
}

module.exports = { compareOpenAPI, compareGraphQL, compareGRPC, impact };
