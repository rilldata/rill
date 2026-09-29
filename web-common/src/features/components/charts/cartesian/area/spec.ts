import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
  createStackOverride,
} from "@rilldata/web-common/features/components/charts/builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { LayerSpec } from "vega-lite/types_unstable/spec/layer.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import type { CartesianChartSpec } from "../CartesianChartProvider";
import { toVerticalSpec } from "../orientation";

export function generateVLAreaChartSpec(
  chartConfig: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  // Area charts are always drawn with the dimension on x. The reconciler
  // rejects a measure on x for them; reading the fields by role here keeps
  // the builder consistent with the provider, which queries the same way.
  const config = toVerticalSpec(chartConfig);

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const sanitizedXField = sanitizeValueForVega(config.x?.field);
  const yField = config.y?.field;
  const sanitizedYField = sanitizeValueForVega(yField);

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y, config.color],
    data,
  );
  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  const xEncoding = createPositionEncoding(config.x, data, "x");
  xEncoding.scale = { ...(xEncoding.scale ?? {}), padding: 8 };
  spec.encoding = { x: xEncoding };

  const inner: UnitSpec<Field>[] = [
    { mark: "area" },
    { mark: { type: "line", opacity: 0.5 } },
    buildHoverPointOverlay(),
  ];

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    {
      encoding: {
        y: {
          ...createPositionEncoding(config.y, data, "y"),
          stack: "zero",
          ...createStackOverride(config.y, colorField),
        },
        color: createColorEncoding(config.color, data),
      },
      layer: inner,
    },
    buildHoverRuleLayer({
      xField: sanitizedXField,
      domainValues: data.domainValues,
      defaultTooltip: defaultTooltipChannel,
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
      isDarkMode: data.isDarkMode,
      isInteractive: config.isInteractive,
      pivot:
        sanitizedXField &&
        sanitizedYField &&
        colorField &&
        multiValueTooltipChannel?.length
          ? {
              field: colorField,
              value: sanitizedYField,
              groupby: [sanitizedXField],
            }
          : undefined,
    }),
  ];

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
    ...(config.isInteractive && sanitizedXField
      ? { usermeta: { brushTemporalField: sanitizedXField } }
      : {}),
  };
}
