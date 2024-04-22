import { COMPARIONS_COLORS } from "@rilldata/web-common/features/dashboards/config";
import { MainLineColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import { TDDAlternateCharts, TDDChart } from "../types";

export function escapeHTML(value: any): string {
  return String(value).replace(/&/g, "&amp;").replace(/</g, "&lt;");
}

export function tddTooltipFormatter(
  chartType: TDDAlternateCharts,
  measureLabel: string,
  dimensionLabel: string | undefined,
  selectedDimensionValues: (string | null)[],
) {
  let colorMap: Record<string, string> = {};
  selectedDimensionValues.forEach((dimValue, i) => {
    colorMap[String(dimValue)] = COMPARIONS_COLORS[i];
  });
  colorMap[measureLabel] = MainLineColor;

  return (value: Record<string, any>) => {
    let content = "";
    const { Time, ...rest } = value;

    if (Time) {
      content += `<h2>${Time}</h2>`;
    }

    const keys = Object.keys(rest);

    if (keys.length > 0) {
      content += "<table>";

      const selectedValuesLength = selectedDimensionValues.length;
      if (chartType === TDDChart.STACKED_AREA && selectedValuesLength) {
        for (const key of keys) {
          let val = rest[key];

          if (val === null || val === "NaN") continue;

          content += `
            <tr>
              <td class="color">
                <svg width="16" height="16">
                  <circle cx="8" cy="8" r="6" style="fill:${colorMap[String(key)] || "#000"};">
                  </reccit>
                </svg>
              </td>
              <td class="key">${escapeHTML(key)}</td>
              <td class="value">${escapeHTML(String(val))}</td>
            </tr>`;
        }
      } else {
        const key =
          selectedValuesLength && dimensionLabel
            ? rest[dimensionLabel]
            : measureLabel;
        const val = rest[measureLabel];

        content += `
          <tr>
          <td class="color">
            <svg width="16" height="16">
              <circle cx="8" cy="8" r="6" style="fill:${colorMap[String(key)] || "#000"};">
              </reccit>
            </svg>
          </td>
            <td class="key">${escapeHTML(key)}</td>
            <td class="value">${escapeHTML(String(val))}</td>
          </tr>`;
      }
      content += `</table>`;
    }

    return content;
  };
}
