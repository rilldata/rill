import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { AdaptedChartBase } from "./adapt";

/** What compiling a Flint chart reported: a hard error, or the layout warnings Flint attached. */
export interface CompileNotices {
  error?: string;
  /** Flint's non-info warnings, e.g. values dropped because a discrete axis overflowed its budget. */
  warnings: string[];
}

export const NO_NOTICES: CompileNotices = { warnings: [] };

/**
 * The banner under the chart: the server's truncation note, else a generic line when only Flint
 * dropped something; the server's explanation and Flint's messages are the hover detail.
 */
export function truncationBanner(
  chart: Pick<AdaptedChartBase, "truncationNote" | "truncationWarning">,
  notices: CompileNotices,
): { note: string; detail: string | undefined } | undefined {
  if (!chart.truncationNote && notices.warnings.length === 0) return undefined;
  const detail = [chart.truncationWarning, ...notices.warnings]
    .filter((text): text is string => !!text)
    .join(" ");
  return {
    note: chart.truncationNote ?? m.mcp_chart_layout_truncated(),
    detail: detail || undefined,
  };
}
