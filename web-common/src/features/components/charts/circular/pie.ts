import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
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
import { resolveCSSVariable } from "../util";
import type { CircularChartSpec } from "./CircularChartProvider";
import {
  DEFAULT_LABELS_FORMAT,
  type LabelsConfig,
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

function getLabelRadius(innerRadiusPercentage: number | undefined) {
  const inner =
    !innerRadiusPercentage ||
    innerRadiusPercentage <= 0 ||
    innerRadiusPercentage >= 100
      ? 0
      : innerRadiusPercentage / 100;
  const ratio = inner + (1 - inner) / 2;
  return { expr: `${ratio}*min(width,height)/2` };
}

function getLabelFontSize() {
  return { expr: "max(9, min(16, min(width, height) / 22))" };
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

  const labels: LabelsConfig = config.labels ?? {};
  const labelsFormat = labels.format ?? DEFAULT_LABELS_FORMAT;
  const labelsThreshold =
    typeof labels.threshold === "number" ? labels.threshold : 0;
  const showLabels = Boolean(
    labels.show && config.measure?.field && typeof totalValue === "number",
  );
  const labelsNeedPercent =
    showLabels && (labelsFormat === "percent" || labelsThreshold > 0);

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

  const needsPercentOfTotal =
    (showPercentOfTotal || labelsNeedPercent) && !!config.measure?.field;
  if (needsPercentOfTotal && config.measure?.field) {
    const measureFieldRaw = config.measure.field;
    const transforms: Transform[] = [
      {
        calculate: `datum['${measureFieldRaw}'] / ${totalValue}`,
        as: PERCENT_OF_TOTAL_FIELD,
      },
    ];
    arcLayer.transform = transforms;
  }

  let labelLayer: LayerSpec<Field> | UnitSpec<Field> | undefined;
  if (showLabels && config.measure?.field && config.color?.field) {
    const measureFieldRaw = config.measure.field;
    const colorFieldRaw = config.color.field;
    const labelEncoding: Record<string, unknown> = {
      theta: {
        field: sanitizeValueForVega(measureFieldRaw),
        type: "quantitative",
        stack: true,
      },
      order,
    };

    if (labelsFormat === "percent") {
      labelEncoding.text = {
        field: PERCENT_OF_TOTAL_FIELD,
        type: "quantitative",
        format: ".0%",
      };
    } else {
      labelEncoding.text = {
        field: measureFieldRaw,
        type: "quantitative",
        formatType: sanitizeFieldName(measureFieldRaw),
        ...(measureMetaData &&
          "format" in measureMetaData && { format: measureMetaData.format }),
      };
    }

    if (labelsThreshold > 0) {
      labelEncoding.opacity = {
        condition: {
          test: `datum['${PERCENT_OF_TOTAL_FIELD}'] >= ${labelsThreshold / 100}`,
          value: 1,
        },
        value: 0,
      };
    }

    // Background-aware text color: pick dark text on light slices and
    // light text on dark slices, mirroring the heatmap pattern.
    const darkTextColor = resolveCSSVariable(
      "var(--color-gray-900)",
      data.isDarkMode,
    );
    const lightTextColor = resolveCSSVariable(
      "var(--color-gray-50)",
      data.isDarkMode,
    );

    labelLayer = {
      mark: {
        type: "text",
        radius: getLabelRadius(config.innerRadius),
        fontSize: getLabelFontSize(),
        fontWeight: "bold",
        color: {
          expr: `luminance(scale('color', datum['${sanitizeValueForVega(colorFieldRaw)}'])) > 0.45 ? '${darkTextColor}' : '${lightTextColor}'`,
        },
      },
      encoding: labelEncoding,
    };
  }

  const useMultiLayer =
    (shouldShowTotal && totalValue !== undefined) || !!labelLayer;

  if (useMultiLayer) {
    const spec = createMultiLayerBaseSpec();
    const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [arcLayer];

    // When the label layer is present, lift the arc layer's data and
    // transform to the top-level spec so both layers share the same
    // dataset
    if (labelLayer) {
      if (arcLayer.data) {
        spec.data = arcLayer.data;
        delete arcLayer.data;
      }
      if (arcLayer.transform) {
        spec.transform = arcLayer.transform;
        delete arcLayer.transform;
      }
      layers.push(labelLayer);
    }

    if (shouldShowTotal && totalValue !== undefined) {
      const totalLayer: LayerSpec<Field> | UnitSpec<Field> = {
        data: {
          values: [{ total_value: totalValue }],
        },
        mark: {
          type: "text",
          align: "center",
          color: data.isDarkMode ? "#eeeeee" : "#353535",
          baseline: "middle",
          fontWeight: "bold",
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
              "format" in measureMetaData && {
                format: measureMetaData.format,
              }),
          },
          tooltip: null,
        },
      };
      layers.push(totalLayer);
    }

    spec.layer = layers;
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
