# Verified: drivers-object-store.md

Snapshot: `e2e42fb963508b9b9a75e146632e6192805f957d` (`origin/main`, 2026-07-31).

Exhaustive source finding count: **16** numbered findings. Every numbered finding appears exactly once below and once in the machine-readable ledger.

## Pending

| ID | Original severity | Current priority | Concise title | Current code evidence | Recommended fix |
|---|---:|---:|---|---|---|
| drivers-object-store-01 | HIGH | P1 | Redshift UNLOAD breaks quoted SQL | `runtime/drivers/redshift/warehouse.go:120`; source SQL remains interpolated inside a single-quoted UNLOAD literal | Quote-double/escape the SQL and add a literal-string regression test. |
| drivers-object-store-02 | MEDIUM | P0 | Glob/download bypasses `path_prefixes` authorization | `runtime/drivers/s3/object_store.go:70-101`, with equivalent GCS/Azure paths; only non-glob list receives allowed prefixes | Centralize prefix authorization for list, glob, and download using boundary-aware matching. |
| drivers-object-store-05 | MEDIUM | P0 | Prefix authorization uses raw string prefixes | `runtime/drivers/object_store.go:177-180,253-287`; `strings.HasPrefix` lets `data` authorize `database` | Normalize and compare exact path components/delimiter boundaries. |
| drivers-object-store-07 | MEDIUM | P2 | Redshift page fetch drops caller context | `runtime/drivers/redshift/olap.go:242`; subsequent page uses `context.Background()` | Retain the query context in rows and use it for every page. |
| drivers-object-store-08 | MEDIUM | P2 | Iterator close error skips S3 cleanup | `runtime/drivers/redshift/warehouse.go:238-245` and `runtime/drivers/athena/warehouse.go:253-260` | Always run cleanup and return `errors.Join` of close/cleanup failures. |
| drivers-object-store-09 | MEDIUM | P2 | HTTPS connector ignores connector headers/prefixes | `runtime/drivers/https/https.go:105-115,238-283`; decoded config is discarded and FilePaths only reads model headers | Persist decoded config, merge connector headers, and explicitly enforce prefix policy. |
| drivers-object-store-11 | MEDIUM | P2 | Blob download buffers complete glob listing | `runtime/pkg/blob/download.go:76-89`; `pagination.CollectAll` completes before downloading | Stream paginated entries through a bounded channel or enforce an explicit cap. |
| drivers-object-store-12 | LOW | P2 | Salesforce empty batch list can panic | `runtime/drivers/salesforce/bulk.go:160-178`; non-nil empty slice is indexed before error/length checks | Handle the error first, then require `len(allBatches) > 0`. |
| drivers-object-store-13 | LOW | P3 | Athena resolves output location twice | `runtime/drivers/athena/warehouse.go:52,125`; both paths may call GetWorkGroup | Resolve once and pass the location to `unload`. |
| drivers-object-store-14 | LOW | P2 | Azure connection-string getter mutates shared config | `runtime/drivers/azure/azure.go:108-113`; getter writes the receiver used by concurrent client construction | Return a local resolved value without mutating shared state. |
| drivers-object-store-15 | MEDIUM | P0 | HTTPS FilePaths permits SSRF | `runtime/drivers/https/https.go:259-275`; arbitrary URL reaches `http.DefaultClient` without scheme/address restrictions | Enforce HTTP(S), resolve all redirects/addresses, reject loopback/link-local/private ranges unless explicitly allowed. |
| drivers-object-store-16 | MEDIUM | P0 | Object-store Delete can over-delete sibling prefixes | `runtime/drivers/s3/model_manager.go:42-43`; unbounded literal prefix is sent to recursive deletion | Reject root/empty prefixes and require delimiter-bounded model prefixes before deletion. |

## Fixed / Remove

| ID | Disposition | Evidence | Merged PR / commit |
|---|---|---|---|
| drivers-object-store-03 | FIXED — remove | `runtime/drivers/bigquery/warehouse.go:212-215` closes `f.client` from iterator `Close` | PR #9706, commit `6e1615722e60ea902260bac6afad586a23c66e93` |
| drivers-object-store-06 | FIXED — remove | `runtime/pkg/fileutil/fileutil.go:317-340` cleans both paths and enforces equality or `root + separator`; regression coverage includes sibling-prefix escape | PR #9726, commit `ec882259df62e6d1d9a08987ac37945e8a2a3a89` |

