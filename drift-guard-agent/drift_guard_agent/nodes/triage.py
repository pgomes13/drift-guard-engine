"""Triage node: LLM scores each breaking change as high / medium / low risk."""

from __future__ import annotations

import json

import anthropic

from drift_guard_agent.state import DriftState, RiskScore

_client = anthropic.Anthropic()


def triage(state: DriftState) -> dict:
    diff = state["diff"]
    breaking = [c for c in diff.changes if c.severity == "breaking"]
    if not breaking:
        return {"triage": [], "max_risk": "low"}

    changes_text = "\n".join(
        f"{i}. [{c.type}] {c.method} {c.path} — {c.description}"
        + (f" (before: {c.before!r}, after: {c.after!r})" if c.before or c.after else "")
        for i, c in enumerate(breaking)
    )

    prompt = f"""You are an API risk analyst. For each breaking API change below, rate the real-world risk to consumers as high, medium, or low.

Rules:
- high: endpoint removed, required field removed, response type changed, authentication changed
- medium: optional field removed, enum value removed, response field renamed
- low: description/documentation-only change, field addition to response (shouldn't break consumers but classified as breaking)

Breaking changes:
{changes_text}

Respond with a JSON array only, one object per change, in the same order:
[{{"index": 0, "risk": "high", "reason": "..."}}]"""

    response = _client.messages.create(
        model=state.get("model", "claude-opus-4-6"),
        max_tokens=1024,
        thinking={"type": "adaptive"},
        messages=[{"role": "user", "content": prompt}],
    )

    text = next(
        (b.text for b in response.content if b.type == "text"), "[]"
    )
    # Strip markdown fences if present
    text = text.strip()
    if text.startswith("```"):
        text = text.split("```")[1]
        if text.startswith("json"):
            text = text[4:]

    raw_scores = json.loads(text.strip())
    scores = [
        RiskScore(
            change_index=s["index"],
            risk=s["risk"],
            reason=s.get("reason", ""),
        )
        for s in raw_scores
    ]

    risk_order = {"high": 2, "medium": 1, "low": 0}
    max_risk = max(scores, key=lambda s: risk_order.get(s.risk, 0)).risk if scores else "low"

    return {"triage": scores, "max_risk": max_risk}
