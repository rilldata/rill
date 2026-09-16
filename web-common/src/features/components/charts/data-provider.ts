import { timeGrainToVegaTimeUnitMap } from "@rilldata/web-common/components/vega/util";
import type { MetricsViewSelectors } from "@rilldata/web-common/features/metrics-views/metrics-view-selectors";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import {
  type MetricsViewSpecDimension,
  type MetricsViewSpecMeasure,
  type V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { derived, type Readable } from "svelte/store";
import type { CanvasEntity } from "../../canvas/stores/canvas-entity";
import { primary, secondary } from "../../themes/colors";
import type {
  ChartDataQuery,
  ChartDataResult,
  ChartDomainValues,
  ChartSpec,
  TimeDimensionDefinition,
} from "./types";
import { adjustDataForTimeZone, getFieldsByType } from "./util";
import type { ExpressionState } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";

export interface ChartDataDependencies<T extends ChartSpec = ChartSpec> {
  config: T;
  chartDataQuery: ChartDataQuery;
  metricsView: MetricsViewSelectors;
  /** Theme colors (primary/secondary) - updates when theme changes */
  themeStore: CanvasEntity["theme"];
  timeControlStore: Readable<TimeControlState>;
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
    timeControlStore,
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
      return getTimeDimensionDefinition(field.field, timeControlStore);
    }
  });

  // Use themeModeStore if provided (canvas context), otherwise create a static store from the flag
  const modeStore =
    themeModeStore || derived([], () => staticThemeModeDark ?? false);

  return derived(
    [
      chartDataQuery,
      timeControlStore,
      themeStore,
      modeStore,
      ...fieldReadableMap,
    ],
    ([chartData, $timeControlStore, theme, isThemeModeDark, ...fieldMap]) => {
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

      if (timeDimensions?.length && $timeControlStore.timeGrain) {
        data = adjustDataForTimeZone(
          data,
          timeDimensions,
          $timeControlStore.timeZone || "UTC",
        );
      }

      const domainValues = getDomainValues();
      const hasComparison = $timeControlStore.showComparison;
      const waitingForTimeState =
        $timeControlStore.hasTimeSeries === undefined ||
        ($timeControlStore.hasTimeSeries === true &&
          (!$timeControlStore.apiTimeRange?.start ||
            !$timeControlStore.apiTimeRange?.end));

      return {
        data: data || [],
        isFetching: waitingForTimeState || (chartData?.isFetching ?? false),
        error: chartData?.error,
        fields: fieldSpecMap,
        domainValues,
        isDarkMode: isThemeModeDark,
        hasComparison,
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
  timeControlStore: Readable<TimeControlState>,
): Readable<TimeDimensionDefinition> {
  return derived(timeControlStore, ($timeControlStore) => {
    const grain = $timeControlStore?.timeGrain;
    const displayName = "Time";

    if (grain) {
      const timeUnit = timeGrainToVegaTimeUnitMap[grain];
      const format = TIME_GRAIN[grain]?.d3format;
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

export function getFieldsForSpec<T extends ChartSpec = ChartSpec>(
  config: ChartDataDependencies<T>["config"],
  metricsView: V1MetricsViewSpec,
) {
  const { measures, dimensions, timeDimensions } = getFieldsByType(config);

  const fields = {} as Record<
    string,
    | MetricsViewSpecMeasure
    | MetricsViewSpecDimension
    | TimeDimensionDefinition
    | undefined
  >;

  measures.forEach((measure) => {
    fields[measure] = metricsView.measures?.find((m) => m.name === measure);
  });

  dimensions.forEach((dimension) => {
    fields[dimension] = metricsView.dimensions?.find(
      (d) => d.name === dimension,
    );
  });

  timeDimensions.forEach((timeDimension) => {
    fields[timeDimension] = {
      field: timeDimension,
      displayName: "Time",
    };
  });

  return fields;
}