## Partial / Needs test

| ID | Original severity | Current priority | Disposition | Evidence / next action |
|---|---:|---:|---|---|
| drivers-object-store-04 | MEDIUM | P1 | NEEDS MALICIOUS-KEY TEST | `runtime/pkg/blob/download.go:230+` still joins an object key under a temp directory without confinement. Add a bucket fixture containing `../escape` and verify the current provider preserves such keys, then reject all escapes. |
| drivers-object-store-10 | MEDIUM | P2 | NEEDS REDSHIFT FIXTURE | `runtime/drivers/redshift/olap.go:390-401` still parses fixed no-fraction timestamp layouts. Capture an actual Data API fractional response and add flexible layouts. |

Counts: **12 pending, 2 fixed/remove, 2 partial/needs-test, 0 invalid = 16 total**.

## Machine-readable ledger

```text
PENDING|drivers-object-store.md|drivers-object-store-01|HIGH|P1|-|Redshift UNLOAD quoted SQL breakage|runtime/drivers/redshift/warehouse.go:120
PENDING|drivers-object-store.md|drivers-object-store-02|MEDIUM|P0|-|Path-prefix authorization bypass in glob/download|runtime/drivers/s3/object_store.go:70-101
FIXED|drivers-object-store.md|drivers-object-store-03|HIGH|-|-|BigQuery client leak|runtime/drivers/bigquery/warehouse.go:212-215; PR #9706 6e1615722e60ea902260bac6afad586a23c66e93
NEEDS_TEST|drivers-object-store.md|drivers-object-store-04|MEDIUM|P1|-|Object-key zip slip|runtime/pkg/blob/download.go:230+
PENDING|drivers-object-store.md|drivers-object-store-05|MEDIUM|P0|-|Non-boundary path-prefix authorization|runtime/drivers/object_store.go:177-180,253-287
FIXED|drivers-object-store.md|drivers-object-store-06|MEDIUM|-|-|ResolveLocalPath sibling-prefix escape|runtime/pkg/fileutil/fileutil.go:317-340; PR #9726 ec882259df62e6d1d9a08987ac37945e8a2a3a89
PENDING|drivers-object-store.md|drivers-object-store-07|MEDIUM|P2|-|Redshift pagination loses context|runtime/drivers/redshift/olap.go:242
PENDING|drivers-object-store.md|drivers-object-store-08|MEDIUM|P2|-|Iterator close skips S3 cleanup|runtime/drivers/redshift/warehouse.go:238-245; runtime/drivers/athena/warehouse.go:253-260
PENDING|drivers-object-store.md|drivers-object-store-09|MEDIUM|P2|-|HTTPS connector config ignored|runtime/drivers/https/https.go:105-115,238-283
NEEDS_TEST|drivers-object-store.md|drivers-object-store-10|MEDIUM|P2|-|Redshift fractional timestamps rejected|runtime/drivers/redshift/olap.go:390-401
PENDING|drivers-object-store.md|drivers-object-store-11|MEDIUM|P2|-|Blob listing fully materialized|runtime/pkg/blob/download.go:76-89
PENDING|drivers-object-store.md|drivers-object-store-12|LOW|P2|-|Salesforce empty batch panic|runtime/drivers/salesforce/bulk.go:160-178
PENDING|drivers-object-store.md|drivers-object-store-13|LOW|P3|-|Athena output resolved twice|runtime/drivers/athena/warehouse.go:52,125
PENDING|drivers-object-store.md|drivers-object-store-14|LOW|P2|-|Azure getter mutates config|runtime/drivers/azure/azure.go:108-113
PENDING|drivers-object-store.md|drivers-object-store-15|MEDIUM|P0|-|HTTPS FilePaths SSRF|runtime/drivers/https/https.go:259-275
PENDING|drivers-object-store.md|drivers-object-store-16|MEDIUM|P0|-|Object-store delete over-deletes prefixes|runtime/drivers/s3/model_manager.go:42-43
```
