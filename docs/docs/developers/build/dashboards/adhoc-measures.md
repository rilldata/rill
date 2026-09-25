---
title: Adhoc Measures
description: Define arithmetic measures on the fly from a dashboard without editing the metrics view
sidebar_label: Adhoc Measures
sidebar_position: 20
---

**Adhoc measures** let you build a new measure directly from a dashboard by writing an arithmetic expression over the measures your metrics view already exposes — no YAML edit, no reload. Use them to answer a one-off question ("what's revenue per request today?"), to prototype a formula before committing it to the metrics view, or to compose a KPI a canvas component needs without changing the underlying spec.

The feature is available on **explore** and **canvas** dashboards. Where the definition lives is the only real difference:

- On an **explore** dashboard, the definition rides in the URL (the `adhoc_m` query parameter), so a shared link reproduces every adhoc measure the sender sees.
- On a **canvas** dashboard, the definition is stored on the component in YAML (`adhoc_measures:`), so it's part of the persisted dashboard design.

## When to Use One

- You need a derived metric for a single investigation and don't want to touch the metrics view.
- You want to try out an expression (a margin, a rate, a per-thousand ratio) before promoting it into the metrics view.
- You're building a canvas component that needs one composed measure alongside the standard ones on the metrics view.
- You want to hand a colleague a link that already has the derived measure applied.

If a measure is going to be used across many dashboards, add it to the [metrics view](/developers/build/metrics-view/measures) instead — adhoc measures are per-dashboard by design.

## Create an Adhoc Measure on an Explore Dashboard

### 1. Open the measures menu

From any explore dashboard, open the **All Measures** dropdown at the top of the leaderboards. At the bottom of the list, next to the existing measures, you'll find **+ Create adhoc measure**.

<img src="/img/build/dashboard/adhoc-measures/adhoc-measure-menu.png" alt="Create adhoc measure entry in the All Measures dropdown" width="320" />

The same entry appears everywhere else a measure is picked — the big-number tile chooser, the time-series chart's measure picker, and the pivot's **Add field** dropdown.

### 2. Fill in the dialog

Give the new measure a display name, write its expression, and pick a format preset. Everything except the name and expression is optional.

<img src="/img/build/dashboard/adhoc-measures/adhoc-measure-dialog.png" alt="New adhoc measure dialog with display name, expression, description, and format fields" width="520" />

