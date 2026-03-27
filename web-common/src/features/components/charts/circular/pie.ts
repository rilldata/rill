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
  createMultiLayerBaseSpec,
  createOrderEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import { OTHER_FLAG_FIELD, OTHER_LABEL } from "./other-grouping";
import type { ChartDataResult } from "../types";
import type { CircularChartSpec } from "./CircularChartProvider";

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

  // Override "Other" slice color to use muted fill
  const hasOther = data.data?.some((d) => d[OTHER_FLAG_FIELD] === true);
  if (hasOther && "scale" in color && color.scale) {
    const scale = color.scale as {
      domain?: unknown[];
      range?: unknown[];
      [key: string]: unknown;
    };
    const mutedColor = data.isDarkMode ? "#374151" : "#e5e7eb";
    if (Array.isArray(scale.domain) && !scale.domain.includes(OTHER_LABEL)) {
      scale.domain.push(OTHER_LABEL);
    }
    if (Array.isArray(scale.range) && !scale.range.includes(mutedColor)) {
      scale.range.push(mutedColor);
    }
  }

  const resolvedBorderColor = data.isDarkMode ? "#374151" : "#e5e7eb";

  const arcMark = {
    type: "arc" as const,
    padAngle: 0.01,
    innerRadius: getInnerRadius(config.innerRadius),
    stroke: {
      expr: `datum.${OTHER_FLAG_FIELD} ? '${resolvedBorderColor}' : null`,
    },
    strokeWidth: { expr: `datum.${OTHER_FLAG_FIELD} ? 1 : 0` },
    strokeDash: {
      expr: `datum.${OTHER_FLAG_FIELD} ? [4, 3] : [0, 0]`,
    },
  };

  const arcEncoding = {
    theta,
    color,
    order,
    tooltip: { value: null },
  };

  const arcLayer: LayerSpec<Field> | UnitSpec<Field> = {
    mark: arcMark,
    encoding: arcEncoding,
  } as LayerSpec<Field> | UnitSpec<Field>;

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
    spec.encoding = arcEncoding;

    return {
      ...spec,
      ...(vegaConfig && { config: vegaConfig }),
    };
  }
}
