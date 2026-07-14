---
description: Detailed instructions and examples for developing theme resources in Rill
---

# Instructions for developing a theme in Rill

## Introduction

Themes are resources that define custom color palettes for dashboards in a Rill project. They allow you to customize the visual appearance of explore and canvas dashboards to match your brand or design preferences.

Themes are lightweight resources with no reconciliation cost. When a theme file is saved, Rill validates it but performs no heavy operations. Themes are referenced from `rill.yaml` for project-wide styling or directly from individual explore or canvas dashboards.

## Core Concepts

### Referencing themes

Themes can be applied in two ways:

1. **Project-wide** via `rill.yaml`:
  ```yaml
  # rill.yaml
  explores:
    theme: brand
  canvases:
    theme: brand
  ```

2. **Per-dashboard** in an explore or canvas file:
   ```yaml
   # dashboards/sales.yaml
   type: explore
   metrics_view: sales_metrics
   theme: brand
   ```

### Color formats

Themes support multiple color formats:

- **Hex with `#`** (recommended): `"#FF6A00"`, `"#6366f1"`
- **Hex without `#`**: `FF6A00` (works but less explicit)
- **Named CSS colors**: `blue`, `plum`, `darkgreen`, `seagreen`
- **HSL values**: `hsl(180, 100%, 50%)`, `hsl(236, 34%, 34%)`

For consistency and clarity, we recommend using quoted hex values with the `#` prefix.

## Recommended Theme Structure

The recommended approach uses separate `light:` and `dark:` blocks to define mode-specific colors. This ensures your dashboards look great in both light and dark modes.

