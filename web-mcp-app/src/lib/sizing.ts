import type { McpUiHostContext } from "@modelcontextprotocol/ext-apps";

/** Inline cards keep the chart between these heights; the host auto-fits the card around it. */
export const INLINE_CHART_MIN_HEIGHT = 220;
export const INLINE_CHART_MAX_HEIGHT = 360;
export const INLINE_CHART_DEFAULT_HEIGHT = 300;
/** Header, footer and paddings around the inline chart, in pixels. */
export const INLINE_CARD_CHROME_HEIGHT = 112;
/** Design guidelines ask for at least 320px; the card never lays out narrower than this. */
export const MIN_WIDGET_WIDTH = 320;

export type ContainerDimensions = McpUiHostContext["containerDimensions"];

export function availableHeight(
  dimensions: ContainerDimensions,
): number | undefined {
  if (!dimensions) return undefined;
  if ("height" in dimensions && typeof dimensions.height === "number") {
    return dimensions.height;
  }
  if ("maxHeight" in dimensions && typeof dimensions.maxHeight === "number") {
    return dimensions.maxHeight;
  }
  return undefined;
}

export function availableWidth(
  dimensions: ContainerDimensions,
): number | undefined {
  if (!dimensions) return undefined;
  if ("width" in dimensions && typeof dimensions.width === "number") {
    return dimensions.width;
  }
  if ("maxWidth" in dimensions && typeof dimensions.maxWidth === "number") {
    return dimensions.maxWidth;
  }
  return undefined;
}

/**
 * Chart height for the inline card: the host's available height minus the card chrome,
 * clamped to the inline range. Without a known height the default applies.
 */
export function inlineChartHeight(
  dimensions: ContainerDimensions,
  chromeHeight = INLINE_CARD_CHROME_HEIGHT,
): number {
  const available = availableHeight(dimensions);
  if (available === undefined) return INLINE_CHART_DEFAULT_HEIGHT;
  return Math.min(
    INLINE_CHART_MAX_HEIGHT,
    Math.max(INLINE_CHART_MIN_HEIGHT, Math.floor(available - chromeHeight)),
  );
}
