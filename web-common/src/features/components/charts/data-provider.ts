import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type { TimeAndFilterStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";
import type {
  ChartDataQuery,
  ChartDataResult,
  ChartDomainValues,
  ChartSpec,
  TimeDimensionDefinition,
} from "./types";
import { adjustDataForTimeZone, getFieldsByType } from "./util";
import type { CanvasEntity } from "../../canvas/stores/canvas-entity";
import { primary, secondary } from "../../themes/colors";

export interface ChartDataDependencies<T extends ChartSpec = ChartSpec> {
  config: T;
  chartDataQuery: ChartDataQuery;
  metricsView: MetricsViewSelectors;
  /** Theme colors (primary/secondary) - updates when theme changes */
  themeStore: CanvasEntity["theme"];
  timeAndFilterStore: Readable<TimeAndFilterStore>;
  /** Reactive theme mode (light/dark toggle) - used in canvas context */
  themeModeStore?: Readable<boolean>;
  /** Static theme mode flag - used in standalone chart context */
  isThemeModeDark?: boolean;
  getDomainValues: () => ChartDomainValues;
}

/**
 * Pure function to create a chart data store without canvas context dependency.
 * This can be used outside of canvas context by providing all dependencies explicitly.
 */
export function getChartData<T extends ChartSpec = ChartSpec>(
  deps: ChartDataDependencies<T>,
): Readable<ChartDataResult> {
  const {
    config,
    chartDataQuery,
    metricsView,
    themeStore,
    getDomainValues,
    timeAndFilterStore,
    themeModeStore,
    isThemeModeDark: staticThemeModeDark,
  } = deps;

  const { measures, dimensions, timeDimensions } = getFieldsByType(config);

  // Combine all fields with their types
  const allFields = [
    ...measures.map((field) => ({ field, type: "measure" })),
    ...dimensions.map((field) => ({ field, type: "dimension" })),
    ...timeDimensions.map((field) => ({ field, type: "time" })),
  ];

  const fieldReadableMap = allFields.map((field) => {
    if (field.type === "measure") {
      return metricsView.getMeasureForMetricView(
        field.field,
        config.metrics_view,
      );
    } else if (field.type === "dimension") {
      return metricsView.getDimensionForMetricView(
        field.field,
        config.metrics_view,
      );
    } else {
      return getTimeDimensionDefinition(field.field, timeAndFilterStore);
    }
  });

  // Use themeModeStore if provided (canvas context), otherwise create a static store from the flag
  const modeStore =
    themeModeStore || derived([], () => staticThemeModeDark ?? false);

  return derived(
    [
      chartDataQuery,
      timeAndFilterStore,
      themeStore,
      modeStore,
      ...fieldReadableMap,
    ],
    ([chartData, $timeAndFilterStore, theme, isThemeModeDark, ...fieldMap]) => {
      const fieldSpecMap = allFields.reduce(
        (acc, field, index) => {
          acc[field.field] = fieldMap?.[index];
          return acc;
        },
        {} as Record<
          string,
          | MetricsViewSpecMeasure
          | MetricsViewSpecDimension
          | TimeDimensionDefinition
          | undefined
        >,
      );

      let data = chartData?.data?.data;

      if (timeDimensions?.length && $timeAndFilterStore.timeGrain) {
        data = adjustDataForTimeZone(
          data,
          timeDimensions,
          $timeAndFilterStore.timeGrain,
          $timeAndFilterStore.timeRange.timeZone || "UTC",
        );
      }

      const domainValues = getDomainValues();

      return {
        data: data || [],
        isFetching: chartData?.isFetching ?? false,
        error: chartData?.error,
        fields: fieldSpecMap,
        domainValues,
        isDarkMode: isThemeModeDark,
        theme: {
          primary:
            theme?.colors?.[isThemeModeDark ? "dark" : "light"]?.primary ||
            primary["500"],
          secondary:
            theme?.colors?.[isThemeModeDark ? "dark" : "light"]?.secondary ||
            secondary["500"],
        },
      };
    },
  );
}

export function getTimeDimensionDefinition(
  field: string,
  timeAndFilterStore: Readable<TimeAndFilterStore>,
): Readable<TimeDimensionDefinition> {
  return derived(timeAndFilterStore, ($timeAndFilterStore) => {
    const grain = $timeAndFilterStore?.timeGrain;
    const displayName = "Time";

    if (grain) {
      const timeUnit = timeGrainToVegaTimeUnitMap[grain];
      const format = TIME_GRAIN[grain]?.d3format as string;
      return {
        field,
        timeUnit,
        displayName,
        format,
      };
    }
    return {
      field,
      displayName,
    };
  });
}
