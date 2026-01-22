import { resourceIconMapping } from "@rilldata/web-common/features/entity-management/resource-icon-mapping.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { prettyFormatResolvedV1TimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import Measure from "@rilldata/web-common/features/chat/core/context/icons/Measure.svelte";
import Dimension from "@rilldata/web-common/features/chat/core/context/icons/Dimension.svelte";
import { fieldTypeToSymbol } from "@rilldata/web-common/lib/duckdb-data-types.ts";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { InlineContextMetadata } from "@rilldata/web-common/features/chat/core/context/metadata.ts";
import type { V1ComponentSpec } from "@rilldata/web-common/runtime-client";
import {
  getHeaderForComponent,
  getIconForComponent,
} from "@rilldata/web-common/features/canvas/components/util.ts";

type ContextConfigPerType = {
  editable: boolean;
  typeLabel?: string;
  getLabel: (ctx: InlineContext, meta: InlineContextMetadata) => string;
  getTooltip?: (ctx: InlineContext, meta: InlineContextMetadata) => string;
  getIcon?: (
    ctx: InlineContext,
    meta: InlineContextMetadata,
  ) => any | undefined;
};

/**
 * Configuration for each inline context type.
 * Currently defines:
 * - Whether the context is editable. In future all will be editable.
 * - A function to get the label for the context, given the context and metadata.
 * - An optional function to get the icon for the context, given the context.
 */
export const InlineContextConfig: Record<
  InlineContextType,
  ContextConfigPerType
> = {
  [InlineContextType.MetricsView]: {
    editable: true,
    typeLabel: "Metrics View",
    getLabel: (ctx, meta) =>
      meta.metricsViewSpecs[ctx.metricsView!]?.metricsViewSpec?.displayName ||
      ctx.metricsView!,
    getIcon: () => resourceIconMapping[ResourceKind.MetricsView],
  },

  [InlineContextType.Canvas]: {
    editable: true,
    typeLabel: "Canvas",
    getLabel: (ctx, meta) =>
      meta.canvasSpecs[ctx.canvas!]?.displayName || ctx.canvas!,
    getIcon: () => resourceIconMapping[ResourceKind.Canvas],
  },

  [InlineContextType.CanvasComponent]: {
    editable: true,
    typeLabel: "Canvas Component",
    getLabel: (ctx, meta) =>
      getCanvasComponentLabel(
        ctx.canvasComponent!,
        meta.componentSpecs[ctx.canvasComponent!],
      ),
    getIcon: (ctx, meta) => {
      const componentSpec = meta.componentSpecs[ctx.canvasComponent!];
      return getIconForComponent(componentSpec?.renderer as any);
    },
    getTooltip: (ctx, meta) =>
      `For ${InlineContextConfig[InlineContextType.Canvas].getLabel(ctx, meta)}`,
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
    getTooltip: (ctx, meta) =>
      `For ${InlineContextConfig[InlineContextType.MetricsView].getLabel(ctx, meta)}`,
  },

  [InlineContextType.Measure]: {
    editable: true,
    getLabel: (ctx, meta) => {
      const mes =
        meta.metricsViewSpecs[ctx.metricsView!]?.measures[ctx.measure!];
      return getMeasureDisplayName(mes) || ctx.measure!;
    },
    getTooltip: (ctx, meta) =>
      `From ${InlineContextConfig[InlineContextType.MetricsView].getLabel(ctx, meta)}`,
    getIcon: () => Measure,
  },

  [InlineContextType.Dimension]: {
    editable: true,
    getLabel: (ctx, meta) => {
      const dim =
        meta.metricsViewSpecs[ctx.metricsView!]?.dimensions[ctx.dimension!];
      return getDimensionDisplayName(dim) || ctx.dimension!;
    },
    getTooltip: (ctx, meta) =>
      `From ${InlineContextConfig[InlineContextType.MetricsView].getLabel(ctx, meta)}`,
    getIcon: () => Dimension,
  },

  [InlineContextType.DimensionValues]: {
    editable: true,
    getLabel: (ctx, meta) => {
      const dim =
        meta.metricsViewSpecs[ctx.metricsView!]?.dimensions[ctx.dimension!];
      const dimLabel = getDimensionDisplayName(dim) || ctx.dimension!;
      return dimLabel + ": " + (ctx.values ?? []).join(", ");
    },
  },

  [InlineContextType.Model]: {
    editable: true,
    typeLabel: "Model",
    getLabel: (ctx) => ctx.model ?? "",
    getIcon: () => resourceIconMapping[ResourceKind.Model],
  },

  [InlineContextType.Column]: {
    editable: true,
    getLabel: (ctx) => ctx.column ?? "",
    getTooltip: (ctx, meta) =>
      `From ${InlineContextConfig[InlineContextType.Model].getLabel(ctx, meta)}`,
    getIcon: (ctx) => fieldTypeToSymbol(ctx.columnType ?? ""),
  },
};

const rowColMatcher = /-(\d+)-(\d+)$/;
export function getCanvasComponentLabel(
  componentName: string,
  componentSpec: V1ComponentSpec | undefined,
) {
  const userDefinedTitle =
    (componentSpec?.rendererProperties?.title as string | undefined) ||
    componentSpec?.displayName;
  if (userDefinedTitle) return userDefinedTitle;
  const header = getHeaderForComponent(componentSpec?.renderer as any);

  const rowColMatch = rowColMatcher.exec(componentName);
  if (!rowColMatch) return header;
  const rowColPart = ` at Row: ${rowColMatch[1]}, Col: ${rowColMatch[2]}`;

  return header + rowColPart;
}
