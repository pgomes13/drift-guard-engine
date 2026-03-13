---
layout: home

hero:
  name: DriftGuard
  text: Agentic API Safety
  tagline: AI agent that detects breaking API changes in provider PRs and automatically notifies every affected consumer — zero config required.
  actions:
    - theme: brand
      text: API Drift Agent
      link: /api-drift-agent
    - theme: alt
      text: GitHub Marketplace
      link: https://github.com/marketplace/actions/api-drift-agent
    - theme: alt
      text: View on GitHub
      link: https://github.com/pgomes13/drift-guard-engine

features:
  - title: Agentic workflow
    details: LangGraph-powered agent autonomously discovers consumer repos, scans affected files, and opens GitHub Issues — no manual steps.
  - title: Multi-schema support
    details: Parses OpenAPI 3.x (YAML/JSON), GraphQL SDL, and Protobuf (.proto) schemas.
  - title: Severity classification
    details: Every change is classified as breaking, non-breaking, or info — with detailed rules per schema type.
  - title: MCP tools for AI assistants
    details: Expose schema diffing as native tools to Claude Desktop and other MCP-compatible AI assistants.
---
