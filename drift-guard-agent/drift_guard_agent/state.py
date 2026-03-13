"""LangGraph state definition for the drift-guard agent."""

from __future__ import annotations

from dataclasses import dataclass, field
from typing import List, Dict, Literal, Optional, TypedDict


@dataclass
class Change:
    type: str
    severity: str  # "breaking" | "non-breaking" | "info"
    path: str
    method: str
    location: str
    description: str
    before: str = ""
    after: str = ""


@dataclass
class DiffResult:
    base_file: str
    head_file: str
    changes: list[Change]
    summary: dict  # {total, breaking, non_breaking, info}


@dataclass
class RiskScore:
    change_index: int
    risk: Literal["high", "medium", "low"]
    reason: str


@dataclass
class Hit:
    file: str
    line_num: int
    line: str
    change_type: str
    change_path: str


@dataclass
class ConsumerRepo:
    full_name: str       # "org/repo"
    clone_url: str
    local_path: str = ""
    scan_dir: str = "."


class DriftState(TypedDict):
    # Input
    diff: Optional[DiffResult]
    org: str
    token: str
    github_token: str
    pr_number: int
    provider_repo: str   # full name of the provider repo, excluded from search
    model: str
    dry_run: bool

    # Pipeline state
    consumers: List[ConsumerRepo]
    hits: Dict[str, List[Hit]]                   # repo full_name → hits
    explanations: Dict[str, List[str]]            # repo full_name → explanations

    # Outputs
    consumer_issues: Dict[str, str]              # repo full_name → issue body


def initial_state(**kwargs) -> DriftState:
    defaults: DriftState = {
        "diff": None,
        "org": "",
        "token": "",
        "github_token": "",
        "pr_number": 0,
        "provider_repo": "",
        "model": "claude-opus-4-6",
        "dry_run": False,
        "consumers": [],
        "hits": {},
        "explanations": {},
        "consumer_issues": {},
    }
    return {**defaults, **kwargs}
