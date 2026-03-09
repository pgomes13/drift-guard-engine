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
          { text: 'Playground ↗', link: 'https://drift-guard-theta.vercel.app/', target: '_blank' },
          { text: 'Supported', link: '/supported' },
          { text: 'Generating Specs', link: '/generating-specs' },
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
          { text: 'CI / GitHub Actions', link: '/ci' },
          { text: 'Microservices', link: '/microservices' },
          { text: 'npm SDK', link: '/npm' },
          { text: 'Go SDK', link: '/sdk' },
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
