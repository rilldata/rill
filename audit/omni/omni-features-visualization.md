# Omni Analytics — Visualization, Dashboard, Filter, Drill & Consumption Feature Inventory

Source: docs.omni.co (fetched raw via the Mintlify `.md` endpoints, 2026-08).

Note: Omni shipped a **new dashboard editor on May 27, 2026** (containers, pages, advanced filter display, etc.). Older "classic" dashboards must be upgraded (one-way) to use those features.

---

## 1. Visualization types

### Built-in chart types (complete list from `/visualize-present/visualizations/types`)

| Type | Notable options | Weight |
|---|---|---|
| **Bar** | Modes: basic, grouped, stacked, stack %, stacked measures, grouped+stacked measures. Per-series mark override (each series independently bar/line/scatter/area → mixed/combo charts). Stacking set globally or per Y-axis. Totals labels on stacked charts (position, font, bolding, value format). Ordinal axes; missing rows not plotted (fill via "fill missing data"). | Major |
| **Line** | Basic, stack 100, dual axes (left/right Y). Show points toggle, dashed vs solid, thickness, opacity, interpolation (linear, monotone/smoothed, step, step-before, step-after). Temporal (continuous) time axis. Two tooltip styles (all-series vertical hover vs nearest point). | Major |
| **Area** | Stacked, stack 100, overlay. Same line-style options as line charts (points, dash, weight, opacity, interpolation). | Major |
| **Scatterplot / bubble** | Color facet (discrete or continuous with gradient pins), size facet (bubble), per-point data labels mappable to any field, drag-to-configure tooltips. | Major |
| **Pie & donut** | Inner radius slider (pie↔donut), label series mapping + inner/outer radius and layout controls, gradient color-scale pins for numeric coloring. Only two facets: color + size. | Major |
| **Boxplot** | Five-number summary (min, Q1, median, Q3, max). Requires category on Y + unaggregated numerics on X; dates must be cast to string. Minimal specific config. | Minor |
| **Heatmap** | 2 dims + measure grid (commonly two time facets, e.g. week × day-of-week). No heatmap-specific config; uses color (with gradient min/mid/max pins) and tooltip controls. | Major |
| **Funnel** | Orientation (vertical/horizontal), alignment (left/center/right), gap spacing, per-stage colors/palettes/modeled colors, data labels with % change (percent-of-first / percent-of-previous), tooltip percent formatting. | Major |
| **Sankey** | Multi-level flows (node as both source+target), horizontal/vertical orientation, node color discrete/gradient, link color (source-target gradient, source, target, custom), node & link data labels, tooltip percentages (% of total with two denominator modes, % of source/target node). No cycles. | Major |
| **KPI** | Component-based builder: rows of **Number** (dynamic auto-scaling size), **Comparison** (across columns/rows/pivots), **Progress** (bar or circle, progress-to-goal), **Chart** (mini sparkline/bar/line), **Text** (Markdown), **Image** (URL, Mustache-parameterizable). Style tab for alignment/text sizes. One-click convert whole KPI to raw Markdown for full customization. | Major |
| **Map** | **Point maps** (lat/lon; pins colored & sized; heatmap density projection) and **region maps/choropleth** (built-in countries by name/ISO-2/ISO-3, US states by name/2-letter code, custom regions via GeoJSON/TileJSON URL + region property). 2D Mercator & 3D globe projections. Gradient color-scale pins. Saved initial viewport (workbook editor only). | Major |
| **Single record** | One row pivoted into a column view; layouts: horizontal / vertical / compact; independent text formatting (color, size, weight, italic) for field names vs values. | Minor |
| **Table** | See §4 — richest type (conditional formatting, inline bars/images/links/markdown, subtotals via grouped dimensions, freezing, theming). | Major |
| **Markdown tile (visualization)** | Markdown + sanitized HTML/CSS + Mustache query data; loops over result rows and pivots; built-in `<Sparkline>` and `<ChangeArrow>` components; iframes (sandboxed) with data-parameterized URLs; dark-mode class (`prefers-dark`); AI generation via Query Helper. | Major |
| **AI summary** | LLM-generated narrative of the query results as a viz type; steerable with custom context (audience, tone, follow-up questions); requires topic-based query + AI feature flag. | Minor |
| **Custom (Vega-Lite advanced editor)** | Any built-in Vega-Lite-backed chart can be opened as editable Vega-Lite JSON ("View source" / "Open in advanced editor"), or start from scratch. Full spec control; escaped field refs (`view\\.field`). **Drill is unavailable in custom Vega charts.** Also iframe-in-Markdown as an alternative custom-viz path. | Major |

