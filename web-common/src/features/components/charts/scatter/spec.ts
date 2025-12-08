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
    (typeof config.color === "object" && config.color?.legendOrientation) || "top";

  const tooltip: TooltipValue[] = [];

  if (config.dimension && config.dimension.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.dimension.field),
      title: data.fields[config.dimension.field]?.displayName || config.dimension.field,
      type: config.dimension.type,
    });
  }

  if (config.x && config.x.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.x.field),
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x.type,
      formatType: sanitizeFieldName(config.x.field),
    });
  }

  if (config.y && config.y.type !== "value") {
    tooltip.push({
      field: sanitizeValueForVega(config.y.field),
      title: data.fields[config.y.field]?.displayName || config.y.field,
      type: config.y.type,
      formatType: sanitizeFieldName(config.y.field),
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

  const colorTooltip = createDefaultTooltipEncoding(
    [config.color],
    data,
  );

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

  spec.encoding = {
    x: xEncoding,
    y: yEncoding,
    ...(config.size && {
      size: {
        field: sanitizeValueForVega(config.size.field),
        title:
          data.fields[config.size.field]?.displayName || config.size.field,
        type: "quantitative",
        scale: {
          zero: false,
        },
        legend: {
          orient: legendOrientation,
        },
      },
    }),
    color: createColorEncoding(config.color, data),
    tooltip: [...tooltip, ...colorTooltip],
  };

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

