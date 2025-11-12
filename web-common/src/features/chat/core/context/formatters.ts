import { ConversationContextType } from "@rilldata/web-common/features/chat/core/context/context-type-data.ts";
import type { MessageContext } from "@rilldata/web-common/features/chat/core/context/context.ts";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  type V1Expression,
  V1Operation,
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { DateTime, Interval } from "luxon";
import { derived } from "svelte/store";

export function createTimeRangeFormatter(context: MessageContext) {
  return derived(context.record, (contextRecord) => {
    const timeRange = contextRecord[ConversationContextType.TimeRange];
    return formatV1TimeRange(timeRange);
  });
}

export function formatV1TimeRange(timeRange: V1TimeRange | undefined) {
  if (!timeRange?.start || !timeRange?.end) return "";

  return prettyFormatTimeRange(
    Interval.fromDateTimes(
      DateTime.fromISO(timeRange.start),
      DateTime.fromISO(timeRange.end),
    ),
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
  );
}

export function createWhereFiltersFormatter(context: MessageContext) {
  return derived(context.record, (contextRecord) => {
    const whereFilters = contextRecord[ConversationContextType.Where];
    return formatV1Expression(whereFilters);
  });
}

export function formatV1Expression(expr: V1Expression | undefined) {
  if (!expr?.cond?.exprs?.length) return [];

  const formattedExpr = expr.cond.exprs
    .map(formatInExpression)
    .filter(Boolean) as string[];

  return formattedExpr;
}

function formatInExpression(expr: V1Expression) {
  if (
    !expr?.cond ||
    !expr.cond.exprs ||
    (expr.cond.op !== V1Operation.OPERATION_IN &&
      expr.cond.op !== V1Operation.OPERATION_NIN)
  ) {
    return undefined;
  }
  const ident = expr.cond.exprs[0].ident ?? "";
  const values = expr.cond.exprs.slice(1).map((e) => e.val);
  return `${ident}: ${values.join(", ")}`;
}
