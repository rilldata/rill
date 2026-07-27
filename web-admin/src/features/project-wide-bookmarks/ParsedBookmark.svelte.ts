import type { V1Bookmark } from "@rilldata/web-admin/client";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { parseRillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import {
  createQueryServiceMetricsViewTimeRange,
  createRuntimeServiceListResources,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { getDashboardStateFromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto.ts";
import { convertPartialExploreStateToUrlParams } from "@rilldata/web-common/features/dashboards/url-state/convert-partial-explore-state-to-url-params.ts";
import { getTimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store.ts";
import { page } from "$app/state";
import { getRillDefaultExploreUrlParams } from "@rilldata/web-common/features/dashboards/url-state/get-rill-default-explore-url-params.ts";
import { cleanUrlParams } from "@rilldata/web-common/features/dashboards/url-state/clean-url-params.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";

interface BookmarkParser {
  // Url params directly converted from bookmark.
  // Default url params are cleaned to match the actual url.
  bookmarkUrlParams: URLSearchParams;

  cleanup: (() => void) | undefined;
}

export class DefaultBookmarkParser implements BookmarkParser {
  public bookmarkUrlParams: URLSearchParams;
  public cleanup: (() => void) | undefined = undefined;

  public constructor(bookmark: V1Bookmark) {
    this.bookmarkUrlParams = $state(
      new URLSearchParams(bookmark.urlSearch ?? ""),
    );
  }
}

export class ExploreBookmarkParser implements BookmarkParser {
  public bookmarkUrlParams: URLSearchParams;
  public cleanup: (() => void) | undefined = undefined;

  private exploreSpec = $state<V1ExploreSpec | undefined>(undefined);
  private metricsViewSpec = $state<V1MetricsViewSpec | undefined>(undefined);
  private timeRangeSummary = $state<V1TimeRangeSummary | undefined>(undefined);

  public constructor(
    private readonly runtimeClient: RuntimeClient,
    private readonly bookmark: V1Bookmark,
  ) {
    const allResourcesQuery = createRuntimeServiceListResources(
      this.runtimeClient,
      {},
    );

    let timeRangeSummaryUnsub: (() => void) | undefined = undefined;
    const allResourcesUnsub = allResourcesQuery.subscribe(
      (allResourcesResp) => {
        this.exploreSpec = allResourcesResp.data?.resources?.find(
          (r) => r.meta?.name?.name === this.bookmark.resourceName,
        )?.explore?.state?.validSpec;
        if (!this.exploreSpec?.metricsView) return;

        this.metricsViewSpec = allResourcesResp.data?.resources?.find(
          (r) => r.meta?.name?.name === this.exploreSpec.metricsView,
        )?.metricsView?.state?.validSpec;

        if (!timeRangeSummaryUnsub) {
          const timeRangeSummaryQuery = createQueryServiceMetricsViewTimeRange(
            this.runtimeClient,
            { metricsViewName: this.exploreSpec.metricsView },
          );
          timeRangeSummaryUnsub = timeRangeSummaryQuery.subscribe(
            (timeRangeSummaryResp) => {
              this.timeRangeSummary =
                timeRangeSummaryResp.data?.timeRangeSummary;
            },
          );
        }
      },
    );

    this.cleanup = () => {
      allResourcesUnsub();
      timeRangeSummaryUnsub?.();
    };

    this.bookmarkUrlParams = $derived.by(() => {
      let searchParams: URLSearchParams;
      if (this.bookmark.data) {
        const exploreStateFromBookmark = getDashboardStateFromUrl(
          this.bookmark.data,
          this.metricsViewSpec,
          this.exploreSpec,
        );

        // We need to check if the bookmark's url is equal to current url or not to show an "active" state.
        // To avoid calculating it everytime we directly convert it to final url.
        searchParams = convertPartialExploreStateToUrlParams(
          this.exploreSpec,
          this.metricsViewSpec,
          exploreStateFromBookmark,
          getTimeControlState(
            this.metricsViewSpec,
            this.exploreSpec,
            this.timeRangeSummary,
            exploreStateFromBookmark,
          ),
        );
      } else {
        searchParams = new URLSearchParams(this.bookmark.urlSearch ?? "");
      }

      const defaultUrlParams = getRillDefaultExploreUrlParams(
        this.metricsViewSpec,
        this.exploreSpec,
        this.timeRangeSummary,
      );
      return cleanUrlParams(searchParams, defaultUrlParams);
    });
  }
}

export class ParsedBookmark {
  // If bookmark has filter-only params then this is true. Non-filter params from the current url are retained.
  public filtersOnly: boolean;
  // If bookmark has an absolute time range. Start and end times will be hardcoded here.
  public absoluteTimeRange: boolean;
  public isActive: boolean;
  // Full url to be used to navigate.
  // This contains existing non-filter params on top of filter params for filters only bookmark.
  public fullUrl: string;

  public cleanup: (() => void) | undefined;

  public constructor(
    runtimeClient: RuntimeClient,
    public bookmark: V1Bookmark,
  ) {
    const parser =
      bookmark.resourceKind === ResourceKind.Explore
        ? new ExploreBookmarkParser(runtimeClient, bookmark)
        : new DefaultBookmarkParser(bookmark);

    this.filtersOnly = $derived(isFilterOnlyBookmark(parser.bookmarkUrlParams));
    this.absoluteTimeRange = $derived(
      isAbsoluteTimeRangeBookmark(parser.bookmarkUrlParams),
    );

    this.isActive = $derived.by(() => {
      if (this.filtersOnly) {
        return (
          parser.bookmarkUrlParams.toString() ===
          page.url.searchParams.toString()
        );
      }

      return [...parser.bookmarkUrlParams.entries()].every(([key, value]) => {
        const curValue = page.url.searchParams.get(key);
        return curValue === value;
      });
    });

    this.fullUrl = $derived.by(() => {
      const fullUrlParams = new URLSearchParams(parser.bookmarkUrlParams);
      if (this.filtersOnly) {
        page.url.searchParams.forEach((v, p) => {
          if (fullUrlParams.has(p)) return;
          fullUrlParams.set(p, v);
        });
      }

      const projectPath = `${page.params.organization}/${page.params.project}`;
      const resourceRoute =
        this.bookmark.resourceKind === ResourceKind.Explore
          ? "explore"
          : "canvas";
      const dashboardPath = `${resourceRoute}/${this.bookmark.resourceName ?? ""}`;

      return `/${projectPath}/${dashboardPath}?${fullUrlParams.toString()}`;
    });
  }
}

// These are the only parameters that are stored in a filter-only bookmark
const FILTER_ONLY_PARAMS = new Set([
  ExploreStateURLParams.Filters,
  ExploreStateURLParams.TimeRange,
  ExploreStateURLParams.TimeGrain,
]) as Set<string>;

function isFilterOnlyBookmark(bookmarkUrlParams: URLSearchParams): boolean {
  const urlParams = Array.from(bookmarkUrlParams.keys());
  return urlParams.every((param) => FILTER_ONLY_PARAMS.has(param));
}

function isAbsoluteTimeRangeBookmark(bookmarkUrlParams: URLSearchParams) {
  const timeRange = bookmarkUrlParams.get(ExploreStateURLParams.TimeRange);
  if (!timeRange) return false;

  try {
    const rt = parseRillTime(timeRange);
    return rt.isAbsoluteTime();
  } catch {
    return false;
  }
}
