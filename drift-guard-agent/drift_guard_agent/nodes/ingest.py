"""Ingest node: parse drift-guard JSON diff into DriftState."""

from __future__ import annotations

import json

from drift_guard_agent.state import Change, DiffResult, DriftState


def ingest(state: DriftState) -> dict:
    """Parse raw JSON from stdin/file into a DiffResult. Already set if passed directly."""
    # diff may already be a DiffResult (set by CLI) or a raw dict
    raw = state.get("diff")
    if isinstance(raw, DiffResult):
        return {}

    if isinstance(raw, dict):
        changes = [
            Change(
                type=c.get("type", ""),
                severity=c.get("severity", "info"),
                path=c.get("path", ""),
                method=c.get("method", ""),
                location=c.get("location", ""),
                description=c.get("description", ""),
                before=c.get("before", ""),
                after=c.get("after", ""),
            )
            for c in raw.get("changes", [])
        ]
        diff = DiffResult(
            base_file=raw.get("base_file", ""),
            head_file=raw.get("head_file", ""),
            changes=changes,
            summary=raw.get("summary", {}),
        )
        return {"diff": diff}

    raise ValueError(f"Unexpected diff type: {type(raw)}")


def parse_diff_json(raw_json: str) -> DiffResult:
    """Helper: parse a JSON string into a DiffResult."""
    data = json.loads(raw_json)
    changes = [
        Change(
            type=c.get("type", ""),
            severity=c.get("severity", "info"),
            path=c.get("path", ""),
            method=c.get("method", ""),
            location=c.get("location", ""),
            description=c.get("description", ""),
            before=c.get("before", ""),
            after=c.get("after", ""),
        )
        for c in data.get("changes", [])
    ]
    return DiffResult(
        base_file=data.get("base_file", ""),
        head_file=data.get("head_file", ""),
        changes=changes,
        summary=data.get("summary", {}),
    )
