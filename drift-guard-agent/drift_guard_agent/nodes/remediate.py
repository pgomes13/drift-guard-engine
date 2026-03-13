"""Remediate node: LLM suggests a fix per consumer per breaking change."""

from __future__ import annotations

import json

import anthropic

from drift_guard_agent.state import DriftState

_client = anthropic.Anthropic()


def remediate(state: DriftState) -> dict:
    diff = state["diff"]
    hits_by_repo = state.get("hits", {})
    triage = state.get("triage", [])
    model = state.get("model", "claude-opus-4-6")

    if not hits_by_repo:
        return {"remediations": {}}

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    remediations: dict[str, list[str]] = {}

    risk_map = {s.change_index: s.risk for s in triage}

    for repo, hits in hits_by_repo.items():
        hits_text = "\n".join(
            f"  {h.file}:{h.line_num}  `{h.line.strip()}`"
            for h in hits[:20]
        )
        changes_text = "\n".join(
            f"{i}. [{risk_map.get(i, 'unknown')} risk] [{c.type}] {c.method} {c.path}: {c.description}"
            + (f"\n   Before: {c.before!r}\n   After:  {c.after!r}" if c.before or c.after else "")
            for i, c in enumerate(breaking)
        )

        prompt = f"""You are a senior API engineer. A provider API has breaking changes and the consumer repo {repo} is affected.

Breaking changes (with risk scores):
{changes_text}

Affected lines in {repo}:
{hits_text}

For each breaking change that has hits in this repo, suggest a concise, actionable remediation. Options include:
- Update the consumer code to use the new API
- Add a backward-compat shim/adapter
- Pin to the previous API version while migrating
- Add a deprecation warning and schedule migration

Be specific. Reference the actual changed field/endpoint where possible.

Respond as a JSON array of strings, one per affected breaking change (same order, skip changes with no hits):
["remediation 1", "remediation 2"]"""

        response = _client.messages.create(
            model=model,
            max_tokens=1024,
            thinking={"type": "adaptive"},
            messages=[{"role": "user", "content": prompt}],
        )

        text = next((b.text for b in response.content if b.type == "text"), "[]").strip()
        if text.startswith("```"):
            text = text.split("```")[1]
            if text.startswith("json"):
                text = text[4:]

        try:
            repo_remediations = json.loads(text.strip())
        except json.JSONDecodeError:
            repo_remediations = [text]

        remediations[repo] = repo_remediations
        print(f"[remediate] {repo}: {len(repo_remediations)} suggestion(s)")

    return {"remediations": remediations}