Selection UX: quick-chart accelerator icons for one-click pre-configured variants (stacked vs grouped, mixed bar-line, table/chart/markdown toggle); Omni auto-guesses the best chart per query, with "Reset chart" to re-guess.

### Chart types achieved via the documented Vega-Lite example gallery (not native menu items)
US state/zip choropleths, labeled point maps, radial chart, **waterfall**, tapered & centered funnels, **Gantt/timeline**, isotope (stacked icons), **gauge**, **sunburst**, trellis/small multiples/faceted charts, word cloud, quadrant scatter, **bullet charts**, flag-mark scatter, cross-filtered chart pair.

### Markdown "showcase" recipes (official gallery, built with Markdown viz)
Calendar heatmap, card grid, card list, criteria-checklist card, strip-plot card, target candlesticks card, cohort table, colored progress bars, dashboard tabs, dashboard tile anchors (TOC), data readout, dumbbell plot, gauges, gradient card, honeycomb categorical, honeycomb map, image grid, invoice report, kanban board, KPI conditional colors, KPI table, KPI with sparkline, location + national flag, metric tree, pie/donut custom legend, pills, record lookup, repeating pie/square-fill/waffle charts, simple subtotal table, star rating, status indicator, symmetric funnel, table with tiny bars, team progress tracker, test results cards, thermometer, uptime heatmap, user profile, word cloud.

### Cross-cutting chart configuration
- **Series mapping**: drag any column/pivot to X, Y (left/right), color, size, tooltip facets; fields reusable across facets. (Major)
- **Series configuration**: per-series label & number format (number/percent/currency/integer, decimals), data labels (show-all, position, label with a different field), **trendlines: moving average (center/lagging) and regression lines** (no r²/formula display). (Major)
- **Color/stacking**: org default palettes auto-applied; discrete + continuous (gradient) palettes; per-chart **custom palettes** with bulk hex editor; reverse colors; **gradient scale min/mid/max pinning** (consistent color mapping across charts — supported on scatter, pie, heatmap, funnel, maps, table gradients); per-series color pickers; **model-defined dimension colors** (`colors:` conditions on a dimension → consistent colors across all docs) and measure `colors` param; stacking modes automatic/stack/group/overlay/stack %; **stack segment sorting** (label/value/sum/unsorted); legend position right/left/top/bottom/hidden. (Major)
- **X/Y axis**: title, type override (ordinal/temporal/quantitative), labels (show, axis line, ticks, font+color, format strings, rotation, tick density), **reference lines** (simple targets), grid lines, domain/scale (Y); long-tail config via Vega-Lite spec. (Major)
- **Tooltips**: drag any fields in (need not be plotted), ordering controlled, formatting inherits table; deeper customization via Vega spec. (Minor)
- **Size**: bubble size legend show/hide. (Minor)
- **Dynamic axis labels**: Mustache tags allowed in axis labels (totals, filter values). (Minor)
- **AI viz generation**: Omni Agent creates/refines charts, KPIs, tables, funnels, maps, sankeys, markdown layouts from natural-language prompts ("freeze first column", "highlight values over 1000 in red"). (Major)

---

## 2. Dashboards

