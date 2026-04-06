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
} from "@rilldata/web-common/features/components/charts/builder";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { LayerSpec } from "vega-lite/types_unstable/spec/layer.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import type { CartesianChartSpec } from "../CartesianChartProvider";
export function generateVLAreaChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y, config.color],
    data,
  );
  const multiValueTooltipChannel = createCartesianMultiValueTooltipChannel(
    { x: config.x, colorField, yField },
    data,
  );

  const xEncoding = createPositionEncoding(config.x, data);
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
        y: { ...createPositionEncoding(config.y, data), stack: "zero" },
        color: createColorEncoding(config.color, data),
      },
      layer: inner,
    },
    buildHoverRuleLayer({
      xField,
      domainValues: data.domainValues,
      defaultTooltip: defaultTooltipChannel,
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
      isDarkMode: data.isDarkMode,
      isInteractive: config.isInteractive,
      pivot:
        xField && yField && colorField && multiValueTooltipChannel?.length
          ? { field: colorField, value: yField, groupby: [xField] }
          : undefined,
    }),
  ];

  spec.layer = layers;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
