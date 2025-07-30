import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { UnitSpec } from "vega-lite/build/src/spec";
import {
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { FunnelChartSpec } from "./FunnelChart";

export function generateVLFunnelChartSpec(
  config: FunnelChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.stage);

  const tooltip = createDefaultTooltipEncoding(
    [config.measure, config.stage],
    data,
  );

  const yEncoding = createPositionEncoding(config.stage, data);
  const xEncoding = createPositionEncoding(config.measure, data);

  spec.encoding = {
    y: {
      ...yEncoding,
      sort: null,
      axis: {
        labels: false,
        title: null,
        ticks: false,
        domain: false,
      },
    },
    tooltip,
  };

  // Main bar layer
  const barLayer: UnitSpec<Field> = {
    mark: "bar",
    encoding: {
      x: {
        ...xEncoding,
        stack: "center",
        axis: {
          labels: false,
          title: null,
          ticks: false,
          domain: false,
        },
      },
      color: {
        field: config.stage?.field,
        type: "nominal",
        legend: null,
      },
    },
  };

  const valueTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      align: "right",
      fontWeight: 600,
      color: {
        expr: `luminance ( scale ( 'color', datum['${sanitizeValueForVega(config.stage?.field ?? "")}'] ) ) > 0.45 ? '#222' : '#efefef'`,
      },
    },
    encoding: {
      text: {
        field: config.measure?.field,
        type: config.measure?.type || "quantitative",
        ...(config.measure?.type === "quantitative" &&
          config.measure?.field && {
            formatType: sanitizeFieldName(config.measure.field),
          }),
      },
    },
  };

  const labelTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      dx: 10,
      align: "left",
      limit: 200,
    },
    encoding: {
      x: {
        field: config.measure?.field,
        type: "quantitative",
        stack: "center",
      },
      text: {
        field: config.stage?.field,
        type: "nominal",
      },
    },
  };

  spec.layer = [barLayer, valueTextLayer, labelTextLayer];

  return {
    ...spec,
    height: 500,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
