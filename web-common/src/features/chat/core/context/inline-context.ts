import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { prettyFormatResolvedV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import type {
  MetricsViewSpecDimension,
  MetricsViewSpecMeasure,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";

export const INLINE_CHAT_CONTEXT_TAG = "chat-reference";

export enum InlineContextType {
  Explore = "explore",
  MetricsView = "metricsView",
  TimeRange = "timeRange",
  Where = "where",
  Measure = "measure",
  Dimension = "dimension",
  DimensionValues = "dimensionValues",
}

export type InlineContext = {
  type: InlineContextType;
  label?: string;
  metricsView?: string;
  measure?: string;
  dimension?: string;
  timeRange?: string;
  values?: string[];
};

export function inlineContextsAreEqual(
  ctx1: InlineContext,
  ctx2: InlineContext,
) {
  const nonValuesAreEqual =
    ctx1.type === ctx2.type &&
    ctx1.metricsView === ctx2.metricsView &&
    ctx1.measure === ctx2.measure &&
    ctx1.dimension === ctx2.dimension &&
    ctx1.timeRange === ctx2.timeRange;
  if (!nonValuesAreEqual) return false;
  if (!ctx1.values && !ctx2.values) return true;
  else if (!ctx1.values || !ctx2.values) return false;

  return (
    ctx1.values.length === ctx2.values.length &&
    ctx1.values.every((value, index) => value === ctx2.values![index])
  );
}

export function normalizeInlineContext(ctx: InlineContext) {
  return Object.fromEntries(
    Object.entries(ctx).filter(([, v]) => v !== null && v !== undefined),
  ) as InlineContext;
}

export type InlineContextMetadata = Record<string, MetricsViewMetadata>;
export type MetricsViewMetadata = {
  metricsViewSpec: V1MetricsViewSpec;
  measures: Record<string, MetricsViewSpecMeasure>;
  dimensions: Record<string, MetricsViewSpecDimension>;
};

type ContextConfigPerType = {
  getLabel: (ctx: InlineContext, meta: InlineContextMetadata) => string;
};

export const InlineContextConfig: Partial<
  Record<InlineContextType, ContextConfigPerType>
> = {
  [InlineContextType.MetricsView]: {
    getLabel: (ctx, meta) =>
      meta[ctx.metricsView!]?.metricsViewSpec?.displayName || ctx.metricsView!,
  },

  [InlineContextType.TimeRange]: {
    getLabel: (ctx) => {
      if (!ctx.timeRange) return "";
      const [start, end] = ctx.timeRange.split(" to ");
      return prettyFormatResolvedV1TimeRange({
        start: start,
        end: end ?? start,
      });
    },
  },

  [InlineContextType.Measure]: {
    getLabel: (ctx, meta) => {
      const mes = meta[ctx.metricsView!]?.measures[ctx.measure!];
      return getMeasureDisplayName(mes) || ctx.measure!;
    },
  },

  [InlineContextType.Dimension]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      return getDimensionDisplayName(dim) || ctx.dimension!;
    },
  },

  [InlineContextType.DimensionValues]: {
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      const dimLabel = getDimensionDisplayName(dim) || ctx.dimension!;
      return dimLabel + ": " + (ctx.values ?? []).join(", ");
    },
  },
};
