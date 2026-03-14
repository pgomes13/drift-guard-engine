import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Drift Agent',
  description: 'API type safety across OpenAPI, GraphQL, and gRPC. Catch breaking changes before they reach production.',
  base: '/api-drift-engine/',

  themeConfig: {
    nav: [
      { text: 'Guide', link: '/install' },
      { text: 'GitHub', link: 'https://github.com/DriftAgent/api-drift-engine' },
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
        text: 'Integrations',
        items: [
          { text: 'npm', link: '/npm' },
          { text: 'Go', link: '/sdk' },
          { text: 'CI', link: '/ci' },
          { text: 'gRPC', link: '/grpc-server' },
          { text: 'MCP (AI)', link: '/mcp' },
        ],
      },
      {
        text: 'Troubleshooting',
        items: [
          { text: 'Generating Specs', link: '/generating-specs' },
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
      { icon: 'github', link: 'https://github.com/DriftAgent/api-drift-engine' },
    ],

    footer: {
      message: 'Released under the MIT License.',
    },
  },
})
