"""CLI entry point for drift-guard-agent."""

from __future__ import annotations

import os
import sys
import uuid

import click

from drift_guard_agent.graph import build_graph
from drift_guard_agent.nodes.ingest import parse_diff_json
from drift_guard_agent.state import initial_state


@click.command()
@click.option("--diff", "diff_path", default="-", show_default=True,
              help="Path to drift-guard JSON diff file, or '-' to read from stdin.")
@click.option("--org", envvar="GITHUB_ORG",
              default=lambda: os.environ.get("GITHUB_REPOSITORY_OWNER", ""),
              help="GitHub org to search for consumer repos. Defaults to GITHUB_REPOSITORY_OWNER.")
@click.option("--token", default="", envvar="ORG_READ_TOKEN",
              help="GitHub PAT with repo:read + read:org for consumer discovery and checkout.")
@click.option("--github-token", default="", envvar="GITHUB_TOKEN",
              help="GitHub token for posting PR comments and opening Issues.")
@click.option("--pr", "pr_number", default=0, type=int, envvar="PR_NUMBER",
              help="Pull request number to link in consumer Issues.")
@click.option("--provider-repo", default="", envvar="GITHUB_REPOSITORY",
              help="Full name of the provider repo (e.g. org/repo) — excluded from consumer search.")
@click.option("--model", default="claude-opus-4-6", show_default=True, envvar="DRIFT_GUARD_MODEL",
              help="Anthropic model to use for analysis (requires ANTHROPIC_API_KEY).")
@click.option("--dry-run", is_flag=True, default=False,
              help="Print output without posting to GitHub.")
def main(
    diff_path: str,
    org: str,
    token: str,
    github_token: str,
    pr_number: int,
    provider_repo: str,
    model: str,
    dry_run: bool,
):
    """Scan consumer repos for impact when a provider PR has breaking API changes.

    Triggered by drift-guard detecting breaking changes. Searches the org for
    repos referencing the broken endpoints, scans them, and opens GitHub Issues
    in any that are affected.
    """
    if not org:
        click.echo("Error: --org is required (or set GITHUB_REPOSITORY_OWNER)", err=True)
        sys.exit(1)

    # Read diff JSON
    if diff_path == "-":
        raw = sys.stdin.read()
    else:
        with open(diff_path) as f:
            raw = f.read()

    diff = parse_diff_json(raw)

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    if not breaking:
        click.echo("No breaking changes — nothing to do.")
        return

    click.echo(f"[drift-guard-agent] {len(breaking)} breaking change(s) detected. Scanning org '{org}' for impacted consumers...")

    # Fall back to GITHUB_TOKEN for consumer operations if no dedicated PAT
    if not token:
        token = github_token or os.environ.get("GITHUB_TOKEN", "")

    graph = build_graph()
    config = {"configurable": {"thread_id": str(uuid.uuid4())}}

    state = initial_state(
        diff=diff,
        org=org,
        token=token,
        github_token=github_token or os.environ.get("GITHUB_TOKEN", ""),
        pr_number=pr_number,
        provider_repo=provider_repo,
        model=model,
        dry_run=dry_run,
    )

    graph.invoke(state, config=config)
    click.echo("Done.")


if __name__ == "__main__":
    main()
