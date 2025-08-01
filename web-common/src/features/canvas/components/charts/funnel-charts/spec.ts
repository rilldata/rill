import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { UnitSpec } from "vega-lite/build/src/spec";
import {
  createConfig,
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
  spec.height = 500;

  const colorField =
    config.color === "measure" ? config.measure?.field : config.stage?.field;
  const colorType =
    config.color === "measure"
      ? config.measure?.type || "quantitative"
      : "nominal";

  const vegaConfig = createConfig(config);

  const yEncoding = createPositionEncoding(config.stage, data);
  const tooltip = createDefaultTooltipEncoding(
    [config.measure, config.stage],
    data,
  );

  if (config.measure?.field) {
    if (config.mode === "order") {
      spec.transform = [
        {
          window: [{ op: "row_number", as: "funnel_rank" }],
          sort: [{ field: config.measure.field, order: "descending" }],
        },
        {
          calculate: `pow(0.85, datum.funnel_rank - 1)`,
          as: "funnel_width",
        },
      ];
    } else {
      spec.transform = [
        {
          calculate: `datum['${sanitizeValueForVega(config.measure.field)}']`,
          as: "funnel_width",
        },
      ];
    }
  }

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
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
        axis: {
          labels: false,
          title: null,
          ticks: false,
          domain: false,
        },
      },
      color: {
        field: colorField,
        type: colorType,
        legend: null,
      },
    },
  };

  const valueTextLayer: UnitSpec<Field> = {
    mark: {
      type: "text",
      fontWeight: 600,
      dx: {
        expr: `-(scale('x', datum['funnel_width']) / 2)`,
      },
      color: {
        expr: `luminance ( scale ( 'color', datum['${sanitizeValueForVega(colorField ?? "")}'] ) ) > 0.45 ? '#222' : '#efefef'`,
      },
    },
    encoding: {
      x: {
        field: "funnel_width",
        type: "quantitative",
        stack: "center",
      },
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
        field: "funnel_width",
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
    ...(vegaConfig && { config: vegaConfig }),
  };
}