- **Name** — the label you'll see in the UI. A snake-case query alias is derived from it automatically (`Cost per Request` → `cost_per_request`).
- **Expression** — the formula (see [Expression syntax](#expression-syntax) below). Referenced measures must exist on the underlying metrics view.
- **Description** *(optional)* — shown in the measure's tooltip in place of the expression.
- **Format** *(optional)* — `Humanize` (default), `Currency (USD)`, `Currency (EUR)`, or `Percentage`.

Click one of the **Insert a measure** chips to drop a measure name into the expression at the cursor.

### 3. Reference measures with `@`

Inside the expression input, type `@` to open an inline measure picker. Arrow keys navigate; **Enter** inserts the selected measure by its query alias — so you don't have to remember the exact identifier.

<img src="/img/build/dashboard/adhoc-measures/adhoc-measure-mention-picker.png" alt="Inline @ mention picker showing referenceable measures" width="520" />

The picker only lists measures that are valid to reference: window measures, measures with required dimensions, and time-comparison measures are excluded because a wrapped expression can't be filtered the way those need.

### 4. Save and see it on the dashboard

Once the expression parses cleanly and every reference resolves, the **Save** button lights up. Save it, and the measure joins the KPIs on the left-hand column with its own sparkline, participates in the leaderboards, and is available as a pivot column.

![Explore dashboard with a Cost per Request adhoc measure added to the KPI list](/img/build/dashboard/adhoc-measures/adhoc-measure-dashboard.png)

The URL updates with an `adhoc_m` parameter that encodes the definition, so bookmarks and shared links carry it along.

## Create an Adhoc Measure on a Canvas Dashboard

Canvas dashboards expose the same dialog from any component that has a **measure** or **measures** field. Open the component's inspector, click into the field, and pick **+ Create adhoc measure** at the bottom of the dropdown — the same **Name / Expression / Description / Format** form appears.

Unlike explore, canvas persists the definition alongside the component spec:

```yaml
# dashboards/kpis.yaml (excerpt)
type: canvas
rows:
  - items:
      - kpi_grid:
          metrics_view: auction_metrics
          measures:
            - requests
            - cost_per_request
          adhoc_measures:
            - name: cost_per_request
              display_name: Cost per Request
              expression: avg_bid_floor / requests * 1000
              format_preset: currency_usd
```

The definition is scoped to the single component that owns it. Removing the measure from the field selection also cleans it out of `adhoc_measures`, so the YAML never carries entries that aren't referenced.

## Expression Syntax

An adhoc measure expression is a small arithmetic language over your metrics view's measures. It runs server-side through the same parser the metrics API uses.

**Operators**

- Arithmetic: `+`, `-`, `*`, `/`, `%`
- Unary minus, parentheses
- Numeric literals and `NULL`

**Functions** (from the server-side allowlist)

| Function | Arguments |
| --- | --- |
| `abs(x)` | 1 |
| `round(x)` / `round(x, digits)` | 1 or 2 |
| `floor(x)`, `ceil(x)` | 1 |
| `sqrt(x)`, `ln(x)`, `exp(x)` | 1 |
| `power(x, y)` | 2 |
| `coalesce(a, b, ...)` | 2 or more |
| `nullif(a, b)` | 2 |
| `greatest(a, b, ...)` / `least(a, b, ...)` | 2 or more |

**Referencing measures**

- Refer to metrics view measures by their `name` (not display name).
- Type `@` in the expression input to pick one from a menu.
- Adhoc measures cannot reference other adhoc measures — only measures defined in the metrics view.

**Examples**

```text
# Simple ratio
revenue / requests

# Guard against division by zero
revenue / nullif(requests, 0)

# eCPM
avg_bid_floor / requests * 1000

# Rounded margin percent
round(100 * (revenue - cost) / nullif(revenue, 0), 2)
```

## Limits & Validation

The parser and validator enforce a handful of rules so a shared link or saved YAML always resolves cleanly:

- **Up to 10 adhoc measures per dashboard.** The 11th create is blocked with an explicit error.
- **Expressions max out at 1024 characters** and 32 levels of nesting.
- **Name rules**: must start with a letter and contain only letters, digits, and underscores; must not end with a reserved suffix (`_prev`, `_delta`, `_delta_perc`, `_percent_of_total`); must not contain `_rill_`; must not collide with any measure, dimension, time dimension, or other adhoc measure name.
- **Referenceable measures only**: window measures, measures with required dimensions, and time-comparison measures are hidden from the picker and rejected in expressions.
- **SQL reserved words** used as bare measure names must be double-quoted in the expression (for example `"select"`).

Invalid expressions surface an inline error under the expression input while you type, so you never save something that would fail at query time.

## Sharing and Persistence

| Dashboard | Where the definition lives | Sharing model |
| --- | --- | --- |
| Explore | URL parameter `adhoc_m` on the dashboard's state | Copy the URL and the recipient sees the same adhoc measures. |
| Canvas   | `adhoc_measures:` on the component in the canvas YAML | Commit the YAML; the definition is part of the deployed dashboard. |

When you edit an existing adhoc measure, both the URL (for explore) and the component YAML (for canvas) update in place, so a page reload or a re-open of the file picks up the edit without any extra step.

## Related

- [Metrics View Measures](/developers/build/metrics-view/measures) — for measures you want to reuse across many dashboards.
- [Explore Dashboards](/developers/build/dashboards/explore) — where explore-level adhoc measures live and how the URL state is composed.
- [Canvas Dashboards](/developers/build/dashboards/canvas) — the component YAML format that persists canvas adhoc measures.
