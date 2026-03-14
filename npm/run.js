#!/usr/bin/env node
// Thin shim: passes all arguments to the driftabot binary.
"use strict";

const path = require("path");
const { spawnSync } = require("child_process");

const bin = path.join(
  __dirname,
  "bin",
  process.platform === "win32" ? "driftabot.exe" : "driftabot"
);

const result = spawnSync(bin, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  process.stderr.write(
    `driftabot: could not run binary: ${result.error.message}\n` +
    `Make sure the package installed correctly or download manually:\n` +
    `https://github.com/DriftaBot/driftabot-engine/releases\n`
  );
  process.exit(1);
}

process.exit(result.status ?? 0);