### Layout
- **Containers**: two layout types — **Grid** (free-form drag/resize positioning) and **Stack** (row/column flex with gap, padding, wrap, align, justify); containers **nest**; named containers; **lockable** (freeze arrangement recursively); item resize in column stacks (pixel-based, reset-sizing). (Major)
- **Item types**: charts, text (Markdown), filters, controls, page navigation, placeholders, **spacer**, **divider**. (Major)
- **Catalog**: browsable library of all workbook queries/filters/controls for placing/re-adding items. (Minor)
- **Placeholders**: wireframe items (query / filter-control / text typed) with labels+descriptions, visible in published view, one-click replace with real content. (Minor)
- **Pages (tabs)**: up to **15 pages** per dashboard, each with own layout, items, and filters; duplicate/reorder/delete pages; page keys for URL/reference. (Major)
- **Page navigation**: auto-added inline switcher; appearances include tabs, buttons, list, dropdown; custom labels (connected to page title or disconnected with `{{page.KEY.name}}` dynamic refs); multiple tabs → same page; **custom URL tabs** (relative paths, external URLs, mailto, anchors; same-tab/new-tab; optional same-origin query-param forwarding). (Major)
- **Undo/redo** for layout changes (cleared by non-layout edits). (Minor)
- **Mobile**: automatic responsive adaptation with per-container overrides — auto, stack vertically, wrap items, keep layout, hide-on-mobile; mobile preview mode. (Major)
- **Query text items**: live-bound tiles showing a query's title/subtitle/description or **inline/tooltip AI summary**; convertible to Markdown with live Mustache binding (`{{queries.<id>.name}}`) or detachable to static text; hideable. (Minor)
- **Text/Markdown tiles**: markdown + images + links + Mustache dynamic content (filter values, result values). Images = markdown/HTML images in Markdown tiles or KPI image components (no dedicated image tile). (Major)
- **Hidden tiles**: exclude a chart from the dashboard while other queries can still reference it (XLOOKUP, alerts, filter sources); hidden tiles don't run unless invoked. (Minor)
- **Tile chrome**: title, subtitle, description (AI can write subtitle/description); copy tile to another dashboard/workbook; copy link to workbook tab; **Inspect** (opens SQL/cache/execution inspector). (Minor)

### Filters
- **Field-based dashboard filters** (string/numeric/date), auto-mapped to tiles with matching fields. (Major)
- Filter settings: label, hover description, **default values** (including **user-attribute defaults** for per-viewer personalization), apply-to-new-queries, **required filters** (block mapped tiles, or block whole dashboard until set), **hidden filters** (still applied + settable via URL). (Major)
- **Per-tile mapping**: remap to a different same-typed field per tile; disable per tile; map to **table calculations** in a tile's query; calculations can also be the filter source. (Major)
- **Display options** (new editor): string → single/multi **button toggles** (inline or popover), **button group**, multi-select dropdown, exact-value curation (default first 50 values); date → default advanced picker, **single-day calendar**, **timeframe control** (curated presets like "Last 7 days" + optional custom range). (Major)
- **Cross-filtering** (opt-in per dashboard): click any chart element to filter all compatible tiles; **brush selection** on continuous axes for range filters; Cmd/Ctrl+click multi-select; removable chips; source-tile highlight + incompatible-tile dimming; works on Vega charts, pie/donut, maps, tables, sankey; joins resolved through the semantic model; not applied to downloads; SQL tiles need templated filters. (Major)
- **Faceted filters** (opt-in): one filter narrows another's suggestions (state → cities); respects access filters. (Major)
- **Dynamic filter controls**: viewers add their **own ad-hoc filters** on a published dashboard; creator scopes available fields (all fields / by topic / manual list); viewer filters live in URL only. (Major)
- **URL parameters**: filter state serialized in the URL as `f--<filter_id>={"values":[...]}` (shareable/bookmarkable; used for embeds via `filterSearchParam`); apps use `af--` params; browser back/forward steps through filter history (documented for apps). (Major)
- **Filter bar UX**: fixed (sticky) control bar option; collapsible filters (+ "display all initially"). (Minor)
- Workbook-side: "Filters from dashboard" dropdown pulls dashboard filters into workbook queries; extensive date-filter expression language and operator set at the model layer (is/not/contains/between/offsets/day-of-week etc.). (Minor)

### Controls (field-level interactivity, distinct from filters)
- **Field switcher** — swap one field for another (dimension regrouping or measure swap; optional "no field"). (Major)
- **Multi-field picker** — viewer picks zero+ fields to show (optionally controls field order). (Major)
- **Timeframe switcher** — change date grain (month→week→day). (Major)
- **Period-over-period control** — pick current + comparison periods for PoP queries. (Major)
- **Parent control** — one control drives multiple hidden child controls. (Minor)
- **Top N limit control** — viewer changes N on a Top-N field. (Minor)

