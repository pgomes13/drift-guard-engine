#!/usr/bin/env node
// Thin shim: passes all arguments to the drift-guard binary.
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");

const bin = path.join(
  __dirname,
  "bin",
  process.platform === "win32" ? "drift-guard.exe" : "drift-guard"
);

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  process.stderr.write(
    `drift-guard: could not run binary: ${result.error.message}\n` +
    `Make sure the package installed correctly or download manually:\n` +
    `https://github.com/pgomes13/api-drift-engine/releases\n`
  );
  process.exit(1);
}

process.exit(result.status ?? 0);
