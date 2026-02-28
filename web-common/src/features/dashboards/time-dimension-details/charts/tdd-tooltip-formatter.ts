import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import {
  MainAreaColorGradientDark,
  MainLineColor,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { type TDDAlternateCharts, TDDChart } from "../types";

export function escapeHTML(value: any): string {
  return String(value).replace(/&/g, "&amp;").replace(/</g, "&lt;");
}

function formatMeasureTooltipValue(
  value: unknown,
  measureValueFormatter?: (value: number) => string | number | null | undefined,
): string {
  if (typeof value === "number" && measureValueFormatter) {
    const formattedValue = measureValueFormatter(value);
    if (formattedValue !== null && formattedValue !== undefined) {
      return String(formattedValue);
    }
  }
  return String(value);
}

function createColorMap(
  measureLabel: string,
  selectedDimensionValues: (string | null)[],
  isTimeComparison: boolean,
): Record<string, string> {
  const colorMap: Record<string, string> = {};
  selectedDimensionValues.forEach((dimValue, i) => {
    colorMap[String(dimValue)] = COMPARIONS_COLORS[i];
  });

  colorMap[measureLabel] = MainLineColor;

  if (isTimeComparison && selectedDimensionValues.length === 0) {
    colorMap["current_period"] = MainLineColor;
    colorMap["comparison_period"] = MainAreaColorGradientDark;
  }

  return colorMap;
}

function formatTimeHeader(
  Time: any,
  interval: V1TimeGrain | undefined,
): string {
  const formattedTime = interval
    ? new Date(Time).toLocaleDateString(
        undefined,
        TIME_GRAIN[interval].formatDate,
      )
    : Time.toString();
  return `<h2>${formattedTime}</h2>`;
}

function generateStackedAreaContent(
  keys: string[],
  rest: Record<string, any>,
  colorMap: Record<string, string>,
  isTimeComparison: boolean,
  measureValueFormatter?: (value: number) => string | number | null | undefined,
) {
  let content = "";
  for (const key of keys) {
    const val = rest[key];
    if (val === undefined || val === null || val === "NaN") continue;

    let label = key;
    let keyColor = colorMap[String(key)] || "#000";

    if (isTimeComparison) {
      if (key.startsWith("comparison.")) {
        keyColor = colorMap["comparison_period"];
        label = "Previous Period";
      } else {
        keyColor = colorMap["current_period"];
        label = "Current Period";
      }
    }

    content += `
      <tr>
        <td class="color">
          <svg width="16" height="16">
            <circle cx="8" cy="8" r="6" style="fill:${keyColor};">
            </reccit>
          </svg>
        </td>
        <td class="key">${escapeHTML(label)}</td>
        <td class="value">${escapeHTML(
          formatMeasureTooltipValue(val, measureValueFormatter),
        )}</td>
      </tr>`;
  }

  return content;
}

function generateBarContent(
  isDimensional: boolean,
  dimensionLabel: string | undefined,
  measureLabel: string,
  rest: Record<string, any>,
  colorMap: Record<string, string>,
  isTimeComparison: boolean,
  measureValueFormatter?: (value: number) => string | number | null | undefined,
) {
  let content = "";
  const key =
    isDimensional && dimensionLabel ? rest[dimensionLabel] : measureLabel;

  let keyColor = colorMap[measureLabel];

  if (isDimensional && dimensionLabel) {
    keyColor = colorMap[String(key)];
  } else if (isTimeComparison) {
    if (rest["Comparing"] === "ts") {
      keyColor = colorMap["current_period"];
    } else {
      keyColor = colorMap["comparison_period"];
    }
  }

  const val = rest[measureLabel];

  content += `
  <tr>
  <td class="color">
    <svg width="16" height="16">
      <circle cx="8" cy="8" r="6" style="fill:${keyColor};">
      </reccit>
    </svg>
  </td>
    <td class="key">${escapeHTML(key)}</td>
    <td class="value">${escapeHTML(
      formatMeasureTooltipValue(val, measureValueFormatter),
    )}</td>
  </tr>`;

  return content;
}

function generateTableContent(
  rest: Record<string, any>,
  colorMap: Record<string, string>,
  chartType: TDDAlternateCharts,
  measureLabel: string,
  dimensionLabel: string | undefined,
  isTimeComparison: boolean,
  selectedDimensionValues: (string | null)[],
  measureValueFormatter?: (value: number) => string | number | null | undefined,
): string {
  const keys = Object.keys(rest);
  let content = "";

  if (keys.length > 0) {
    content += "<table>";

    const selectedValuesLength = !!selectedDimensionValues.length;
    const isNonDimensionalTimeComparison =
      isTimeComparison && !selectedValuesLength;
    if (
      chartType === TDDChart.STACKED_AREA &&
      (selectedValuesLength || isTimeComparison)
    ) {
      content += generateStackedAreaContent(
        keys,
        rest,
        colorMap,
        isNonDimensionalTimeComparison,
        measureValueFormatter,
      );
    } else {
      content += generateBarContent(
        selectedValuesLength,
        dimensionLabel,
        measureLabel,
        rest,
        colorMap,
        isTimeComparison,
        measureValueFormatter,
      );
    }
    content += `</table>`;
  }

  return content;
}

function generateTooltipContent(
  value: Record<string, any>,
  colorMap: Record<string, string>,
  interval: V1TimeGrain | undefined,
  chartType: TDDAlternateCharts,
  measureLabel: string,
  dimensionLabel: string | undefined,
  isTimeComparison: boolean,
  selectedDimensionValues: (string | null)[],
  measureValueFormatter?: (value: number) => string | number | null | undefined,
): string {
  let content = "";
  const { Time, ...rest } = value;

  if (Time) {
    content += formatTimeHeader(Time, interval);
  }

  content += generateTableContent(
    rest,
    colorMap,
    chartType,
    measureLabel,
    dimensionLabel,
    isTimeComparison,
    selectedDimensionValues,
    measureValueFormatter,
  );

  return content;
}

export function tddTooltipFormatter(
  chartType: TDDAlternateCharts,
  measureLabel: string,
  dimensionLabel: string | undefined,
  isTimeComparison: boolean,
  selectedDimensionValues: (string | null)[],
  interval: V1TimeGrain | undefined,
  measureValueFormatter?: (value: number) => string | number | null | undefined,
) {
  const colorMap = createColorMap(
    measureLabel,
    selectedDimensionValues,
    isTimeComparison,
  );

  return (value: Record<string, any>) =>
    generateTooltipContent(
      value,
      colorMap,
      interval,
      chartType,
      measureLabel,
      dimensionLabel,
      isTimeComparison,
      selectedDimensionValues,
      measureValueFormatter,
    );
}