### Appearance / theming
- **Document themes** (admin-created, reusable; org default theme assignable): background, padding, tile background/shadow/margin/borders (CSS values incl. gradients), title/text fonts (remote woff2, e.g. Google Fonts), control + popover styling, buttons; **default discrete & continuous palettes per theme**; theme JSON import/export; embed themes via URL param. (Major)
- **Dark mode**: instance-level appearance setting; theme defaults adapt; Markdown viz gets `prefers-dark` class. (Minor)
- **Custom state messages**: custom Markdown for tile **Error** and **No results** states. (Minor)
- **Chart palettes admin**: org-level built-in + custom discrete/continuous palettes with visibility control. (Minor)

### Dashboard settings & management
- **Auto-refresh** at chosen interval (cache-bypassing; for TV/wallboard use). (Minor)
- **On-load query scope**: current page only (lazy per-page loading, default) vs all pages. (Minor)
- **Timezone override**: viewer-facing timezone picker (viewer profile TZ / UTC / any named TZ) on dashboard header+footer; also appliable to deliveries and downloads; triple opt-in (connection, org, document). (Major)
- **Preview mode**: desktop + mobile viewer simulation inside the editor, full-screen option. (Minor)
- **View as** (admin): evaluate dashboard under another user's attributes/access grants; **Impersonate** for full-session simulation. (Minor)
- **Performance profiler** for slow-tile diagnosis; query Inspect per tile. (Minor)
- **Document abilities** (per-org defaults + per-document overrides): Download, Schedules, **Drill**, Duplicate, Upload data, Create spreadsheets, Viewers-can-see-workbook, Request access, Dashboard Agent (AI), Timezone override, Require-PR-to-publish. (Major for governance of the consumption surface)

### Download / export / delivery
- **Dashboard download**: **PDF, PNG, XLSX, CSV (ZIP of per-tile CSVs)**; options: paper size (PDF), expand tables to 1,000 rows (PDF/PNG), single-column arrangement, hide title, apply data formatting (XLSX/CSV), exclude hidden columns, all-possible-results (CSV), timezone override, page selection for multi-page. (Major)
- **Per-tile downloads** (dashboard tile menu; also XLSX export of spreadsheet tiles). (Minor)
- **Workbook query download**: CSV, XLSX, JSON; apply formatting; timezone override; row limit current/custom/all (XLSX up to 1M rows). (Major)
- **Scheduled deliveries & alerts** from dashboards or single tiles: destinations **Email** (white-label domain, external recipients, **per-recipient user-attribute personalization**), **Slack**, **Google Sheets**, **Amazon S3**, **SFTP**, **Webhook**; alert conditions (results changed / stayed same / no results / any results) evaluated against a "results to check" query; timezone override; runs as creator identity; fresh-query default with opt-in cached results; deliveries double as cache warming; also available to embed users. (Major)

---

## 3. Drill-downs & linking

- **Measure drill**: click a measure value → drill to underlying records. Default = first 10 columns of the view; curated via workbook UI ("Drill fields" editor) or model `drill_fields` (measure) / `default_drill_fields` (view). (Major)
- **`drill_queries`**: multiple named drill options per measure with full query control (fields, base_view, additive filters, sorts, limit) — a menu of drill paths on click; view-level `default_drill_query` also exists. Overrides drill_fields. (Major)
- **Dimension drill (hierarchical)**: `drill_fields` on dimensions swaps the dimension for child fields and adds clicked values as filters (e.g. country → state), retaining table structure; per-timeframe drill via `timeframe_metadata`. (Major)
- **Drill removal**: `drill_fields: []` or workbook field editor. (Minor)
- **Dashboard table context menu**: right-click a cell → **drill, filter, copy, or expand cell**. (Minor)
- **Drill gating**: "Drill" is a document ability toggle; drill actions emit embed events (`dashboard:tile-drill`) for host apps. Cross-filter + drill both unavailable in custom Vega charts. (Minor)
- **Link drilling / URL links**: model-level `links` parameter on fields → per-row dynamic URLs with labels and icons (Salesforce records, external tools, **cross-linking to other Omni dashboards**, optionally carrying values). Table fields can also render as links (Display-as-Link with a hidden URL calculation) and linked images. Page-navigation custom URL tabs forward filter params to same-origin destinations. (Major)
- **"Drill to dashboard" pattern**: via links parameter + filter URL params, and via multi-page dashboards ("surface drill-downs without leaving the dashboard"). No dedicated point-and-click "drill to dashboard" configurator. (Notable gap vs. some competitors)
- **Explore from anywhere**: any tile/tab → "Explore" opens a scratch copy of the workbook for ad-hoc slicing (governed self-service drill path). (Major)
- **Dashboard Agent** as AI drill: ask follow-ups, drill into subsets, re-aggregate, summarize per tile in a split-view chat, scoped to selected charts, image paste support. (Major)

