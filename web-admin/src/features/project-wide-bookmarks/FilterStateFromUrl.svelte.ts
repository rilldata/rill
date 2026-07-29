import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import {
  type V1Expression,
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { deriveInterval } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls.ts";
import { formatTimeRange } from "@rilldata/web-admin/features/bookmarks/utils.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { getCanvasStore } from "@rilldata/web-common/features/canvas/state-managers/state-managers.ts";
import { getFiltersFromText } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-search-text-utils.ts";
import { splitWhereFilter } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils.ts";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { DimensionThresholdFilter } from "@rilldata/web-common/features/dashboards/stores/explore-state.ts";
import type { UIFilters } from "@rilldata/web-common/features/canvas/stores/filter-manager.ts";
import { ExploreSpecProvider } from "@rilldata/web-common/features/dashboards/ExploreSpecProvider.svelte.ts";
import { CanvasSpecProvider } from "@rilldata/web-common/features/canvas/CanvasSpecProvider.svelte.ts";
import { getDashboardResourceFromPage } from "@rilldata/web-common/features/dashboards/nav-utils.ts";
import { page } from "$app/state";

export type BaseFilterState = {
  queryTimeStart: string;
  queryTimeEnd: string;
  displayTimeRange: V1TimeRange;
  selectedTimeRange: string;
};
export type ExploreSpecificFilterState = {
  filters: V1Expression;
  dimensionThresholdFilters: DimensionThresholdFilter[];
  dimensionsWithInlistFilter: [];
};
export type CanvasSpecificFilterState = {
  uiFilters: UIFilters;
};
export type FilterState = BaseFilterState &
  (ExploreSpecificFilterState | CanvasSpecificFilterState);

export class FilterStateFromUrl {
  public filterState = $state<FilterState | undefined>(undefined);

  private cleanup: (() => void) | undefined = undefined;
  private provider = $state<
    ExploreSpecProvider | CanvasSpecProvider | undefined
  >(undefined);
  private currentResource =
    $state<ReturnType<typeof getDashboardResourceFromPage>>(null);

  public constructor(private readonly runtimeClient: RuntimeClient) {
    $effect(() => {
      const resource = getDashboardResourceFromPage(page);
      if (resource === null) {
        if (this.currentResource !== null) {
          this.currentResource = null;
          this.provider = undefined;
          this.cleanup?.();
        }
        return null;
      }

      if (
        resource.kind === this.currentResource?.kind &&
        resource.name === this.currentResource?.name
      ) {
        return;
      }

      this.currentResource = resource;
      switch (resource.kind) {
        case ResourceKind.Explore:
          this.provider = new ExploreSpecProvider(
            this.runtimeClient,
            resource.name,
          );
          break;

        case ResourceKind.Canvas:
          this.provider = new CanvasSpecProvider(
            this.runtimeClient,
            resource.name,
          );
          break;

        default:
          this.provider = undefined;
          break;
      }
    });

    $effect(() => {
      if (!this.provider) {
        this.filterState = undefined;
        return;
      }

      const searchParamsObj = new URLSearchParams(page.url.searchParams);
      const rangeExpression = searchParamsObj.get(
        ExploreStateURLParams.TimeRange,
      );
      const timeRange = <V1TimeRange>{
        expression: rangeExpression || "",
      };

      const timeZone =
        searchParamsObj.get(ExploreStateURLParams.TimeZone) || "UTC";

      const promises = this.provider.metricsViewNames.map((mvName) =>
        deriveInterval(
          timeRange.expression || "",
          this.runtimeClient,
          mvName,
          timeZone,
        ),
      );

      void Promise.all(promises).then((intervals) => {
        let intervalWithLatestEndPoint:
          | Awaited<ReturnType<typeof deriveInterval>>
          | undefined;
        intervals.forEach((response) => {
          if (
            !intervalWithLatestEndPoint ||
            (response.interval.end && intervalWithLatestEndPoint.interval.end
              ? response.interval.end > intervalWithLatestEndPoint.interval.end
              : false)
          ) {
            intervalWithLatestEndPoint = response;
          }
        });

        const start = intervalWithLatestEndPoint.interval.start.toISO();
        const end = intervalWithLatestEndPoint.interval.end.toISO();

        const grain =
          (searchParamsObj.get(
            ExploreStateURLParams.TimeGrain,
          ) as V1TimeGrain) ||
          intervalWithLatestEndPoint.grain ||
          V1TimeGrain.TIME_GRAIN_MINUTE;

        const selectedTimeRange = formatTimeRange(start, end, grain, timeZone);

        if (this.currentResource?.kind === ResourceKind.Canvas) {
          const uiFilters = getCanvasStore(
            this.currentResource.name,
            this.runtimeClient.instanceId,
          ).canvasEntity.filterManager.getUIFiltersFromString(page.url.search);

          this.filterState = <BaseFilterState & CanvasSpecificFilterState>{
            uiFilters,
            queryTimeStart: start,
            queryTimeEnd: end,
            displayTimeRange: timeRange,
            selectedTimeRange,
          };
          return;
        }

        const { expr, dimensionsWithInlistFilter } = getFiltersFromText(
          searchParamsObj.get(ExploreStateURLParams.Filters) || "",
        );

        const { dimensionFilters, dimensionThresholdFilters } =
          splitWhereFilter(expr);

        this.filterState = <BaseFilterState & ExploreSpecificFilterState>{
          dimensionThresholdFilters,
          dimensionsWithInlistFilter,
          filters: dimensionFilters,
          queryTimeStart: start,
          queryTimeEnd: end,
          displayTimeRange: timeRange,
          selectedTimeRange,
        };
      });
    });
  }
}
