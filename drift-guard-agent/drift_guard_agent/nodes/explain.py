"""Explain node: LLM explains each breaking change in context of each consumer's hits.

Optional — only runs when ANTHROPIC_API_KEY is set.
"""

from __future__ import annotations

import json
import os

from drift_guard_agent.state import DriftState


def explain(state: DriftState) -> dict:
    if not os.environ.get("ANTHROPIC_API_KEY"):
        return {"explanations": {}}

    import anthropic
    client = anthropic.Anthropic()

    diff = state["diff"]
    hits_by_repo = state.get("hits", {})
    model = state.get("model", "claude-opus-4-6")

    if not hits_by_repo:
        return {"explanations": {}}

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    explanations: dict[str, list[str]] = {}

    for repo, hits in hits_by_repo.items():
        hits_text = "\n".join(
            f"  {h.file}:{h.line_num}  [{h.change_path}]  `{h.line.strip()}`"
            for h in hits[:30]
        )
        changes_text = "\n".join(
            f"- [{c.type}] {c.method} {c.path}: {c.description}"
            for c in breaking
        )

        prompt = f"""You are a senior API integration engineer reviewing breaking API changes and their impact on a consumer codebase.

Breaking changes in the provider API:
{changes_text}

Affected lines in {repo}:
{hits_text}

For each breaking change that has at least one matching hit, write a concise 1-2 sentence explanation of the real impact on this consumer. Focus on what will break at runtime. Skip changes with no hits.

Respond as a JSON array of strings, one per affected breaking change (same order, omit changes with no hits):
["explanation 1", "explanation 2"]"""

        response = client.messages.create(
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
            repo_explanations = json.loads(text.strip())
        except json.JSONDecodeError:
            repo_explanations = [text]

        explanations[repo] = repo_explanations
        print(f"[explain] {repo}: {len(repo_explanations)} explanation(s)")

    return {"explanations": explanations}
