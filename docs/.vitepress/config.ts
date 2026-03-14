import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'DriftGuard',
  description: 'API type safety across OpenAPI, GraphQL, and gRPC. Catch breaking changes before they reach production.',
  base: '/drift-guard-engine/',

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/install' },
      { text: 'GitHub', link: 'https://github.com/pgomes13/drift-guard-engine' },
    ],

    sidebar: [
      {
        text: 'Getting Started',
        items: [
          { text: 'Installation', link: '/install' },
          { text: 'CLI', link: '/cli' },
          { text: 'Usage', link: '/usage' },
          { text: 'Playground ↗', link: 'https://drift-guard-theta.vercel.app/', target: '_blank' },
          { text: 'Supported', link: '/supported' },
          { text: 'Generating Specs', link: '/generating-specs' },
        ],
      },
      {
        text: 'Agentic',
        items: [
          { text: 'API Drift Agent', link: '/api-drift-agent' },
          { text: 'MCP (AI)', link: '/mcp' },
        ],
      },
      {
        text: 'Reference',
        items: [
          { text: 'Output Formats', link: '/output-formats' },
          { text: 'Severity Rules', link: '/severity-rules' },
        ],
      },
      {
        text: 'SDKs & Integrations',
        items: [
          { text: 'GitHub Actions', link: '/ci' },
          { text: 'npm', link: '/npm' },
          { text: 'Go', link: '/sdk' },
          { text: 'gRPC', link: '/grpc-server' },
        ],
      },
      {
        text: 'Contributing',
        items: [
          { text: 'Development', link: '/development' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/pgomes13/drift-guard-engine' },
    ],

    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
