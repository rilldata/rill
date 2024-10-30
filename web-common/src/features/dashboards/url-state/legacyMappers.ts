import { ExplorePresetDefaultChartType } from "@rilldata/web-common/features/dashboards/url-state/defaults";

const LegacyCharTypeToPresetChartType: Record<string, string> = {
  default: ExplorePresetDefaultChartType,
  grouped_bar: "bar",
  stacked_bar: "stacked_bar",
  stacked_area: "stacked_area",
};
export function mapLegacyChartType(chartType: string | undefined) {
  if (!chartType) {
    return ExplorePresetDefaultChartType;
  }
  return (
    LegacyCharTypeToPresetChartType[chartType] ?? ExplorePresetDefaultChartType
  );
}
