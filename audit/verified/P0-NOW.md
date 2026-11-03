# P0 — Must fix now

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (2026-07-31).

Active security, data-loss, corruption, crash, and broad-outage risks. Fixed findings are deliberately excluded.

Count: **16**.

| Finding | Source | Original severity | Status | Current evidence |
|---|---|---:|---|---|
| [admin-auth-rbac-03](./parts/admin-auth-rbac.verified.md) — Magic token replays creator admin attributes | `admin-auth-rbac.md` | MEDIUM | PENDING | admin/server/magic_tokens.go:90-100; admin/server/runtime_jwt.go:113-120 |
| [admin-projects-deploy-01](./parts/admin-projects-deploy.verified.md) — Failed editable checkpoint still destroys work | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/deployments.go:413-476 |
| [admin-projects-deploy-02](./parts/admin-projects-deploy.verified.md) — Public org project exposes private assets | `admin-projects-deploy.md` | MEDIUM | PENDING | admin/server/assets.go:186-194 |
| [PIVO-12](./parts/deep-audit-pivot.verified.md) — Stored XSS via innerHTML | `deep-audit-pivot.md` | MEDIUM | PENDING | RegularTable.svelte:129-231 |
| [TIME-1](./parts/deep-audit-time-dimension-detail.verified.md) — Stored XSS in TDD row header | `deep-audit-time-dimension-detail.md` | HIGH | PENDING | TDDTable.svelte:247-262 |
| [deep-cloud-deploy-13](./parts/deep-cloud-deploy.verified.md) — Editable redeploy deletes active branch | `deep-cloud-deploy.md` | MEDIUM | PENDING | admin/projects.go:336-406; admin/deployments.go:461-473 |
| [deep-connectors-02](./parts/deep-connectors.verified.md) — rduckdb deletes good version after failed metadata upload | `deep-connectors.md` | HIGH | PENDING | runtime/pkg/rduckdb/remote.go:244-255 |
| [deep-connectors-03](./parts/deep-connectors.verified.md) — ClickHouse emits storage credentials in SQL telemetry | `deep-connectors.md` | HIGH | PENDING | runtime/drivers/clickhouse/model_executor_objectstore_self.go:126-163 |
| [deep-connectors-04](./parts/deep-connectors.verified.md) — Salesforce fetch failure becomes empty success | `deep-connectors.md` | HIGH | PENDING | runtime/drivers/salesforce/bulk.go:201-231 |
| [drivers-object-store-02](./parts/drivers-object-store.verified.md) — Path-prefix authorization bypass in glob/download | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/s3/object_store.go:70-101 |
| [drivers-object-store-05](./parts/drivers-object-store.verified.md) — Non-boundary path-prefix authorization | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/object_store.go:177-180,253-287 |
| [drivers-object-store-15](./parts/drivers-object-store.verified.md) — HTTPS FilePaths SSRF | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/https/https.go:259-275 |
| [drivers-object-store-16](./parts/drivers-object-store.verified.md) — Object-store delete over-deletes prefixes | `drivers-object-store.md` | MEDIUM | PENDING | runtime/drivers/s3/model_manager.go:42-43 |
| [controller-delete-error](./REVIEW-FINDINGS-DELTA.md) — Failed delete reconciliation still hard-deletes the catalog resource | `REVIEW_FINDINGS_TODO.md` | HIGH | PENDING | runtime/controller.go:1459-1470 logs the delete error and still calls catalog.delete |
| [embed-singleton-stale-context](./REVIEW-FINDINGS-DELTA.md) — Embed singleton can retain a previous security context | `REVIEW_FINDINGS_TODO.md` | HIGH | PENDING | embed-store.ts:30-53 is process-global; embed/+layout.ts:7-10 initializes only when absent |
| [runtime-pkg-13](./parts/runtime-pkg.verified.md) — GetLogs limit zero panics | `runtime-pkg.md` | HIGH | PENDING | runtime/pkg/logbuffer/buffer.go:73-95 |

