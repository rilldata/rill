import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Config } from "vega-lite";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { ExprRef, SignalRef } from "vega-typings";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createOrderEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult, TooltipValue } from "../types";
import type { CircularChartSpec } from "./CircularChartProvider";
import {
  OTHER_SLICE_COLOR_DARK,
  OTHER_SLICE_COLOR_LIGHT,
  OTHER_SLICE_LABEL,
} from "./other-grouping";

function getInnerRadius(innerRadiusPercentage: number | undefined) {
  if (!innerRadiusPercentage) return 0;

  if (innerRadiusPercentage >= 100 || innerRadiusPercentage < 0) {
    console.warn("Inner radius percentage must be between 0 and 100");
    return { expr: `0.5*min(width,height)/2` };
  }

  const decimal = innerRadiusPercentage / 100;
  return { expr: `${decimal}*min(width,height)/2` };
}

function getTotalFontSize(innerRadiusPercentage: number | undefined) {
  if (
    !innerRadiusPercentage ||
    innerRadiusPercentage <= 0 ||
    innerRadiusPercentage >= 100
  ) {
    return 16;
  }

  const decimal = innerRadiusPercentage / 100;
  return { expr: `max(11, min(32, ${decimal}*min(width,height)/4))` };
}

function createPieTooltipEncoding(
  config: CircularChartSpec,
  data: ChartDataResult,
): TooltipValue[] {
  const tooltip: TooltipValue[] = [];

  if (config.color) {
    const colorMeta = data.fields[config.color.field];
    tooltip.push({
      field: config.color.field,
      title: colorMeta?.displayName || config.color.field,
      type: config.color.type as "nominal",
    });
  }

  if (config.measure) {
    const measureMeta = data.fields[config.measure.field];
    tooltip.push({
      field: config.measure.field,
      title: measureMeta?.displayName || config.measure.field,
      type: "quantitative",
      ...(config.measure.type === "quantitative" && {
        formatType: sanitizeFieldName(config.measure.field),
      }),
    });
  }

  tooltip.push({
    field: "__percentage",
    title: "% of Total",
    type: "quantitative",
    format: ".1f",
  });

  return tooltip;
}

export function generateVLPieChartSpec(
  config: CircularChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const totalValue = data.domainValues?.["total"]?.[0];
  const shouldShowTotal =
    totalValue !== undefined && config.measure?.showTotal === true;

  const measureMetaData = config.measure && data.fields[config.measure.field];

  /**
   * The layout property is not typed in the current version of Vega-Lite.
   * This will be fixed when we upgrade to Svelte 5 and subseqent Vega-Lite versions.
   */
  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    {
      legend: {
        layout: {
          right: { anchor: "middle" },
          left: { anchor: "middle" },
          top: { anchor: "middle" },
          bottom: { anchor: "middle" },
        },
      },
    } as unknown as Config<ExprRef | SignalRef>,
    "right",
  );

  const theta = createPositionEncoding(config.measure, data);
  const color = createColorEncoding(config.color, data);
  const order = createOrderEncoding(config.measure);

  const hasOther =
    config.showOther !== false &&
    data.data?.some(
      (d) =>
        config.color?.field &&
        d[config.color.field] === OTHER_SLICE_LABEL,
    );

  if (hasOther) {
    const colorAny = color as Record<string, unknown>;
    const scale = colorAny.scale as
      | { domain?: string[]; range?: string[] }
      | undefined;
    if (scale?.domain && scale?.range) {
      const otherIdx = scale.domain.indexOf(OTHER_SLICE_LABEL);
      if (otherIdx >= 0) {
        scale.range[otherIdx] = data.isDarkMode
          ? OTHER_SLICE_COLOR_DARK
          : OTHER_SLICE_COLOR_LIGHT;
      }
    }
  }

  const grandTotal = data.domainValues?.["__otherTotal"]?.[0] as
    | number
    | undefined;
  const hasPercentage = !!config.measure?.field && (grandTotal ?? 0) > 0;

  const transforms = hasPercentage
    ? [
        {
          calculate: `datum['${config.measure!.field}'] / ${grandTotal} * 100`,
          as: "__percentage",
        },
      ]
    : [];

  const tooltip = hasPercentage
    ? createPieTooltipEncoding(config, data)
    : createDefaultTooltipEncoding([config.color, config.measure], data);

  const arcMark = {
    type: "arc" as const,
    padAngle: 0.01,
    innerRadius: getInnerRadius(config.innerRadius),
  };

  const arcLayer: UnitSpec<Field> = {
    ...(transforms.length > 0
      ? { transform: transforms as UnitSpec<Field>["transform"] }
      : {}),
    mark: arcMark,
    encoding: {
      theta,
      color,
      order,
      tooltip,
    },
  };

  if (shouldShowTotal && totalValue !== undefined) {
    const spec = createMultiLayerBaseSpec();

    const totalLayer: LayerSpec<Field> | UnitSpec<Field> = {
      data: {
        values: [{ total_value: totalValue }],
      },
      mark: {
        type: "text",
        align: "center",
        color: data.isDarkMode ? "#eeeeee" : "#353535",
        baseline: "middle",
        fontWeight: "normal",
        fontSize: getTotalFontSize(config.innerRadius),
      },
      encoding: {
        text: {
          field: "total_value",
          type: "quantitative",
          ...(config.measure?.type === "quantitative" &&
            config.measure?.field && {
              formatType: sanitizeFieldName(config.measure.field),
            }),
          ...(measureMetaData &&
            "format" in measureMetaData && { format: measureMetaData.format }),
        },
        tooltip: null,
      },
    };

    spec.layer = [arcLayer, totalLayer];
    spec.description = "A arc chart with embedded data";

    return {
      ...spec,
      ...(vegaConfig && { config: vegaConfig }),
    };
  } else {
    const spec = createSingleLayerBaseSpec("arc");
    spec.mark = arcMark;
    spec.encoding = arcLayer.encoding;
    if (transforms.length > 0) {
      (spec as unknown as Record<string, unknown>).transform = transforms;
    }

    return {
      ...spec,
      ...(vegaConfig && { config: vegaConfig }),
    };
  }
}
