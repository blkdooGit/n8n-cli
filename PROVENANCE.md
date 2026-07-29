# Provenance

This repository is a derivative of [edenreich/n8n-cli](https://github.com/edenreich/n8n-cli)
by Eden Reich (MIT license) — a stripped-history local copy pushed to
`blkdooGit/n8n-cli`. It is NOT original BLKdoo work.

- Upstream attribution is intact and must remain so: `LICENSE`
  ("Copyright © 2025 Eden Reich"), the `go.mod` module path
  (`github.com/edenreich/n8n-cli`), and per-file copyright headers.
- Local git history starts at an artificial "Initial commit" (2025-05-11).
  `CHANGELOG.md` was copied wholesale from upstream and references upstream
  releases and compare links (e.g. v0.7.1, 2026-04-25 — newer than this
  copy's last functional commit, 2026-03-10).
- BLKdoo's actual use: workflow JSON sync (`workflows refresh` / `sync`)
  for BOS-v4.1, EVD-social-v3.0 and MeerSentry. BLKdoo-specific gotchas
  and scope live in `openspec/config.yaml` and each consumer project's
  CLAUDE.md.
- **Decision (founder directive, 2026-07-29):** keep as a documented
  derivative — no rebranding, no Go module rename, no claims of original
  authorship. Any change to its public representation requires a new,
  explicit founder decision.

Provenance evidence verified 2026-07-29: LICENSE, go.mod module path,
118 upstream references in CHANGELOG.md, upstream copyright headers in
10 `.go` files.
