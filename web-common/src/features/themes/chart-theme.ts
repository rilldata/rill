import type {
  V1ThemeChartGrid,
  V1ThemeChartStyle,
  V1ThemeChartTextRole,
  V1ThemeSpec,
} from "@rilldata/web-common/runtime-client";
import type { Config } from "vega-lite";

/**
 * A theme's `chart` section is the one place a project styles its charts. It reaches a chart
 * through the Vega-Lite config Rill builds (`getRillTheme`), which covers built-in charts,
 * `vega_spec` charts and the Vega-Lite a custom chart compiles to. The mapping lives here so the
 * renderers stay in step.
 */

/** The chart style a theme sets for a mode; the parser has merged the shared settings into it. */
export function resolveChartTheme(
  spec: V1ThemeSpec | undefined,
  isDarkMode: boolean,
): V1ThemeChartStyle | undefined {
  return isDarkMode ? spec?.chart?.dark : spec?.chart?.light;
}

/** Dash patterns of the named grid styles, in pixels. They match Flint's, so custom charts draw the same lines. */
const GRID_STYLE_DASH: Record<string, number[]> = {
  solid: [],
  dashed: [3, 3],
  dotted: [1, 3],
};

/** A grid's dash pattern: its own, else its style's; undefined when the theme sets neither. */
export function gridDash(
  grid: V1ThemeChartGrid | undefined,
): number[] | undefined {
  if (grid?.dash?.length) return grid.dash;
  return grid?.style ? GRID_STYLE_DASH[grid.style] : undefined;
}

/** Length of the tick marks a theme turns on, in pixels. Rill's own axes draw none. */
const TICK_SIZE = 4;

/** Vega's names for the legend symbols a theme can ask for. */
const LEGEND_SYMBOL_TYPE: Record<string, string> = {
  circle: "circle",
  square: "square",
  line: "stroke",
};

/**
 * The Vega-Lite config a chart style asks for, holding only what the style sets. `getRillTheme`
 * merges it over Rill's defaults; a spec's own `config` still takes precedence over both.
 */
export function chartThemeVegaConfig(style: V1ThemeChartStyle): Config {
  const { text, grid, axis, line, area, bar, point, legend } = style;

  // Rill's defaults style the axes through `axisX` and `axisY`, which take precedence over `axis`.
  const axisConfig = {
    ...guideText("label", text?.axisLabel),
    ...guideText("title", text?.axisTitle),
    gridDash: gridDash(grid),
    gridWidth: grid?.width,
    gridColor: grid?.color,
    domain: axis?.line,
    domainWidth: axis?.lineWidth,
    domainColor: axis?.color,
    tickColor: axis?.color,
    ticks: axis?.ticks,
    tickSize: axis?.ticks ? TICK_SIZE : undefined,
  };
  const pointConfig = { size: point?.size, opacity: point?.opacity };

  return prune({
    axisX: axisConfig,
    axisY: axisConfig,
    // Which gridlines are drawn follows what an axis holds rather than where it sits, so a
    // horizontal bar chart keeps the lines its values are read off. The configs by scale type
    // take precedence over `axisX` and `axisY`, where the defaults turn the x grid off.
    axisQuantitative: { grid: grid?.measure },
    axisDiscrete: { grid: grid?.category },
    axisTemporal: { grid: grid?.category },
    axisXTemporal: { grid: grid?.category },
    header: {
      ...guideText("label", text?.axisLabel),
      ...guideText("title", text?.axisTitle),
    },
    legend: {
      ...guideText("label", text?.legendLabel),
      ...guideText("title", text?.legendTitle),
      orient: legend?.position,
      symbolType: legend?.symbol
        ? LEGEND_SYMBOL_TYPE[legend.symbol]
        : undefined,
    },
    title: {
      fontSize: text?.title?.size,
      fontWeight: text?.title?.weight,
      color: text?.title?.color,
    },
    text: {
      fontSize: text?.valueLabel?.size,
      fontWeight: text?.valueLabel?.weight,
      fill: text?.valueLabel?.color,
    },
    line: { strokeWidth: line?.width },
    area: {
      fillOpacity: area?.fillOpacity,
      line: { strokeWidth: area?.lineWidth },
    },
    bar: { opacity: bar?.opacity, cornerRadiusEnd: bar?.cornerRadius },
    point: pointConfig,
    circle: pointConfig,
    square: pointConfig,
  }) as Config;
}

