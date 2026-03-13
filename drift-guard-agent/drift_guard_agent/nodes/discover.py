"""Discover node: find consumer repos via GitHub code search.

Search terms are derived directly from the diff's breaking endpoint paths —
no service URL configuration required.
"""

from __future__ import annotations

import os
import time

import httpx

from drift_guard_agent.state import Change, ConsumerRepo, DriftState

_GITHUB_API = "https://api.github.com"
_SEARCH_LIMIT = 20  # max repos to process


def discover_consumers(state: DriftState) -> dict:
    token = state.get("token", "") or os.environ.get("ORG_READ_TOKEN", "")
    org = state["org"]
    provider_repo = state.get("provider_repo", "")
    diff = state["diff"]

    if not token or not org:
        print("[discover] No token/org — skipping consumer discovery")
        return {"consumers": []}

    breaking = [c for c in diff.changes if c.severity == "breaking"]
    if not breaking:
        return {"consumers": []}

    search_terms = _search_terms_from_diff(breaking)
    print(f"[discover] Searching org '{org}' for: {search_terms}")

    headers = {
        "Authorization": f"Bearer {token}",
        "Accept": "application/vnd.github+json",
        "X-GitHub-Api-Version": "2022-11-28",
    }

    repo_names: set[str] = set()

    with httpx.Client(headers=headers, timeout=30) as client:
        for term in search_terms:
            query = f'org:{org} "{term}"'
            try:
                resp = client.get(
                    f"{_GITHUB_API}/search/code",
                    params={"q": query, "per_page": 30},
                )
                resp.raise_for_status()
                items = resp.json().get("items", [])
                for item in items:
                    full_name = item["repository"]["full_name"]
                    # Exclude the provider repo itself
                    if full_name != provider_repo:
                        repo_names.add(full_name)
                # GitHub code search rate limit: 10 req/min for authenticated
                time.sleep(6)
            except httpx.HTTPStatusError as e:
                if e.response.status_code == 422:
                    print(f"[discover] Search term too short/invalid: {term!r}")
                elif e.response.status_code == 429:
                    print("[discover] Rate limited — waiting 30s")
                    time.sleep(30)
                else:
                    print(f"[discover] GitHub search error {e.response.status_code}: {e}")
            except httpx.HTTPError as e:
                print(f"[discover] GitHub search error: {e}")

            if len(repo_names) >= _SEARCH_LIMIT:
                break

    consumers = [
        ConsumerRepo(
            full_name=name,
            clone_url=f"https://x-access-token:{token}@github.com/{name}.git",
        )
        for name in list(repo_names)[:_SEARCH_LIMIT]
    ]

    print(f"[discover] Found {len(consumers)} potential consumer(s): {[c.full_name for c in consumers]}")
    return {"consumers": consumers}


def _search_terms_from_diff(breaking: list[Change]) -> list[str]:
    """Derive GitHub code search terms from broken endpoint paths.

    /users/{id}  →  "/users"
    /coffees     →  "/coffees"
    """
    terms: set[str] = set()
    for c in breaking:
        parts = [p for p in c.path.split("/") if p and not p.startswith("{")]
        if parts:
            # Use first stable path segment as the search term (no trailing slash
            # so we match both "/coffees" and "/coffees/" in consumer code)
            terms.add("/" + parts[0])
    return sorted(terms)
