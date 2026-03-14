#!/usr/bin/env node
// postinstall: downloads the drift-guard binary from GitHub Releases.
"use strict";

const https = require("https");
const fs = require("fs");
const os = require("os");
const path = require("path");
const { execSync } = require("child_process");

const VERSION = require("./package.json").version;
const REPO = "pgomes13/api-drift-engine";
const BIN_DIR = path.join(__dirname, "bin");
const BIN_PATH = path.join(BIN_DIR, process.platform === "win32" ? "drift-guard.exe" : "drift-guard");

// Map Node.js platform/arch → goreleaser archive naming
function getPlatformInfo() {
  const p = process.platform;
  const a = process.arch;

  const osMap = { linux: "linux", darwin: "darwin", win32: "windows" };
  const archMap = { x64: "amd64", arm64: "arm64" };

  const goos = osMap[p];
  const goarch = archMap[a];

  if (!goos || !goarch) {
    throw new Error(`Unsupported platform: ${p}/${a}`);
  }
  if (goos === "windows" && goarch === "arm64") {
    throw new Error("Windows arm64 is not supported");
  }

  const ext = goos === "windows" ? "zip" : "tar.gz";
  const archive = `drift-guard_${VERSION}_${goos}_${goarch}.${ext}`;
  const url = `https://github.com/${REPO}/releases/download/v${VERSION}/${archive}`;

  return { url, archive, ext };
}

function download(url, dest) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(dest);
    function get(u) {
      https.get(u, (res) => {
        if (res.statusCode === 301 || res.statusCode === 302) {
          return get(res.headers.location);
        }
        if (res.statusCode !== 200) {
          reject(new Error(`Download failed: HTTP ${res.statusCode} for ${u}`));
          return;
        }
        res.pipe(file);
        file.on("finish", () => file.close(resolve));
      }).on("error", reject);
    }
    get(url);
  });
}

async function install() {
  if (fs.existsSync(BIN_PATH)) {
    return; // already installed (e.g. cached node_modules)
  }

  const { url, archive, ext } = getPlatformInfo();
  const tmpDir = fs.mkdtempSync(path.join(os.tmpdir(), "drift-guard-"));
  const archivePath = path.join(tmpDir, archive);

  process.stdout.write(`Downloading drift-guard v${VERSION}...\n`);

  try {
    await download(url, archivePath);

    fs.mkdirSync(BIN_DIR, { recursive: true });

    if (ext === "tar.gz") {
      execSync(`tar -xzf "${archivePath}" -C "${tmpDir}"`);
    } else {
      execSync(`unzip -o "${archivePath}" -d "${tmpDir}"`);
    }

    const extracted = path.join(tmpDir, process.platform === "win32" ? "drift-guard.exe" : "drift-guard");
    fs.copyFileSync(extracted, BIN_PATH);
    fs.chmodSync(BIN_PATH, 0o755);

    process.stdout.write(`drift-guard installed to ${BIN_PATH}\n`);
  } finally {
    fs.rmSync(tmpDir, { recursive: true, force: true });
  }
}

install().catch((err) => {
  process.stderr.write(`drift-guard install failed: ${err.message}\n`);
  process.stderr.write("You can install it manually: https://github.com/pgomes13/api-drift-engine/releases\n");
  // Do not exit(1) — allow npm install to succeed even if binary download fails.
});
