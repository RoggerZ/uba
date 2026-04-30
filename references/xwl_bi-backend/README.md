# xwl_bi Backend Reference Snapshot

This directory is a temporary, read-only reference snapshot copied from the local `xwl_bi` workspace for `analytics-core` implementation study.

Source information:

- Source path: `C:\Users\admin\Documents\src\xwl_bi`
- Source commit: `90636d80def26cf6eb1f53e0cba2c415835b6973`
- Snapshot intent: backend-only implementation reference for ingestion, consumer pipeline, ClickHouse write path, metadata flow, and analysis service design

Rules:

- Treat this directory as read-only reference material, not as active product code.
- Do not build new features directly inside this snapshot.
- Do not refactor or rename files here as part of `analytics-core` work.
- If upstream `xwl_bi` needs to be re-snapshotted, replace the snapshot intentionally and record the source commit again.

Included in this snapshot:

- Go backend source directories such as `application`, `cmd`, `controller`, `engine`, `middleware`, `model`, `platform-basic-libs`, and `router`
- Top-level `go.mod` and `go.sum`
- Selected top-level Markdown documents from `xwl_bi/docs`

Excluded from this snapshot:

- Vue2 admin frontend
- SDK frontend code
- Runtime logs
- Built binaries and local execution artifacts
- Large screenshot-heavy documentation subdirectories
- Source repository metadata such as `.git`, `.omx`, and IDE folders

Recommended usage:

- Read this snapshot when mapping `xwl_bi` concepts into `analytics-core`.
- Copy ideas, interfaces, query patterns, and ingestion flow decisions into `analytics-core` through clean reimplementation.
- Keep business-specific naming, legacy UI coupling, and old product boundaries out of `analytics-core`.
