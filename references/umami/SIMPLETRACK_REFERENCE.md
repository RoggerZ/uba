# Umami Source Reference Snapshot

This directory is a read-only reference snapshot of the upstream Umami source tree for SimpleTrack implementation study.

Source information:

- Source repository: `https://github.com/umami-software/umami.git`
- Source branch: `master`
- Source commit: `c78ff36db0c82e13c86e5073020472c6546313a3`
- Source commit date: `2026-04-16T18:57:37-04:00`
- Snapshot date: `2026-05-01`
- Upstream license: MIT, see `LICENSE`
- Snapshot intent: implementation reference for tracker collection, event/session models, Realtime and Events read paths, filter/query patterns, and ClickHouse/PostgreSQL storage layout

Rules:

- Treat this directory as read-only reference material, not as active product code.
- Do not build SimpleTrack features directly inside this snapshot.
- Do not refactor, rename, or patch upstream files here as part of SimpleTrack work.
- Do not copy Umami code verbatim into `analytics-core` or `simpletrack-saas`; reimplement concepts through SimpleTrack-owned interfaces and naming.
- If upstream Umami needs to be refreshed, replace the snapshot intentionally and record the new commit here and in the SimpleTrack decision documents.

Included in this snapshot:

- Upstream Next.js app, API routes, tracker and recorder source
- Prisma schema and migrations
- PostgreSQL and ClickHouse database schema assets
- Docker, CI, test, localization, and public assets from the upstream repository

Excluded from this snapshot:

- Upstream Git metadata such as `.git`
- Installed dependencies such as `node_modules`
- Runtime build artifacts generated after cloning

Recommended usage:

- Read `src/app/api/send/route.ts` and `src/tracker/index.js` for collection semantics.
- Read `prisma/schema.prisma` and `db/clickhouse/schema.sql` for event/session storage modeling.
- Read `src/queries/sql/` for Realtime, Events, reports, filter parsing, and backend-specific query patterns.
- Read `src/app/(main)/websites/[websiteId]/` for product UI composition patterns.
- Keep SimpleTrack P1 focused on tracker, collect, Realtime, Events, website setup, and docs/quickstart; treat Umami Boards, Replays, Revenue, Attribution, Links, Pixels, and Teams as later-stage reference only.
