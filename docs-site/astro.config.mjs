// @ts-check
import { defineConfig } from 'astro/config';
import starlight from '@astrojs/starlight';

// a2acli documentation site.
// Project (not user) GitHub Pages site -> base must be '/a2acli'.
// NOTE: public/metadata.json and public/.nojekyll are load-bearing:
// Astro copies public/ to the site root, preserving the OAuth CIMD
// client_id document served at /metadata.json.
export default defineConfig({
  site: 'https://ghchinoy.github.io',
  base: '/a2acli',
  integrations: [
    starlight({
      customCss: ['./src/styles/custom.css'],
      title: 'a2acli',
      description:
        'a2acli — an A2A-Spec-v1.0-compliant command-line client for talking to A2A agents (gRPC / JSON-RPC / REST, OAuth 2.1).',
      social: [
        {
          icon: 'github',
          label: 'GitHub',
          href: 'https://github.com/ghchinoy/a2acli',
        },
      ],
      sidebar: [
        {
          label: 'Introduction',
          items: [
            { label: 'Overview', slug: 'introduction/overview' },
            { label: 'What is A2A?', slug: 'introduction/what-is-a2a' },
          ],
        },
        {
          label: 'Getting started',
          items: [
            { label: 'Installation', slug: 'getting-started/installation' },
            {
              label: 'Quickstart (5 minutes)',
              slug: 'getting-started/quickstart',
            },
          ],
        },
        {
          label: 'Capabilities & command reference',
          items: [
            { label: 'Command model', slug: 'reference/command-model' },
            {
              label: 'Discovery & Identity',
              slug: 'reference/discovery-identity',
            },
            {
              label: 'Messaging & Tasks',
              slug: 'reference/messaging-tasks',
            },
            { label: 'Serving & Mocking', slug: 'reference/serving-mocking' },
            {
              label: 'Configuration & Auth',
              slug: 'reference/configuration-auth',
            },
            {
              label: 'Conformance & Validation',
              slug: 'reference/conformance-validation',
            },
            {
              label: 'Global flags reference',
              slug: 'reference/global-flags',
            },
          ],
        },
        {
          label: 'Direct-usage tutorial',
          items: [{ label: 'Tutorial', slug: 'tutorial' }],
        },
        {
          label: 'Agent usage',
          items: [
            { label: 'Overview', slug: 'agent-usage/overview' },
            {
              label: 'Install & register',
              slug: 'agent-usage/install-register',
            },
            {
              label: 'Skills reference',
              items: [
                { label: 'Overview', slug: 'agent-usage/skills-reference' },
                {
                  label: 'a2acli',
                  slug: 'agent-usage/skills-reference/a2acli',
                },
                {
                  label: 'a2a-expose',
                  slug: 'agent-usage/skills-reference/a2a-expose',
                },
                {
                  label: 'a2a-conformance',
                  slug: 'agent-usage/skills-reference/a2a-conformance',
                },
              ],
            },
            { label: 'Worked example', slug: 'agent-usage/worked-example' },
          ],
        },
        {
          label: 'Reference & external docs',
          items: [
            {
              label: 'Repo docs & further reading',
              slug: 'external/repo-docs',
            },
          ],
        },
      ],
    }),
  ],
});
