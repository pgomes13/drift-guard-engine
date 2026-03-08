---
layout: home

hero:
  name: DriftGuard
  text: API Contract Change Detection
  tagline: Detect and classify breaking vs. non-breaking schema changes across OpenAPI, GraphQL, and gRPC.
  actions:
    - theme: brand
      text: Get Started
      link: /install
    - theme: alt
      text: View on GitHub
      link: https://github.com/pgomes13/drift-guard-engine

features:
  - title: Multi-schema support
    details: Parses OpenAPI 3.x (YAML/JSON), GraphQL SDL, and Protobuf (.proto) schemas.
  - title: Severity classification
    details: Every change is classified as breaking, non-breaking, or info — with detailed rules per schema type.
  - title: CI-ready
    details: Posts PR comments with the full diff, updates a drift log on GitHub Pages, and supports --fail-on-breaking to block merges.
---
