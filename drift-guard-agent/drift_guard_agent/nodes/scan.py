"""Scan node: grep consumer checkouts for references to broken API paths."""

from __future__ import annotations

import re
from pathlib import Path

from drift_guard_agent.state import DriftState, Hit

# File extensions to scan
_SCAN_EXTENSIONS = {
    ".js", ".ts", ".jsx", ".tsx",
    ".py", ".go", ".java", ".kt",
    ".rb", ".php", ".cs", ".rs",
    ".swift", ".c", ".cpp", ".h",
    ".json", ".yaml", ".yml",
    ".sh", ".env", ".txt", ".md",
}

_SKIP_DIRS = {"node_modules", ".git", "__pycache__", ".venv", "vendor", "dist", "build"}


def scan_consumers(state: DriftState) -> dict:
    consumers = state.get("consumers", [])
    diff = state["diff"]

    if not consumers or not diff:
        return {"hits": {}}

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    if not breaking:
        return {"hits": {}}

    # Build search patterns from breaking change paths
    # /users/{id} → search for "/users" and "/users/"
    patterns = _build_patterns(breaking)

    hits_by_repo: dict[str, list[Hit]] = {}

    for consumer in consumers:
        if not consumer.local_path:
            continue
        scan_dir = Path(consumer.local_path) / consumer.scan_dir
        if not scan_dir.exists():
            scan_dir = Path(consumer.local_path)

        repo_hits = _scan_dir(scan_dir, patterns, breaking)
        if repo_hits:
            hits_by_repo[consumer.full_name] = repo_hits
            print(f"[scan] {consumer.full_name}: {len(repo_hits)} hit(s)")
        else:
            print(f"[scan] {consumer.full_name}: no hits")

    return {"hits": hits_by_repo}


def _build_patterns(breaking):
    """Return (compiled_regex, change) pairs for each breaking change path."""
    results = []
    for c in breaking:
        # Strip path params to get stable prefix: /users/{id} → /users
        parts = [p for p in c.path.split("/") if p and not p.startswith("{")]
        if not parts:
            continue
        # Match the stable prefix anywhere in a line (quoted or unquoted)
        stable = "/" + "/".join(parts)
        pattern = re.compile(re.escape(stable), re.IGNORECASE)
        results.append((pattern, stable, c))
    return results


def _scan_dir(scan_dir: Path, patterns, breaking) -> list[Hit]:
    hits: list[Hit] = []
    for fpath in _walk(scan_dir):
        try:
            text = fpath.read_text(encoding="utf-8", errors="ignore")
        except OSError:
            continue
        rel = str(fpath.relative_to(scan_dir))
        for i, line in enumerate(text.splitlines(), 1):
            for pattern, stable, change in patterns:
                if pattern.search(line):
                    hits.append(Hit(
                        file=rel,
                        line_num=i,
                        line=line,
                        change_type=change.type,
                        change_path=f"{change.method} {change.path}",
                    ))
                    break  # one hit per line is enough
    return hits


def _walk(directory: Path):
    for item in directory.rglob("*"):
        if item.is_file() and item.suffix in _SCAN_EXTENSIONS:
            if not any(part in _SKIP_DIRS for part in item.parts):
                yield item
