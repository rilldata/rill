import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  createCartesianMultiValueTooltipChannel,
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

export function generateVLLineChartSpec(
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

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
    {
      encoding: {
        y: createPositionEncoding(config.y, data),
        color: createColorEncoding(config.color, data),
      },
      layer: [{ mark: "line" }, buildHoverPointOverlay()],
    },
    buildHoverRuleLayer({
      xField,
      domainValues: data.domainValues,
      defaultTooltip: defaultTooltipChannel,
      multiValueTooltipChannel,
      xSort: config.x?.sort,
      primaryColor: data.theme.primary,
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