---

## 4. Table / pivot features

- **Results grid capabilities**: 1,000-row default limit adjustable to 50,000 with pagination; unlimited measures; **pivots limited to 200 columns**; multi-column sort (shift-click); drag-and-drop column reorder; column freeze (freeze-left, saved with viz; pivoted tables auto-freeze first dimension); show/hide columns; pivot sort incl. sort-by-values/totals. (Major)
- **Table options**: column width stretch vs fixed (+ per-column exact px width 50–800), header truncate/wrap, show view names, **group dimensions → expandable/collapsible groups with subtotals** (expand/collapse all), row numbers, banding, per-element font sizes. (Major)
- **Column formatting**: number/currency/percent/date formats, custom label override per column, alignment, word wrap. (Major)
- **Display-as** per field: original value, **image** (sizing, linked images), **link**, **bar** (in-cell data bars with auto or fixed min/max range, negative-value splitting), **Markdown** (model `markdown` param; template + CSS editable in workbook and promotable to model — powers pills, star ratings, status dots, user profiles). (Major)
- **Conditional formatting**: rule-based row/column styling (with transparent backgrounds) + **color-scale gradients** with pinned min/mid/max. (Major)
- **Colors/style**: full background/text styling for header, values, banding, hover, totals; palette copy/paste as JSON. (Minor)
- **Totals**: column totals with per-column **total calculation override** (e.g. average), dimension totals opt-in, calculation-aware totals (formula recomputed on totals row); **row totals** (pivot only); **transpose (swap rows/columns)**; **group measures in pivot** (merged measure header groups). (Major)
- **Spreadsheet-style interaction on dashboards** (read-only): cell/range/row/column selection, keyboard navigation, Cmd/Ctrl+C copy as TSV (headers included on select-all), right-click context menu. (Major)

---

## 5. Workbook / spreadsheet consumption UX

- **Workbook model**: every dashboard is the presentation layer of a workbook; tabs = queries (point-and-click, SQL, or spreadsheet); view vs draft modes with autosave; version history + restore; AI-suggested titles/descriptions; tab duplication, cross-workbook tab import, **linked tabs** (multiple visualizations off one query). (Major)
- **Spreadsheet tabs**: full Excel-style sheets in a workbook fed by **live query-connected protected sheets** that auto-update on query changes; Excel formula library with insert-function builder; import CSV/XLSX preserving formulas/formatting; copy-with-references; manual data entry; export back to XLSX; **on dashboards**: read-only by default, optional multi-sheet tab strip, and **unlockable for viewer cell-editing (what-if analysis on a published dashboard)**. No in-sheet pivot tables or in-sheet charts. (Major)
- **Saved query views**: promote a workbook query to a reusable view (compiled as CTE) for subquery/join workflows. (Minor)
- **Viewer workbook access**: "Viewers can see workbook" ability gives read-only workbook access (topic tabs only). (Minor)
- **Timezone, downloads, deliveries, AI agent**: all available at both workbook and dashboard level (see §2).
- **Apps (beta)**: fully custom HTML/JS interfaces on top of workbook queries (AI-built or manual), sandboxed iframe, Omni design-system CSS vars + org theme, URL-persisted `af--` filters; no scheduling/downloads yet, embed in beta. The "beyond dashboards" consumption surface. (Major, but beta)
- **Embedding**: dashboards/workbooks embeddable with URL-param filters, themes, vanity domains, embed events (filter change, tile download, tile drill, viz rendered), create-mode, embedded deliveries and AI chat. (Major; noted for completeness — deep-dive covered in the AI/embed/governance report.)
