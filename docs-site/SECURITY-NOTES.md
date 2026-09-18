# docs-site — dependency security triage

This file records the security triage for the `docs-site/` Starlight
documentation toolchain so the rationale survives in the repository. It reflects
`npm audit` run in `docs-site/` against the pinned lockfile.

## Current state (2026-09-18, post Astro 5 → 7 migration)

- Build runtime: **Node >= 22** (Astro 7 requires `node >=22.12.0`). CI
  (`.github/workflows/deploy-docs.yml`) and local builds now run on Node 22.
- Toolchain: `astro@7.3.3`, `@astrojs/starlight@0.42.2`, `sharp@0.35.4`.
- `npm audit` (Node v22.23.2, npm 10.9.8): **0 vulnerabilities.**

### Before / after

| | Before (Astro 5.18.2 / Node 20) | After (Astro 7.3.3 / Node 22) |
|---|---|---|
| Total advisories | 4 (3 low, 1 critical) | **0** |
| CRITICAL AVIF RCE (`GHSA-26w7-cxv4-gfx2`) | present | **cleared** |
| astro chain (SSRF / XSS / etc.) | present | **cleared** |
| esbuild dev-server (`GHSA-g7r4-m6w7-qqqr`) | present (low) | **cleared** |
| sharp / libvips HIGH | already mitigated via override | pinned by astro |

## What the migration changed

- **`astro` 5.18.2 → 7.3.3.** This is the version that clears the CRITICAL AVIF
  image-optimization RCE (`GHSA-26w7-cxv4-gfx2`) along with the rest of the
  Astro dependency-chain advisories that were previously only patched in
  `astro@7.x`:
  - CRITICAL — `GHSA-26w7-cxv4-gfx2` (AVIF image-optimization RCE) — cleared.
  - HIGH — `GHSA-2pvr-wf23-7pc7` (Host-header SSRF in prerendered error page) —
    cleared.
  - HIGH — `GHSA-8hv8-536x-4wqp` (reflected XSS via unescaped slot name) —
    cleared.
  - Moderate/low XSS / replay / authz-boundary variants
    (`GHSA-j687-52p2-xcff`, `GHSA-xr5h-phrj-8vxv`, `GHSA-jrpj-wcv7-9fh9`,
    `GHSA-f48w-9m4c-m7f5`, `GHSA-7pw4-f3q4-r2p2`, `GHSA-4g3v-8h47-v7g6`,
    `GHSA-376h-93r7-7g6f`) — cleared.
- **`@astrojs/starlight` 0.37.7 → 0.42.2.** 0.42.2 declares Astro 7 support
  (peer `astro ^7.2.10`); the 0.37.x line only supported Astro 5. This also
  pulls the `esbuild` transitive past the `GHSA-g7r4-m6w7-qqqr`
  (`astro dev` on Windows) advisory range.
- **`sharp` — override removed.** Previously `docs-site/package.json` carried a
  direct `sharp` dependency plus an `overrides.sharp` entry to force
  `sharp@0.35.4` on top of Astro 5.18.2's `sharp ^0.34.0` pin (needed to clear
  the sharp/libvips HIGH advisory `GHSA-f88m-g3jw-g9cj` /
  `GHSA-rgj7-g3m4-5g8c`). Astro 7.3.3 itself pins `sharp ^0.35.4`, so the
  override and the direct dependency are no longer required and were dropped;
  the installed tree still resolves `sharp@0.35.4`.
- **Node runtime 20 → 22.** Required by Astro 7. Bumped in the deploy workflow
  (`node-version: '22'`).

## Config / breaking-change fallout

No source or config changes were required by the Astro 5 → 7 migration:

- `astro.config.mjs` — no API changes needed (site/base, the Starlight
  integration config, and the sidebar are unchanged).
- `src/content.config.ts` — already used the content-layer API
  (`docsLoader()` + `docsSchema()`), which is current in Astro 7; no migration
  needed.
- The load-bearing invariants are intact: the build still emits
  `dist/metadata.json` byte-identical to `public/metadata.json`
  (sha256 `2327230a30112ae13aceae3497c3cc8d0afbe9eb6fa97e11854ebba708bdfe53`,
  313 bytes) and `dist/.nojekyll`, and the deploy workflow's self-enforcing
  sha256 guard step is unchanged.

## Residual

None. `npm audit` reports 0 vulnerabilities after the migration.

## Real-world risk context (unchanged)

The site is fully static, pre-rendered HTML served from the GitHub Pages CDN —
there is no Astro server runtime, no SSR, and no user-controlled input at
request time. Every Astro-chain advisory above was a build-time / dev-time
concern; the AVIF RCE only applied at build time and this build optimizes only
two first-party, committed `.webp` diagrams. Clearing them nonetheless removes
the entire class from the toolchain and keeps `npm audit` clean.