```yaml
# themes/brand.yaml
type: theme

# Light mode colors
light:
  # Core brand colors (required)
  primary: "#6366f1"     # Primary actions, emphasis, selected states
  secondary: "#8b5cf6"   # Secondary elements, supporting colors

  # UI surface colors (optional - defaults used if omitted)
  background: "#f8fafc"  # Page background
  surface: "#ffffff"     # Elevated surfaces, panels
  card: "#f1f5f9"        # Card backgrounds

  # KPI delta colors (optional - controls comparison/change value colors)
  # kpi-positive: "#16a34a"  # Green for positive deltas (defaults to gray)
  # kpi-negative: "#dc2626"  # Red for negative deltas (defaults to red)

  # Qualitative palette for categorical data (optional, 24 colors)
  # Used for bar charts, pie charts, legend colors by category
  color-qualitative-1: "#6366f1"   # Indigo
  color-qualitative-2: "#8b5cf6"   # Purple
  color-qualitative-3: "#ec4899"   # Pink
  color-qualitative-4: "#06b6d4"   # Cyan
  color-qualitative-5: "#10b981"   # Emerald
  color-qualitative-6: "#f59e0b"   # Amber
  color-qualitative-7: "#3b82f6"   # Blue
  color-qualitative-8: "#a855f7"   # Violet
  color-qualitative-9: "#ef4444"   # Red
  color-qualitative-10: "#14b8a6"  # Teal
  color-qualitative-11: "#84cc16"  # Lime
  color-qualitative-12: "#f97316"  # Orange
  color-qualitative-13: "#d946ef"  # Fuchsia
  color-qualitative-14: "#eab308"  # Yellow
  color-qualitative-15: "#0ea5e9"  # Sky
  color-qualitative-16: "#a855f7"  # Purple alt
  color-qualitative-17: "#22c55e"  # Green
  color-qualitative-18: "#fb923c"  # Orange alt
  color-qualitative-19: "#f43f5e"  # Rose
  color-qualitative-20: "#6366f1"  # Indigo alt
  color-qualitative-21: "#2dd4bf"  # Teal alt
  color-qualitative-22: "#facc15"  # Yellow alt
  color-qualitative-23: "#c084fc"  # Violet alt
  color-qualitative-24: "#4ade80"  # Green alt

  # Sequential palette for ordered data (optional, 9 colors)
  # Used for heatmaps, choropleth maps, intensity scales
  color-sequential-1: "#eef2ff"   # Lightest
  color-sequential-2: "#e0e7ff"
  color-sequential-3: "#c7d2fe"
  color-sequential-4: "#a5b4fc"
  color-sequential-5: "#818cf8"
  color-sequential-6: "#6366f1"
  color-sequential-7: "#4f46e5"
  color-sequential-8: "#4338ca"
  color-sequential-9: "#3730a3"   # Darkest

  # Diverging palette for data with a meaningful midpoint (optional, 11 colors)
  # Used for showing positive/negative deviation from a baseline
  color-diverging-1: "#dc2626"    # Negative extreme (red)
  color-diverging-2: "#f87171"
  color-diverging-3: "#fca5a5"
  color-diverging-4: "#fecaca"
  color-diverging-5: "#fee2e2"
  color-diverging-6: "#f3f4f6"    # Neutral midpoint
  color-diverging-7: "#dbeafe"
  color-diverging-8: "#93c5fd"
  color-diverging-9: "#60a5fa"
  color-diverging-10: "#3b82f6"
  color-diverging-11: "#2563eb"   # Positive extreme (blue)

# Dark mode colors
dark:
  # Core brand colors (brighter for visibility on dark backgrounds)
  primary: "#818cf8"
  secondary: "#a78bfa"

  # UI surface colors
  background: "#0f172a"  # Deep slate background
  surface: "#1e293b"     # Elevated surfaces
  card: "#334155"        # Card backgrounds

  # KPI delta colors (optional)
  # kpi-positive: "#4ade80"  # Green for positive deltas (brighter for dark mode)
  # kpi-negative: "#f87171"  # Red for negative deltas (brighter for dark mode)

  # Qualitative palette (adjusted for dark mode visibility)
  color-qualitative-1: "#818cf8"
  color-qualitative-2: "#a78bfa"
  color-qualitative-3: "#f472b6"
  color-qualitative-4: "#22d3ee"
  color-qualitative-5: "#34d399"
  color-qualitative-6: "#fbbf24"
  color-qualitative-7: "#60a5fa"
  color-qualitative-8: "#c084fc"
  color-qualitative-9: "#f87171"
  color-qualitative-10: "#2dd4bf"
  color-qualitative-11: "#a3e635"
  color-qualitative-12: "#fb923c"
  color-qualitative-13: "#e879f9"
  color-qualitative-14: "#facc15"
  color-qualitative-15: "#38bdf8"
  color-qualitative-16: "#c084fc"
  color-qualitative-17: "#4ade80"
  color-qualitative-18: "#fdba74"
  color-qualitative-19: "#fb7185"
  color-qualitative-20: "#818cf8"
  color-qualitative-21: "#5eead4"
  color-qualitative-22: "#fde047"
  color-qualitative-23: "#d8b4fe"
  color-qualitative-24: "#86efac"

  # Sequential palette (reversed for dark mode)
  color-sequential-1: "#312e81"   # Darkest
  color-sequential-2: "#3730a3"
  color-sequential-3: "#4338ca"
  color-sequential-4: "#4f46e5"
  color-sequential-5: "#6366f1"
  color-sequential-6: "#818cf8"
  color-sequential-7: "#a5b4fc"
  color-sequential-8: "#c7d2fe"
  color-sequential-9: "#e0e7ff"   # Lightest

  # Diverging palette (adjusted for dark backgrounds)
  color-diverging-1: "#ef4444"
  color-diverging-2: "#f87171"
  color-diverging-3: "#fca5a5"
  color-diverging-4: "#fecaca"
  color-diverging-5: "#fee2e2"
  color-diverging-6: "#475569"    # Neutral slate midpoint
  color-diverging-7: "#bfdbfe"
  color-diverging-8: "#93c5fd"
  color-diverging-9: "#60a5fa"
  color-diverging-10: "#3b82f6"
  color-diverging-11: "#2563eb"
```

## Building an on-brand theme from a brand

A common request is "make a theme that matches `<company>`" or "build a theme that looks like `<url>`". Follow this process to produce a theme that emulates a brand while keeping dashboards legible.

### Step 1: Extract the brand palette

From the brand's website or assets, identify a small structured palette before writing YAML. Prefer the brand's *digital* colors (from the live site / CSS) over print or PDF brand guidelines, since those are the colors users associate with the product on screen.

- **Primary**: the dominant brand color, usually the logo color or primary button color. This drives `primary`.
- **Secondary / accent**: a complementary or analogous color used for links or secondary actions. This drives `secondary`. If the brand is effectively monochrome, derive one by rotating the primary hue ~20–40° or using a desaturated variant.
- **Background**: whether surfaces are pure white, warm off-white, cool gray, etc. Informs `background`, `surface`, and `card`.
- **Text**: the body text color (most brands use near-black, not pure `#000`). Informs `fg-primary`.
- **Status colors**: any brand-specific green/red for success/error. Inform `kpi-positive` / `kpi-negative`.

