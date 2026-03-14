#!/usr/bin/env node
// Thin shim: passes all arguments to the drift-agent binary.
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");

const bin = path.join(
  __dirname,
  "bin",
  process.platform === "win32" ? "drift-agent.exe" : "drift-agent"
);

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  process.stderr.write(
    `drift-agent: could not run binary: ${result.error.message}\n` +
    `Make sure the package installed correctly or download manually:\n` +
    `https://github.com/DriftaBot/driftabot-engine/releases\n`
  );
  process.exit(1);
}

process.exit(result.status ?? 0);
