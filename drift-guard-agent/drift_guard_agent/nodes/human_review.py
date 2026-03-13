"""Human review node: LangGraph interrupt for high-risk changes."""

from __future__ import annotations

from langgraph.types import interrupt

from drift_guard_agent.state import DriftState


def human_review(state: DriftState) -> dict:
    """Pause the graph and wait for human approval before notifying consumers.

    Only reached when max_risk == 'high'. The graph is interrupted here and
    resumed by passing {"human_approved": True/False} via the LangGraph API
    or CLI --approve flag.
    """
    diff = state["diff"]
    hits_by_repo = state.get("hits", {})
    triage = state.get("triage", [])
    breaking = [c for c in diff.changes if c.severity == "breaking"]

    high_risk = [s for s in triage if s.risk == "high"]
    affected_repos = list(hits_by_repo.keys())

    summary_lines = [
        f"**drift-guard-agent** detected {len(high_risk)} high-risk breaking change(s):",
        "",
    ]
    for s in high_risk:
        if s.change_index < len(breaking):
            c = breaking[s.change_index]
            summary_lines.append(f"- `{c.method} {c.path}` — {c.description} _(reason: {s.reason})_")

    summary_lines += [
        "",
        f"**Affected consumers ({len(affected_repos)}):** {', '.join(affected_repos)}",
        "",
        "Approve to post GitHub Issues + PR comment + Slack notification.",
        "Reject to abort without notifying consumers.",
    ]

    decision = interrupt({
        "question": "Approve consumer notifications for these high-risk breaking changes?",
        "summary": "\n".join(summary_lines),
        "affected_repos": affected_repos,
        "high_risk_count": len(high_risk),
    })

    approved = decision if isinstance(decision, bool) else str(decision).lower() in ("true", "yes", "approve", "1")
    print(f"[human_review] Decision: {'approved' if approved else 'rejected'}")
    return {"human_approved": approved}
