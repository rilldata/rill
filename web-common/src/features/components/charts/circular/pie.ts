import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { VisualizationSpec } from "svelte-vega";
import type { ExprRef, SignalRef } from "vega";
import type { Config } from "vega-lite";
import type { Field } from "vega-lite/types_unstable/channeldef.js";
import type { LayerSpec } from "vega-lite/types_unstable/spec/layer.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import type { Transform } from "vega-lite/types_unstable/transform.js";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createOrderEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { CircularChartSpec } from "./CircularChartProvider";
import {
  OTHER_VALUE,
  OTHER_VALUE_DOMAIN_KEY,
  PERCENT_OF_TOTAL_FIELD,
  PERCENT_OF_TOTAL_TITLE,
  TOTAL_DOMAIN_KEY,
} from "./constants";

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
  const totalValue = data.domainValues?.[TOTAL_DOMAIN_KEY]?.[0];
  const shouldShowTotal =
    totalValue !== undefined && config.measure?.showTotal === true;

  const measureMetaData = config.measure && data.fields[config.measure.field];

  const validPercentOfTotal =
    measureMetaData &&
    "validPercentOfTotal" in measureMetaData &&
    measureMetaData.validPercentOfTotal === true;

  // Show "% of total" in the tooltip when the measure supports it
  const showPercentOfTotal = Boolean(
    config.measure?.field &&
      typeof totalValue === "number" &&
      validPercentOfTotal,
  );

  // Inject the synthetic "Other" row when the provider has set it.
  const otherValue = data.domainValues?.[OTHER_VALUE_DOMAIN_KEY]?.[0];
  const hasOther =
    typeof otherValue === "number" &&
    !!config.color?.field &&
    !!config.measure?.field;

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

  const tooltip = createDefaultTooltipEncoding(
    [config.color, config.measure],
    data,
  );

  if (showPercentOfTotal) {
    tooltip.push({
      field: PERCENT_OF_TOTAL_FIELD,
      title: PERCENT_OF_TOTAL_TITLE,
      type: "quantitative",
      format: ".1%",
    });
  }

  const arcLayer: LayerSpec<Field> | UnitSpec<Field> = {
    mark: {
      type: "arc",
      padAngle: 0.01,
      innerRadius: getInnerRadius(config.innerRadius),
    },
    encoding: {
      theta,
      color,
      order,
      tooltip,
    },
  };

  if (hasOther && config.color?.field && config.measure?.field) {
    const colorField = config.color.field;
    const measureField = config.measure.field;
    const otherRow: Record<string, unknown> = {
      [colorField]: OTHER_VALUE,
      [measureField]: otherValue,
    };
    arcLayer.data = {
      values: [...data.data, otherRow],
    };
  }

  if (showPercentOfTotal && config.measure?.field) {
    const measureFieldRaw = config.measure.field;
    const transforms: Transform[] = [
      {
        calculate: `datum['${measureFieldRaw}'] / ${totalValue}`,
        as: PERCENT_OF_TOTAL_FIELD,
      },
    ];
    arcLayer.transform = transforms;
  }

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
    spec.mark = arcLayer.mark;
    spec.encoding = arcLayer.encoding;
    if (arcLayer.data) {
      spec.data = arcLayer.data;
    }
    if (arcLayer.transform) {
      spec.transform = arcLayer.transform;
    }

    return {
      ...spec,
      ...(vegaConfig && { config: vegaConfig }),
    };
  }
}
