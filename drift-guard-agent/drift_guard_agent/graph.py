"""LangGraph graph definition for drift-guard-agent."""

from __future__ import annotations

import os
from typing import Literal

from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, START, StateGraph

from drift_guard_agent.nodes.discover import discover_consumers
from drift_guard_agent.nodes.explain import explain
from drift_guard_agent.nodes.fetch import fetch_consumers
from drift_guard_agent.nodes.ingest import ingest
from drift_guard_agent.nodes.notify import notify
from drift_guard_agent.nodes.scan import scan_consumers
from drift_guard_agent.state import DriftState


def _route_after_ingest(state: DriftState) -> Literal["discover_consumers", "__end__"]:
    diff = state.get("diff")
    if not diff:
        return END
    breaking = [c for c in diff.changes if c.severity == "breaking"]
    if not breaking:
        print("[graph] No breaking changes — done")
        return END
    return "discover_consumers"


def _route_after_discover(state: DriftState) -> Literal["fetch_consumers", "__end__"]:
    if not state.get("consumers"):
        print("[graph] No consumers found — done")
        return END
    return "fetch_consumers"


def _route_after_scan(state: DriftState) -> Literal["explain", "notify", "__end__"]:
    if not state.get("hits"):
        print("[graph] No consumer hits — done")
        return END
    # Use LLM explain only if ANTHROPIC_API_KEY is available
    if os.environ.get("ANTHROPIC_API_KEY"):
        return "explain"
    return "notify"


def build_graph(checkpointer=None) -> StateGraph:
    builder = StateGraph(DriftState)

    builder.add_node("ingest", ingest)
    builder.add_node("discover_consumers", discover_consumers)
    builder.add_node("fetch_consumers", fetch_consumers)
    builder.add_node("scan_consumers", scan_consumers)
    builder.add_node("explain", explain)
    builder.add_node("notify", notify)

    builder.add_edge(START, "ingest")
    builder.add_conditional_edges("ingest", _route_after_ingest)
    builder.add_conditional_edges("discover_consumers", _route_after_discover)
    builder.add_edge("fetch_consumers", "scan_consumers")
    builder.add_conditional_edges("scan_consumers", _route_after_scan)
    builder.add_edge("explain", "notify")
    builder.add_edge("notify", END)

    return builder.compile(checkpointer=checkpointer or MemorySaver())
