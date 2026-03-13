"""Fetch node: shallow-clone each consumer repo."""

from __future__ import annotations

import tempfile
from pathlib import Path

import git

from drift_guard_agent.state import ConsumerRepo, DriftState

_WORKDIR = Path(tempfile.gettempdir()) / "drift-guard-agent-clones"


def fetch_consumers(state: DriftState) -> dict:
    consumers = state.get("consumers", [])
    if not consumers:
        return {}

    _WORKDIR.mkdir(parents=True, exist_ok=True)
    updated: list[ConsumerRepo] = []

    for consumer in consumers:
        dest = _WORKDIR / consumer.full_name.replace("/", "__")
        try:
            if dest.exists():
                print(f"[fetch] Updating {consumer.full_name}")
                repo = git.Repo(dest)
                repo.remotes.origin.pull(depth=1)
            else:
                print(f"[fetch] Cloning {consumer.full_name}")
                git.Repo.clone_from(
                    consumer.clone_url,
                    dest,
                    depth=1,
                    single_branch=True,
                )
            updated.append(ConsumerRepo(
                full_name=consumer.full_name,
                clone_url=consumer.clone_url,
                local_path=str(dest),
                scan_dir=consumer.scan_dir,
            ))
        except git.GitCommandError as e:
            print(f"[fetch] Failed to clone {consumer.full_name}: {e}")

    return {"consumers": updated}