### Step 2: Build the light theme

- Set `primary` to a saturated, mid-tone color. It powers buttons, selected states, and emphasis, so avoid very light or very dark primaries. If the raw brand color is too light or dark to act as an action color, keep its hue but nudge lightness toward a usable mid-tone.
- Only override `background`, `surface`, and `card` when the brand calls for a distinctly tinted or off-white canvas; otherwise the defaults work well. Keep light backgrounds genuinely light (high lightness, low saturation) — charts and tables assume a light canvas.
- Ensure `fg-primary` meets WCAG AA contrast (≥ 4.5:1) against the background. Near-black brand text almost always passes; verify if the brand text is a mid-gray.

### Step 3: Adapt the dark theme (do not just copy light)

Colors that look correct on white are often too dark, too saturated, or too low-contrast on a dark canvas. Adapt each role:

- **Lighten and slightly desaturate** `primary` and `secondary` so they sit comfortably on a dark surface, while keeping the same hue identity (the dark primary should read as "the same brand color, brighter," not a different hue).
- **Avoid pure black and pure white.** Use a deep near-neutral background and a soft near-white `fg-primary`. Pure `#000`/`#fff` cause harsh contrast and halation.
- **Brighten status colors** so `kpi-positive` / `kpi-negative` pop against the dark canvas.
- **Re-check contrast** against the dark background, not the light one.

## Data visualization palette guidelines

Setting `primary` does not populate the chart palettes. To make charts on-brand, set `color-qualitative-*`, `color-sequential-*`, and `color-diverging-*` explicitly in both `light:` and `dark:`. For these palettes, legibility and analytical correctness come first, brand fidelity second.

### Qualitative palette (categorical data)

Used for dimension values, series, and legend entries. Up to 24 colors; the earliest entries are used most, so they matter most.

- **Lead with the brand**: make `color-qualitative-1` (and ideally `-2`) the brand primary and secondary, so common single- and two-series charts read on-brand.
- **Maximize distinguishability** after the first one or two: walk *around* the hue wheel (e.g. indigo → pink → cyan → amber → green → purple → red → teal) rather than clustering near the brand hue. Adjacent entries should differ clearly in hue and/or lightness.
- **Hold lightness and saturation roughly constant** so no single category pops just because it is brighter.
- **Be colorblind-aware**: do not rely on red/green adjacency to carry meaning; vary lightness alongside hue.
- In dark mode, shift the whole palette lighter so every color stays visible against the dark canvas.
- You need not define all 24 — define as many as the brand supports with clear separation; Rill cycles through what you provide.

### Sequential palette (ordered / quantitative data)

Used for heatmaps, choropleths, and intensity scales. Nine steps.

- Use a **single hue** (ideally the brand hue) with **monotonic, perceptually even lightness** from light to dark.
- **Orientation differs by mode**: in light mode, `-1` is the lightest and `-9` the darkest; in dark mode, reverse it so `-1` is the darkest and `-9` the lightest. In both modes `-1` sits closest to the background and `-9` is the most prominent.

### Diverging palette (data with a meaningful midpoint)

Used for deviation from a baseline, change vs. prior period, or above/below target. Eleven steps.

- Two contrasting hues meeting at a **neutral midpoint** (`-6`): light gray in light mode, dark gray/slate in dark mode.
- **Map semantics intentionally** (conventionally negative on one end, positive on the other) and keep the **two arms symmetric** in intensity so neither side looks more important.
- Prefer hue pairs that survive color-vision deficiency (e.g. blue↔red, or brand-hue↔complement) over red↔green alone; if using red↔green, keep a clear lightness difference between the arms.

## Minimal Theme Example

If you only need to set brand colors without customizing palettes, use this minimal structure:

```yaml
# themes/brand.yaml
type: theme

light:
  primary: "#0369a1"
  secondary: "#06b6d4"

dark:
  primary: "#38bdf8"
  secondary: "#22d3ee"
```

## Legacy Format

Older Rill projects may use a simpler format with a top-level `colors:` block. This format is still supported but deprecated in favor of the `light:`/`dark:` structure:

```yaml
# Legacy format (deprecated)
type: theme
colors:
  primary: "#FF6A00"
  secondary: "#0F46A3"
```

## Reference documentation

Here is a full JSON schema for the theme syntax:

```
{% json_schema_for_resource "theme" %}
```
