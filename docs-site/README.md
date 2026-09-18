# a2acli documentation site

An [Astro](https://astro.build/) + [Starlight](https://starlight.astro.build/)
documentation site for `a2acli`. Self-contained under `docs-site/` — it does not
touch the Go toolchain, `skills/`, `agent-plugin/`, or existing workflows.

## Develop

```sh
cd docs-site
npm install
npm run dev      # local dev server
npm run build    # production build -> dist/
npm run preview  # preview the built site
```

Node 20+ and npm 10+ are required. Use **npm** (a `package-lock.json` is
committed); pnpm/yarn are not used.

## Load-bearing files (do not remove)

GitHub Pages currently serves `/metadata.json` — the OAuth CIMD `client_id`
document that `a2acli auth login` depends on
(`client_id = https://ghchinoy.github.io/a2acli/metadata.json`).

Astro copies everything in `public/` to the site root, so:

- `public/metadata.json` — **must stay byte-identical** to `docs/metadata.json`
  (sha256 `2327230a30112ae13aceae3497c3cc8d0afbe9eb6fa97e11854ebba708bdfe53`).
- `public/.nojekyll` — Pages Jekyll opt-out.

After a build, confirm `dist/metadata.json` exists at the dist root and its
sha256 still matches before considering any deploy.

## Configuration notes

- `base: '/a2acli'` in `astro.config.mjs` — this is a **project** (not user)
  Pages site.
- `site: 'https://ghchinoy.github.io'`.

## Status

Phase 1 (P1) scaffold. The sidebar locks the information architecture; content
pages are placeholders that later phases fill in.
