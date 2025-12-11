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
import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import Measure from "@rilldata/web-common/features/chat/core/context/icons/Measure.svelte";
import Dimension from "@rilldata/web-common/features/chat/core/context/icons/Dimension.svelte";
import { fieldTypeToSymbol } from "@rilldata/web-common/lib/duckdb-data-types.ts";

export const INLINE_CHAT_CONTEXT_TAG = "chat-reference";

export enum InlineContextType {
  Explore = "explore",
  MetricsView = "metricsView",
  TimeRange = "timeRange",
  Where = "where",
  Measure = "measure",
  Dimension = "dimension",
  DimensionValues = "dimensionValues",
  Model = "model",
  Column = "column",
}

export type InlineContext = {
  type: InlineContextType;
  label?: string;
  value: string; // Main value for this context.
  metricsView?: string;
  measure?: string;
  dimension?: string;
  timeRange?: string;
  values?: string[];
  model?: string;
  column?: string;
  columnType?: string; // TODO: is this needed here?
  filePath?: string;
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
    ctx1.timeRange === ctx2.timeRange &&
    ctx1.model === ctx2.model &&
    ctx1.column === ctx2.column &&
    ctx1.columnType === ctx2.columnType &&
    ctx1.filePath === ctx2.filePath;
  if (!nonValuesAreEqual) return false;
  if (!ctx1.values && !ctx2.values) return true;
  else if (!ctx1.values || !ctx2.values) return false;

  return (
    ctx1.values.length === ctx2.values.length &&
    ctx1.values.every((value, index) => value === ctx2.values![index])
  );
}

export function inlineContextIsWithin(src: InlineContext, tar: InlineContext) {
  switch (src.type) {
    case InlineContextType.MetricsView:
      return src.metricsView === tar.metricsView;
    case InlineContextType.Model:
      return tar.model === src.model;
  }
  return false;
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
  editable: boolean;
  getLabel: (ctx: InlineContext, meta: InlineContextMetadata) => string;
  getIcon?: (ctx: InlineContext) => any | undefined;
};

export const InlineContextConfig: Partial<
  Record<InlineContextType, ContextConfigPerType>
> = {
  [InlineContextType.MetricsView]: {
    editable: true,
    getLabel: (ctx, meta) =>
      meta[ctx.metricsView!]?.metricsViewSpec?.displayName || ctx.metricsView!,
    getIcon: () => resourceIconMapping[ResourceKind.MetricsView],
  },

  [InlineContextType.TimeRange]: {
    editable: false,
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
    editable: true,
    getLabel: (ctx, meta) => {
      const mes = meta[ctx.metricsView!]?.measures[ctx.measure!];
      return getMeasureDisplayName(mes) || ctx.measure!;
    },
    getIcon: () => Measure,
  },

  [InlineContextType.Dimension]: {
    editable: true,
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      return getDimensionDisplayName(dim) || ctx.dimension!;
    },
    getIcon: () => Dimension,
  },

  [InlineContextType.DimensionValues]: {
    editable: true,
    getLabel: (ctx, meta) => {
      const dim = meta[ctx.metricsView!]?.dimensions[ctx.dimension!];
      const dimLabel = getDimensionDisplayName(dim) || ctx.dimension!;
      return dimLabel + ": " + (ctx.values ?? []).join(", ");
    },
  },

  [InlineContextType.Model]: {
    editable: true,
    getLabel: (ctx) => ctx.model ?? "",
    getIcon: () => resourceIconMapping[ResourceKind.Model],
  },

  [InlineContextType.Column]: {
    editable: true,
    getLabel: (ctx) => ctx.column ?? "",
    getIcon: (ctx) => fieldTypeToSymbol(ctx.columnType ?? ""),
  },
};
