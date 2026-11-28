import {
  getRillTheme,
  type RillThemeOptions,
} from "@rilldata/web-common/components/vega/vega-config";
import {
  chartStyleOfVegaConfig,
  gridDash,
  prune,
} from "@rilldata/web-common/features/themes/chart-theme";
import { themeManager } from "@rilldata/web-common/features/themes/theme-manager";
import type {
  V1ThemeChartStyle,
  V1ThemeChartTextRole,
} from "@rilldata/web-common/runtime-client";
import chroma from "chroma-js";
import merge from "deepmerge";
import type { ThemeSpec } from "flint-chart/core";

/**
 * The project theme as a Flint theme (`theme_spec`). It takes what `getRillTheme` takes, and
 * states the same style.
 *
 * Rill styles a chart through the Vega-Lite config it hands the renderer, but Flint decides part
 * of a chart's styling while it compiles: it lays the chart out around its text and legends, and
 * writes mark and scale properties into the spec, where a config cannot reach them. Handing Flint
 * the same style the config holds keeps the two in step.
 *
 * A project that sets no chart style gets no Flint theme, and Flint compiles as it does without
 * one. A style that extends one of Flint's presets carries only what the project theme states,
 * since the preset is a design of its own; any other style is completed with Rill's defaults, so
 * that Flint does not fill the gaps with a look of its own.
 */
export function flintTheme(
  isDarkMode: boolean,
  theme: Record<string, string> | undefined,
  options: RillThemeOptions,
): ThemeSpec | undefined {
  const { chart } = options;
  if (!chart) return undefined;

  if (chart.extends) {
    return prune(
      merge(
        { extends: chart.extends, ...styleTheme(chart) },
        { ink: { series: themePalettes(theme) } },
      ),
    ) as ThemeSpec;
  }

  const config = getRillTheme(isDarkMode, theme, options);
  const { category, heatmap, diverging } = config.range as Record<
    string,
    string[]
  >;
  const surface =
    theme?.["surface-card"] ??
    themeManager.resolveCSSVariable(
      "var(--surface-card)",
      isDarkMode,
      isDarkMode ? DARK_SURFACE : LIGHT_SURFACE,
    );
  return prune(
    merge(styleTheme(chartStyleOfVegaConfig(config)), {
      ink: {
        // Flint tones its inks against the surface the chart sits on. The chart itself stays
        // transparent: `compileFlintSpec` drops the background Flint paints.
        surface: {
          canvas: hex(surface) ?? (isDarkMode ? DARK_SURFACE : LIGHT_SURFACE),
        },
        series: {
          single: hex((config.bar as { fill?: string }).fill),
          byMark: {
            line: { single: hex((config.line as { stroke?: string }).stroke) },
          },
          ...palettes({
            categorical: category,
            sequential: heatmap,
            diverging,
          }),
        },
      },
    }),
  ) as ThemeSpec;
}

/** The card a canvas chart sits on, where neither the theme nor the page says. */
const LIGHT_SURFACE = "#ffffff";
const DARK_SURFACE = "#1b1c21";

/** A chart style in Flint's terms, holding only what the style sets. */
function styleTheme(style: V1ThemeChartStyle): Record<string, unknown> {
  const { text, grid, axis, line, area, bar, point, legend } = style;

  // Rill's axis settings are for both axes; Flint's are by the role an axis plays.
  const axisRole = {
    line: presence(axis?.line),
    lineWeight: axis?.lineWidth,
    ticks: presence(axis?.ticks),
  };
  const dash = gridDash(grid);

  return prune({
    ink: {
      structure: { grid: hex(grid?.color), axis: hex(axis?.color) },
    },
    type: {
      // The other roles are set in the axis labels' family unless they name their own.
      axisLabel: { ...textRole(text?.axisLabel), family: style.fontFamily },
      axisTitle: textRole(text?.axisTitle),
      keyLabel: textRole(text?.legendLabel),
      keyTitle: textRole(text?.legendTitle),
      headline: textRole(text?.title),
      valueLabel: textRole(text?.valueLabel),
    },
    structure: {
      grid: {
        measure: { presence: presence(grid?.measure) },
        category: { presence: presence(grid?.category) },
        // Flint reads an empty pattern as its default dashes, so a solid line is named.
        ...(dash?.length ? { dash } : dash ? { style: "solid" } : {}),
        weight: grid?.width,
      },
      axis: { categorical: axisRole, measure: axisRole },
    },
    marks: {
      strokeWeight: line?.width,
      fillOpacity: area?.fillOpacity,
      areaEdge:
        typeof area?.lineWidth === "number"
          ? {
              presence: presence(area.lineWidth > 0),
              width: area.lineWidth,
            }
          : undefined,
      barOpacity: bar?.opacity,
      cornerRadius: bar?.cornerRadius,
      point: { size: point?.size, opacity: point?.opacity },
    },
    legend: {
      placement: legend?.position ? [legend.position] : undefined,
      swatch: legend?.symbol,
    },
  });
}

/**
 * Flint draws a part at a "presence": how far its ink stands out from the surface. A part the
 * theme turns on is drawn in exactly the ink the theme gives it.
 */
function presence(drawn: boolean | undefined): "full" | "omit" | undefined {
  if (drawn === undefined || drawn === null) return undefined;
  return drawn ? "full" : "omit";
}

function textRole(
  role: V1ThemeChartTextRole | undefined,
): Record<string, unknown> {
  return { size: role?.size, weight: role?.weight, color: hex(role?.color) };
}

/** The palettes a project theme sets itself, which go over a preset's own. */
function themePalettes(theme: Record<string, string> | undefined) {
  const colors = (name: string) =>
    Object.entries(theme ?? {})
      .filter(([variable]) => variable.startsWith(`color-${name}-`))
      .sort(([a], [b]) => a.localeCompare(b, undefined, { numeric: true }))
      .map(([, color]) => color);
  return palettes({
    categorical: colors("qualitative"),
    sequential: colors("sequential"),
    diverging: colors("diverging"),
  });
}

/** Rill's three palettes as Flint's series inks; an empty palette is left to Flint. */
function palettes(colors: {
  categorical: string[];
  sequential: string[];
  diverging: string[];
}) {
  const inks = (palette: string[]) => {
    const parsed = palette.map(hex).filter((color) => color !== undefined);
    return parsed.length ? parsed : undefined;
  };
  // A ramp orders values by color, which takes two colors that differ.
  const ramp = (palette: string[]) => {
    const stops = inks(palette);
    return stops && new Set(stops).size > 1 ? { stops } : undefined;
  };
  return {
    categorical: inks(colors.categorical),
    sequential: ramp(colors.sequential),
    diverging: ramp(colors.diverging),
  };
}

/** A CSS color as the hex Flint reads; undefined for none, or for one that cannot be parsed. */
function hex(color: string | undefined): string | undefined {
  if (!color || !chroma.valid(color)) return undefined;
  return chroma(color).hex();
}
