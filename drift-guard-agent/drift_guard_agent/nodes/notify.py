"""Notify node: open / update a GitHub Issue in each affected consumer repo."""

from __future__ import annotations

import os

import httpx

from drift_guard_agent.state import DriftState

_GITHUB_API = "https://api.github.com"
_ISSUE_LABEL = "drift-guard"


def notify(state: DriftState) -> dict:
    diff = state["diff"]
    hits_by_repo = state.get("hits", {})
    explanations = state.get("explanations", {})
    dry_run = state.get("dry_run", False)
    github_token = state.get("github_token", "") or os.environ.get("GITHUB_TOKEN", "")
    pr_number = state.get("pr_number", 0)
    provider_repo = state.get("provider_repo", "") or os.environ.get("GITHUB_REPOSITORY", "")

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    consumer_issues: dict[str, str] = {}

    for repo, hits in hits_by_repo.items():
        body = _build_issue_body(
            repo=repo,
            hits=hits,
            breaking=breaking,
            explanations=explanations.get(repo, []),
            provider_repo=provider_repo,
            pr_number=pr_number,
        )
        consumer_issues[repo] = body

    if dry_run:
        print("\n[notify] DRY RUN — no GitHub requests sent\n")
        print("=" * 60)
        for repo, body in consumer_issues.items():
            print(f"\nISSUE [{repo}]:\n{body}\n")
        print("=" * 60)
        return {"consumer_issues": consumer_issues}

    if not github_token:
        print("[notify] No GITHUB_TOKEN — skipping issue creation")
        return {"consumer_issues": consumer_issues}

    headers = {
        "Authorization": f"Bearer {github_token}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }

    with httpx.Client(headers=headers, timeout=30) as client:
        for repo, body in consumer_issues.items():
            _upsert_issue(client, repo, body, provider_repo, pr_number)

    return {"consumer_issues": consumer_issues}


def _build_issue_body(
    repo: str,
    hits,
    breaking: list,
    explanations: list[str],
    provider_repo: str,
    pr_number: int,
) -> str:
    pr_link = (
        f"[PR #{pr_number}](https://github.com/{provider_repo}/pull/{pr_number})"
        if provider_repo and pr_number
        else f"PR #{pr_number}" if pr_number else "a provider PR"
    )

    lines = [
        f"## ⚠️ Breaking API changes from `{provider_repo}` ({pr_link})",
        "",
        "Your repository references API endpoints that have been removed or changed.",
        "",
        "### Breaking changes",
        "",
    ]

    for i, c in enumerate(breaking):
        lines.append(f"- `{c.method} {c.path}` — {c.description}")
        if i < len(explanations) and explanations[i]:
            lines.append(f"  > {explanations[i]}")

    lines += [
        "",
        "### Affected files in this repo",
        "",
        "| File | Line | Referenced path |",
        "| ---- | ---- | --------------- |",
    ]
    for h in hits[:50]:
        lines.append(f"| `{h.file}` | {h.line_num} | `{h.change_path}` |")

    lines += [
        "",
        "**Action required:** Update these references before the provider PR is merged.",
        "",
        "---",
        f"_Opened by [drift-guard](https://github.com/pgomes13/drift-guard-engine) · {pr_link}_",
    ]
    return "\n".join(lines)


def _upsert_issue(
    client: httpx.Client,
    repo: str,
    body: str,
    provider_repo: str,
    pr_number: int,
):
    title = f"⚠️ Breaking API changes from {provider_repo}" + (f" (PR #{pr_number})" if pr_number else "")

    # Ensure the label exists
    try:
        client.post(
            f"{_GITHUB_API}/repos/{repo}/labels",
            json={"name": _ISSUE_LABEL, "color": "e11d48", "description": "API drift impact"},
        )
    except Exception:
        pass

    try:
        # Check for an existing open issue with our label
        resp = client.get(
            f"{_GITHUB_API}/repos/{repo}/issues",
            params={"labels": _ISSUE_LABEL, "state": "open", "per_page": 1},
        )
        resp.raise_for_status()
        existing = resp.json()

        if existing:
            issue_number = existing[0]["number"]
            client.patch(
                f"{_GITHUB_API}/repos/{repo}/issues/{issue_number}",
                json={"title": title, "body": body},
            ).raise_for_status()
            print(f"[notify] Updated issue #{issue_number} in {repo}")
        else:
            client.post(
                f"{_GITHUB_API}/repos/{repo}/issues",
                json={"title": title, "body": body, "labels": [_ISSUE_LABEL]},
            ).raise_for_status()
            print(f"[notify] Opened issue in {repo}")

    except httpx.HTTPError as e:
        print(f"[notify] Failed to upsert issue in {repo}: {e}")