/**
 * The chart style a Vega-Lite config draws with: `chartThemeVegaConfig` read backwards. Given the
 * config Rill resolves for a chart, it states Rill's defaults and the theme's settings in one
 * style, which is what a renderer that takes no Vega-Lite config, such as Flint, is handed.
 */
export function chartStyleOfVegaConfig(config: Config): V1ThemeChartStyle {
  const {
    axisY: axis = {},
    legend = {},
    title = {},
    text = {},
    area = {},
    bar = {},
    point = {},
    ...rest
  } = config as Record<string, Record<string, unknown> | undefined>;
  const symbol = Object.keys(LEGEND_SYMBOL_TYPE).find(
    (name) => LEGEND_SYMBOL_TYPE[name] === legend.symbolType,
  );

  return prune({
    fontFamily: axis.labelFont,
    text: {
      title: {
        size: title.fontSize,
        weight: title.fontWeight,
        color: title.color,
      },
      axisLabel: textRole("label", axis),
      axisTitle: textRole("title", axis),
      legendLabel: textRole("label", legend),
      legendTitle: textRole("title", legend),
      valueLabel: {
        size: text.fontSize,
        weight: text.fontWeight,
        color: text.fill,
      },
    },
    grid: {
      // Vega-Lite's own default, which Rill keeps: gridlines across a measure axis only.
      measure: rest.axisQuantitative?.grid ?? true,
      category: rest.axisDiscrete?.grid ?? false,
      dash: axis.gridDash,
      width: axis.gridWidth,
      color: axis.gridColor,
    },
    axis: {
      line: axis.domain,
      lineWidth: axis.domainWidth,
      ticks: axis.ticks !== false && axis.tickSize !== 0,
      color: axis.domainColor ?? axis.tickColor,
    },
    line: { width: rest.line?.strokeWidth },
    area: {
      fillOpacity: area.fillOpacity,
      lineWidth: (area.line as Record<string, unknown> | undefined)
        ?.strokeWidth,
    },
    bar: { opacity: bar.opacity, cornerRadius: bar.cornerRadiusEnd },
    point: { size: point.size, opacity: point.opacity },
    legend: { position: legend.orient, symbol },
  }) as V1ThemeChartStyle;
}

/** The text role an axis' or a legend's font size, font weight and color properties state. */
function textRole(
  part: "label" | "title",
  guide: Record<string, unknown>,
): Record<string, unknown> {
  return {
    size: guide[`${part}FontSize`],
    weight: guide[`${part}FontWeight`],
    color: guide[`${part}Color`],
  };
}

/** A text role as the font size, font weight and color properties of an axis, legend or header. */
function guideText(
  part: "label" | "title",
  role: V1ThemeChartTextRole | undefined,
): Record<string, unknown> {
  return {
    [`${part}FontSize`]: role?.size,
    [`${part}FontWeight`]: role?.weight,
    [`${part}Color`]: role?.color,
  };
}

/**
 * Drops unset values and the objects left empty by it, so merging the result overrides nothing by
 * accident. Unset is undefined, or the null and empty string the API sends for an unset message
 * and string.
 */
export function prune(value: Record<string, unknown>): Record<string, unknown> {
  const pruned: Record<string, unknown> = {};
  for (const [key, entry] of Object.entries(value)) {
    if (entry === undefined || entry === null || entry === "") continue;
    if (entry && typeof entry === "object" && !Array.isArray(entry)) {
      const nested = prune(entry as Record<string, unknown>);
      if (Object.keys(nested).length > 0) pruned[key] = nested;
    } else {
      pruned[key] = entry;
    }
  }
  return pruned;
}
