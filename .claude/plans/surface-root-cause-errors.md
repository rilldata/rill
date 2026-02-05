# Surfacing Root Cause Errors in Dependency Failures

## Problem Statement

When a resource fails due to a dependency error, users see a generic message that doesn't help them debug:

```
404 - Unable to load dashboard preview
dependency error: resource "clickhouse_commits_metrics" (rill.runtime.v1.MetricsView) has an error
```

The user has no way to know _what_ the actual error is. They must manually trace through the DAG to find which upstream resource failed and why.

### How Errors Flow Today

1. **Source fails** (e.g., GCS returns 503)

   - `commits_source` gets error: `"failed to download: 503 Service Unavailable"`

2. **Model sees dependency error**

   - `commits_model` depends on `commits_source`
   - `checkRefs()` in `util.go:30-31` runs:
     ```go
     if res.Meta.ReconcileError != "" {
         return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) has an error", ref.Name, ref.Kind))
     }
     ```
   - The actual error (`"503 Service Unavailable"`) is **discarded**
   - `commits_model` gets: `"dependency error: resource 'commits_source' has an error"`

3. **Metrics view sees dependency error**
   - Same pattern repeats
   - Root cause is completely hidden from user

### Current Frontend Behavior

- Frontend displays `reconcileError` string as-is
- No parsing, no navigation to failing resource
- User must manually click through DAG to find root cause

---

## Options

### Option A: Backend Concatenation (Minimal Change)

**Change:** Modify `runtime/reconcilers/util.go:31` to include the dependency's error:

```go
// Before:
return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) has an error", ref.Name, ref.Kind))

// After:
return runtime.NewDependencyError(fmt.Errorf("resource %q (%s) has an error: %s", ref.Name, ref.Kind, res.Meta.ReconcileError))
```

**Result:**

```
dependency error: resource "commits_model" (Model) has an error:
dependency error: resource "commits_source" (Source) has an error:
failed to download gs://bucket/file.parquet: 503 Service Unavailable
```

| Pros                       | Cons                                           |
| -------------------------- | ---------------------------------------------- |
| One-line fix               | Verbose nested strings                         |
| Works immediately          | Loses structure (can't click to navigate)      |
| No frontend changes needed | Harder to parse programmatically               |
| Root cause visible in logs | May expose sensitive info (connection strings) |

---

### Option B: Structured Proto Field (Moderate Change)

**Change:** Add a field to `ResourceMeta` that points to the failing dependency:

```protobuf
// In proto/rill/runtime/v1/resources.proto
message ResourceMeta {
  // ... existing fields ...
  ResourceName error_source = 18;  // NEW: the dependency that caused this error
}
```

**Backend changes:**

- Modify `checkRefs()` to set `error_source` when returning a dependency error
- Propagate through the reconciler chain

**Frontend changes:**

- Read `error_source` field
- Traverse to root cause resource
- Display "View root cause" button that navigates to failing resource
- Show clickable error chain

**Result:**

- Error message stays simple: `"dependency error: resource 'X' has an error"`
- Frontend can navigate: `X → Y → Z (root cause)`
- Clean separation of concerns

| Pros                                   | Cons                         |
| -------------------------------------- | ---------------------------- |
| Clean architecture                     | Requires proto regeneration  |
| Frontend can show navigation           | More code (~100 lines total) |
| Structured data, not string parsing    | Backend + frontend changes   |
| Future-proof for richer error handling | Slightly more complex        |

---

### Option C: Frontend DAG Traversal (No Backend Change)

**Change:** Frontend parses the error message and traverses the DAG to find root cause.

**Frontend changes:**

- Parse error with regex: `/resource "([^"]+)" \(([^)]+)\)/`
- Look up extracted resource in existing graph data
- Recurse if that resource also has a dependency error
- Show "View root cause" navigation

**Result:**

- Backend unchanged
- Frontend owns the UX entirely

| Pros                                  | Cons                                               |
| ------------------------------------- | -------------------------------------------------- |
| No backend changes needed             | Couples frontend to error message format           |
| Frontend has full control over UX     | Regex parsing is fragile                           |
| All data already available via `refs` | If backend changes message format, frontend breaks |

---

## Questions for Discussion

1. **Verbosity vs. structure:** Is verbose error concatenation acceptable, or do we need structured data?

2. **Ownership:** Should the frontend be responsible for traversing the DAG, or should the backend surface root cause directly?

3. **Security:** Should we filter sensitive info (connection strings, file paths) before exposing in errors? More of a concern for Cloud than Developer.

4. **Logging:** Do we want logs to show the full error chain (helps debugging) or keep them concise?

---

## Related Context

- **Linear issue:** APP-720 (originally ENG-1031)
- **Slack thread:** https://rilldata.slack.com/archives/C02T907FEUB/p1770247184920439
