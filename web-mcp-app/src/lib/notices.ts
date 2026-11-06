import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { AdaptedChartBase } from "./adapt";

/** What compiling a Flint chart reported: a hard error, or that Flint left values out. */
export interface CompileNotices {
  error?: string;
  /** Flint's messages about the values it left out to fit the chart. */
  dropped: string[];
}

export const NO_NOTICES: CompileNotices = { dropped: [] };

/**
 * Flint warning codes that mean values are missing from the chart: an axis, legend or facet
 * over its budget, and a pyramid's groups past the first two. The others (an invalid option
 * value that falls back to the default, negative values on a pyramid) drop nothing.
 */
const DROPPED_VALUES_WARNINGS: ReadonlySet<string> = new Set([
  "overflow",
  "too-many-groups-pyramid",
]);

/** The notices of a `compileFlintSpec` result. */
export function compileNotices(compiled: {
  error: string | null;
  warnings: { code: string; message: string }[];
}): CompileNotices {
  return {
    error: compiled.error ?? undefined,
    dropped: compiled.warnings
      .filter((warning) => DROPPED_VALUES_WARNINGS.has(warning.code))
      .map((warning) => warning.message),
  };
}

/**
 * The banner under the chart: the server's truncation note, else a generic line when only Flint
 * dropped something; the server's explanation and Flint's messages are the hover detail.
 */
export function truncationBanner(
  chart: Pick<AdaptedChartBase, "truncationNote" | "truncationWarning">,
  notices: CompileNotices,
): { note: string; detail: string | undefined } | undefined {
  if (!chart.truncationNote && notices.dropped.length === 0) return undefined;
  const detail = [chart.truncationWarning, ...notices.dropped]
    .filter((text): text is string => !!text)
    .join(" ");
  return {
    note: chart.truncationNote ?? m.mcp_chart_layout_truncated(),
    detail: detail || undefined,
  };
}
