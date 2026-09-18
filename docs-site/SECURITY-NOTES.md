# docs-site — dependency security triage (P5 / F1)

This file records the security triage for the `docs-site/` Starlight
documentation toolchain so the rationale survives in the repository (it was
originally only captured in the PR #48 description, which a squash-merge would
discard). It reflects `npm audit` run in `docs-site/` on 2026-09-18 (Node
v20.20.2, npm 10.8.2) against the pinned lockfile.

## Summary

- Toolchain: `astro@5.18.2`, `@astrojs/starlight@0.37.7`, `sharp@0.35.4`
  (Astro 5.x is required for the Node 20 runtime we build on — see P5 §A1).
- `npm audit fix` (non-breaking) resolves nothing: every remaining fix is
  semver-major (`astro@7.3.3`).
- Current `npm audit`: **4 vulnerabilities (3 low, 1 critical)**.

## Resolved — sharp HIGH advisory (cleared)

- Bumped `sharp` from `^0.34.0` → `^0.35.4`. Because Astro 5.18.2 pins
  `sharp ^0.34.0`, this required an `overrides` entry in `package.json`
  (both `dependencies.sharp` and `overrides.sharp` are `^0.35.4`); the
  installed tree resolves `sharp@0.35.4` (overridden).
- This clears the `sharp`/libvips (libheif) HIGH advisory
  (GHSA-f88m-g3jw-g9cj, GHSA-rgj7-g3m4-5g8c;
  CVE-2026-33327/33328/35590/35591).
- Verified: `npm run build` stays green and both `.webp` diagrams re-process
  correctly under sharp 0.35.4.

## Residual — Astro chain (documented, not shipped blind)

The remaining advisories are all in the Astro dependency chain and are **only**
patched in `astro@7.3.3` — a breaking **5 → 7** major migration that also
requires **Node >= 22**. That migration is out of scope for this phase (we are
on Node 20, which needs Astro 5.x). The latest Astro 5.x is `5.18.2` (already
pinned); no 5.x patch exists for these.

`npm audit` reports 4 vulnerable nodes:

| Package | npm node severity | Notes |
|---|---|---|
| `astro` (<=7.2.7) | **critical** | aggregate of the advisories below |
| `@astrojs/mdx` | low | depends on vulnerable `astro` |
| `@astrojs/starlight` | low | depends on vulnerable `@astrojs/mdx` / `astro` |
| `esbuild` (0.27.3–0.28.0) | low | dev-server, Windows only |

Advisories rolled up under the `astro` node:

- **CRITICAL — `GHSA-26w7-cxv4-gfx2`**: Remote code execution through AVIF
  image optimization.
- **HIGH — `GHSA-2pvr-wf23-7pc7`**: Host-header SSRF in prerendered error-page
  fetch.
- **HIGH — `GHSA-8hv8-536x-4wqp`**: Reflected XSS via unescaped slot name.
- Moderate/low (XSS / replay / authz-boundary variants): `GHSA-j687-52p2-xcff`,
  `GHSA-xr5h-phrj-8vxv`, `GHSA-jrpj-wcv7-9fh9`, `GHSA-f48w-9m4c-m7f5`,
  `GHSA-7pw4-f3q4-r2p2`, `GHSA-4g3v-8h47-v7g6`, `GHSA-376h-93r7-7g6f`.
- `esbuild` — **`GHSA-g7r4-m6w7-qqqr`**: arbitrary file read when running the
  dev server on Windows.

## Why the residual real-world risk is low here

- Every advisory is a **build-time / dev-time** concern only. The site output
  is fully static, pre-rendered HTML served from the GitHub Pages CDN — there
  is **no Astro server runtime, no SSR, and no user-controlled input at request
  time**. The Host-header SSRF and the reflected/stored XSS advisories all
  require a live Astro server that we do not run in production.
- The **AVIF RCE** (`GHSA-26w7-cxv4-gfx2`) requires processing a malicious
  image at build time. This build only optimizes **two first-party, committed
  `.webp` diagrams** — no untrusted images enter the pipeline.
- The **esbuild** advisory is `astro dev` on **Windows only**; CI and our
  builds run `astro build` on Linux.

## Recommended follow-up

- Plan an **Astro 5 → 7 major migration on Node 22** in a dedicated phase to
  clear the CRITICAL AVIF RCE and the rest of the Astro chain. This is a
  breaking upgrade (Node runtime bump + Astro 7 API changes) and should be
  validated against the full site build and the load-bearing `metadata.json` /
  `.nojekyll` invariants before shipping.
