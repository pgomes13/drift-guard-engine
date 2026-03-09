import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'DriftGuard',
  description: 'Detect and classify breaking vs. non-breaking API contract changes across OpenAPI, GraphQL, and gRPC.',
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
        text: 'Playground',
        items: [
          { text: 'Try it online', link: '/playground' },
        ],
      },
      {
        text: 'Integrations',
        items: [
          { text: 'CI / GitHub Actions', link: '/ci' },
          { text: 'gRPC Server', link: '/grpc-server' },
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
