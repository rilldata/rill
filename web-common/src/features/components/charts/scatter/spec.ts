import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "@rilldata/web-common/features/components/charts/builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { ScatterPlotChartSpec } from "./ScatterPlotChartProvider";

export function generateVLScatterPlotSpec(
  config: ScatterPlotChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("point");

  const vegaConfig = createConfigWithLegend(config, config.color);

  const legendOrientation =
    (typeof config.color === "object" && config.color?.legendOrientation) ||
    "top";

  const tooltip: TooltipValue[] = [];

  if (config.dimension && config.dimension.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.dimension.field),
      title:
        data.fields[config.dimension.field]?.displayName ||
        config.dimension.field,
      type: config.dimension.type,
    });
  }

  if (config.x && config.x.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.x.field),
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x.type,
      ...(config.x.type === "quantitative" && {
        formatType: sanitizeFieldName(config.x.field),
      }),
    });
  }

  if (config.y && config.y.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.y.field),
      title: data.fields[config.y.field]?.displayName || config.y.field,
      type: config.y.type,
      ...(config.y.type === "quantitative" && {
        formatType: sanitizeFieldName(config.y.field),
      }),
    });
  }

  if (config.size && config.size.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.size.field),
      title: data.fields[config.size.field]?.displayName || config.size.field,
      type: config.size.type,
      formatType: sanitizeFieldName(config.size.field),
    });
  }

  const colorTooltip = createDefaultTooltipEncoding([config.color], data);

  const xEncoding = createPositionEncoding(config.x, data);
  const yEncoding = createPositionEncoding(config.y, data);

  if (xEncoding.axis) {
    xEncoding.axis = {
      ...xEncoding.axis,
      grid: true,
    };
  } else {
    xEncoding.axis = { grid: true };
  }

  if (xEncoding.scale) {
    xEncoding.scale = {
      ...xEncoding.scale,
      nice: true,
    };
  } else {
    xEncoding.scale = {
      nice: true,
    };
  }

  if (yEncoding.scale) {
    yEncoding.scale = {
      ...yEncoding.scale,
      nice: true,
    };
  } else {
    yEncoding.scale = {
      nice: true,
    };
  }

  spec.encoding = {
    x: xEncoding,
    y: yEncoding,
    ...(config.size && {
      size: {
        field: sanitizeValueForVega(config.size.field),
        title: data.fields[config.size.field]?.displayName || config.size.field,
        type: "quantitative",
        scale: {
          zero: false,
          range: [40, 400], // Set minimum size to 40 (increased from default ~30)
        },
        legend: {
          orient: legendOrientation,
        },
      },
    }),
    color: createColorEncoding(config.color, data),
    tooltip: [...tooltip, ...colorTooltip],
  };

  if (!config.size && spec.mark && typeof spec.mark === "object") {
    spec.mark = {
      ...spec.mark,
      size: 40,
    } as typeof spec.mark;
  }

  return {
    ...spec,
    selection: {
      grid: {
        type: "interval",
        bind: "scales",
        translate: "[mousedown[event.shiftKey], mouseup] > mousemove",
        zoom: "wheel![!event.shiftKey]",
        clear: "dblclick",
      },
    },
    ...(vegaConfig && { config: vegaConfig }),
  } as VisualizationSpec;
}
