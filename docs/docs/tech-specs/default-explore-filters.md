# Default Filters for Explore Dashboards

## Overview

Add `defaults.where` to Explore YAML, allowing dashboards to load with pre-configured filters. Uses the same Metrics SQL filter syntax that Canvas already supports for `defaults.filters`.

## YAML Syntax

Standalone explore file:

```yaml
type: explore
metrics_view: sales_metrics
defaults:
  time_range: P30D
  comparison_mode: time
  where: "country IN ('US', 'CA') AND status = 'active'"
```

Inline explore inside a metrics view:

```yaml
type: metrics_view
model: sales_model
timeseries: order_date
dimensions:
  - column: country
  - column: status
measures:
  - name: total_revenue
    expression: SUM(revenue)
explore:
  defaults:
    time_range: P30D
    where: "country = 'US'"
```

Supported operators: `=`, `!=`, `<`, `>`, `<=`, `>=`, `IN`, `NOT IN`, `LIKE`, `NOT LIKE`, `AND`, `OR`, parentheses, string/numeric/null literals.

## Why This Is a Small Win

The proto field `ExplorePreset.where` (field 11) already exists. The frontend already handles `preset.where` in multiple paths:

- `getDefaultExplorePreset` (line 68): `...(explore.defaultPreset ?? {})` spreads `where` from the backend spec.
- `convertPresetToExploreState` (lines 76-86): calls `splitWhereFilter(preset.where)` to populate `whereFilter` and `dimensionThresholdFilters`.
- `cascadingExploreStateMerge`: first-wins merge handles `whereFilter` correctly (URL params > session > bookmark > YAML defaults).

The backend just never populates the field from YAML. Canvas already has a reference implementation using `metricssql.ParseFilter` + `metricsview.ExpressionToProto`.

## Implementation

### 1. Backend: YAML Parsing

**`runtime/parser/parse_explore.go`**

Add `Where` to the `Defaults` struct (line 29):

```go
Defaults *struct {
    Dimensions          *FieldSelectorYAML `yaml:"dimensions"`
    Measures            *FieldSelectorYAML `yaml:"measures"`
    TimeRange           string             `yaml:"time_range"`
    ComparisonMode      string             `yaml:"comparison_mode"`
    ComparisonDimension string             `yaml:"comparison_dimension"`
    Where               string             `yaml:"where"`         // NEW
} `yaml:"defaults"`
```

In the `if tmp.Defaults != nil` block (after line 247, before `defaultPreset` construction):

```go
var whereExpr *runtimev1.Expression
if tmp.Defaults.Where != "" {
    expr, err := metricssql.ParseFilter(tmp.Defaults.Where)
    if err != nil {
        return fmt.Errorf("invalid filter expression in defaults.where: %w", err)
    }
    whereExpr = metricsview.ExpressionToProto(expr)
}
```

Add `Where: whereExpr` to the `ExplorePreset` construction (line 249):

```go
defaultPreset = &runtimev1.ExplorePreset{
    // ... existing fields ...
    Where: whereExpr,
}
```

Add imports: `"github.com/rilldata/rill/runtime/metricsview"` and `"github.com/rilldata/rill/runtime/metricsview/metricssql"`.

**`runtime/parser/parse_metrics_view.go`**

Identical change to the inline explore `Defaults` struct (line 103) and `parseAndInsertInlineExplore` function (line 991).

### 2. Backend: YAML Schema

**`runtime/parser/schema/project.schema.yaml`**

Add `where` to the `defaults` properties under explore definitions:

```yaml
where:
  description: >-
    A Metrics SQL filter expression applied by default when the dashboard loads.
    Example: country IN ('US', 'CA') AND status = 'active'
  type: string
```

### 3. Frontend: YAML Config to State

**`web-common/src/features/dashboards/stores/get-explore-state-from-yaml-config.ts`**

In `getExploreViewStateFromYAMLConfig` (line 189), add handling for `where`:

```typescript
import { splitWhereFilter } from "../filters/measure-filters/measure-filter-utils";

// In getExploreViewStateFromYAMLConfig, after existing defaultPreset handling:
if (defaultPreset.where) {
  const { dimensionFilters, dimensionThresholdFilters } = splitWhereFilter(
    defaultPreset.where,
  );
  exploreViewState.whereFilter = dimensionFilters;
  exploreViewState.dimensionThresholdFilters = dimensionThresholdFilters;
}
if (defaultPreset.dimensionsWithInlistFilter?.length) {
  exploreViewState.dimensionsWithInlistFilter =
    defaultPreset.dimensionsWithInlistFilter;
}
```

Remove the TODO on line 37 (`// TODO: support all fields from V1ExplorePreset`).

### 4. AI Instructions

**`runtime/ai/instructions/data/resources/explore.md`**

Add `where` to the defaults section in the annotated example.

## Files That Need No Changes

| File | Why |
|------|-----|
| `proto/rill/runtime/v1/resources.proto` | `ExplorePreset.where` already exists (field 11) |
| `getDefaultExplorePreset.ts` | Spread on line 68 already propagates `where` |
| `convertPresetToExploreState.ts` | Lines 76-86 already handle `preset.where` |
| `cascadingExploreStateMerge.ts` | First-wins merge already handles `whereFilter` |
| Explore reconciler | Passes through `DefaultPreset` as-is |

## Data Flow

```
YAML: defaults.where: "country = 'US'"
  ↓ metricssql.ParseFilter → metricsview.ExpressionToProto
Proto: ExplorePreset.where = Expression(...)
  ↓ API response
Frontend: explore.defaultPreset.where
  ↓ getDefaultExplorePreset spreads it
  ↓ convertPresetToExploreState → splitWhereFilter
ExploreState: whereFilter + dimensionThresholdFilters
  ↓ cascadingExploreStateMerge (priority: URL > session > bookmark > YAML)
Dashboard renders with filter applied
```

## Filter Priority (Cascading Merge)

Default filters sit at the lowest priority. They only apply when no higher-priority source provides filters:

1. URL params (explicit user navigation)
2. Most-recent session state (localStorage)
3. Bookmark / shared token state
4. **YAML config defaults** (this feature)
5. Rill built-in defaults (empty filter)

Clearing filters and refreshing restores the YAML default — same behavior as `default_time_range`.

## Edge Cases

- **Invalid syntax**: `metricssql.ParseFilter` returns an error at parse time; the resource fails validation with a clear message.
- **Unknown dimension names**: Not validated at parse time (consistent with Canvas). Fails at query time if the dimension doesn't exist.
- **Security row filters**: Independent. Applied at the query layer and combined with (not replaced by) dashboard-level filters.
- **`dimensionsWithInlistFilter`**: Not set from YAML. All IN filters default to "select" mode, which is correct.

## Testing

### Backend

- `runtime/parser/parse_explore_test.go`: Add test cases for valid `where`, invalid `where`, and complex AND/OR expressions. Verify the parsed `ExploreSpec.DefaultPreset.Where` contains the correct `Expression` proto.

### Frontend

- `web-common/src/features/dashboards/stores/`: Test `getExploreStateFromYAMLConfig` with a spec containing `defaultPreset.where`. Verify correct `whereFilter` and `dimensionThresholdFilters` in the output.

### Manual

- Create an explore YAML with `defaults.where` and verify the dashboard loads filtered.
- Verify URL params override the default filter.
- Verify clearing filters + refresh restores the default.
