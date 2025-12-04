import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { CartesianChartSpec } from "../CartesianChartProvider";

export function generateVLStackedBarNormalizedSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const baseEncoding = createEncoding(config, data);
  const vegaConfig = createConfigWithLegend(config, config.color);

  if (baseEncoding.y && config.y?.field) {
    const yField = config.y.field;

    baseEncoding.y = {
      ...baseEncoding.y,
      stack: "normalize",
      ...(baseEncoding.y && {
        scale: {
          zero: false,
        },
      }),
      axis: {
        ...(!config.y.showAxisTitle && { title: null }),
        format: ".0%",
      },
    };

    // Add a transform to calculate the percentage
    spec.transform = [
      {
        joinaggregate: [
          {
            op: "sum",
            field: yField,
            as: "total",
          },
        ],
        groupby: config.x?.field ? [config.x.field] : [],
      },
      {
        calculate: `datum['${yField}'] / datum.total`,
        as: "percentage",
      },
    ];

    // Add percentage to tooltip
    const tooltipValues = createDefaultTooltipEncoding(
      [config.x, config.y, config.color],
      data,
    );
    baseEncoding.tooltip = tooltipValues
      .map((t: TooltipValue) => {
        if (t.field === yField) {
          return [
            {
              ...t,
            },
            {
              ...t,
              title: `${t.title} (%)`,
              field: "percentage",
              formatType: undefined,
              format: ".1%",
            },
          ];
        }
        return t;
      })
      .flat();
  }

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    buildHoverRuleLayer({
      xField,
      domainValues: data.domainValues,
      isBarMark: true,
      defaultTooltip: baseEncoding.tooltip as TooltipValue[],
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
      isDarkMode: data.isDarkMode,
      xBand: config.x?.type === "temporal" ? 0.5 : undefined,
      pivot:
        xField && yField && colorField && multiValueTooltipChannel?.length
          ? { field: colorField, value: yField, groupby: [xField] }
          : undefined,
    }),
    {
      mark: { type: "bar", clip: true, width: { band: 0.9 } },
      encoding: {
        ...baseEncoding,
      },
    },
  ];

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
