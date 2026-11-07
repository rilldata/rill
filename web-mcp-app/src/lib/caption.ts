import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { DateTime } from "luxon";
import type { WireTimeRange } from "./contract";

export interface CaptionInput {
  range?: WireTimeRange;
  /** The grain in use; shown only when the chart has a temporal axis. */
  grain?: V1TimeGrain;
  showGrain: boolean;
  hasComparison: boolean;
  metricsViewLabel?: string;
  timeZone: string;
  locale: string;
}

const GRAINS_FINER_THAN_DAY: ReadonlySet<V1TimeGrain> = new Set([
  V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
]);

function isStartOfDay(dt: DateTime): boolean {
  return dt.hour === 0 && dt.minute === 0 && dt.second === 0;
}

/**
 * Human-readable time range such as "Sep 9 – Sep 15, 2023".
 * Ranges are half-open on the wire; a day-aligned end is shown as the last included day,
 * which is how Rill's own time controls label ranges.
 */
export function formatTimeRange(
  range: WireTimeRange,
  timeZone: string,
  locale: string,
  grain?: V1TimeGrain,
): string {
  const start = DateTime.fromISO(range.start, { zone: timeZone }).setLocale(
    locale,
  );
  let end = DateTime.fromISO(range.end, { zone: timeZone }).setLocale(locale);
  if (!start.isValid || !end.isValid) return `${range.start} – ${range.end}`;

  const subDay = grain !== undefined && GRAINS_FINER_THAN_DAY.has(grain);
  const dayAligned = isStartOfDay(start) && isStartOfDay(end);
  if (!subDay && dayAligned) {
    end = end.minus({ days: 1 });
  }

  if (subDay || !dayAligned) {
    const withTime: Intl.DateTimeFormatOptions = {
      month: "short",
      day: "numeric",
      year: start.year === end.year ? undefined : "numeric",
      hour: "numeric",
      minute: "2-digit",
    };
    const startLabel = start.toLocaleString(withTime);
    const endLabel = end.toLocaleString({ ...withTime, year: "numeric" });
    return `${startLabel} – ${endLabel}`;
  }

  const sameDay = start.hasSame(end, "day");
  const full: Intl.DateTimeFormatOptions = {
    month: "short",
    day: "numeric",
    year: "numeric",
  };
  if (sameDay) return start.toLocaleString(full);
  const startLabel = start.toLocaleString(
    start.year === end.year ? { month: "short", day: "numeric" } : full,
  );
  return `${startLabel} – ${end.toLocaleString(full)}`;
}

export function grainLabel(grain: V1TimeGrain): string {
  return `${m.time_grain_by()} ${TIME_GRAIN[grain]?.label ?? grain}`;
}

/** Caption parts in display order; the UI joins them with a separator. */
export function buildCaptionParts(input: CaptionInput): string[] {
  const parts: string[] = [];
  if (input.range) {
    parts.push(
      formatTimeRange(input.range, input.timeZone, input.locale, input.grain),
    );
  }
  if (input.showGrain && input.grain) {
    parts.push(grainLabel(input.grain));
  }
  if (input.hasComparison) {
    parts.push(m.mcp_chart_vs_previous_period());
  }
  if (input.metricsViewLabel) {
    parts.push(input.metricsViewLabel);
  }
  return parts;
}

export const CAPTION_SEPARATOR = " · ";
