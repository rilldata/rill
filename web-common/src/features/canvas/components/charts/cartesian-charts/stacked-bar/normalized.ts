import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { TooltipValue } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import {
  buildHoverRuleLayer,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

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

  let multiValueTooltipChannel: TooltipValue[] | undefined;
  if (colorField && config.x && yField) {
    multiValueTooltipChannel = data.data?.map((value) => ({
      field: sanitizeValueForVega(value?.[colorField] as string),
      type: "quantitative",
      formatType: sanitizeFieldName(yField),
    }));

    multiValueTooltipChannel.unshift({
      field: xField,
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x?.type,
      ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });

    multiValueTooltipChannel = multiValueTooltipChannel.slice(0, 50);
  }

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    buildHoverRuleLayer({
      xField,
      yField,
      isBarMark: true,
      defaultTooltip: baseEncoding.tooltip as TooltipValue[],
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
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
