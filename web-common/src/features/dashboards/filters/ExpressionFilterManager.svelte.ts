import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { type V1Expression } from "@rilldata/web-common/runtime-client";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { toStore, type Readable } from "svelte/store";
import {
  type DimensionOrMeasureManager,
  JoinerFilterManager,
} from "@rilldata/web-common/features/dashboards/filters/JoinerFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";
import type {
  FilterChangeSource,
  FilterEvents,
} from "@rilldata/web-common/features/dashboards/filters/filter-events.ts";
import { mergeFilterParams } from "@rilldata/web-common/features/dashboards/filters/expr-utils.ts";
import { getSortFilterManagers } from "@rilldata/web-common/features/dashboards/filters/get-sort-filter-managers.ts";
import { expandCompressedParams } from "@rilldata/web-common/features/dashboards/url-state/compression.ts";
import type { UrlParamsStore } from "@rilldata/web-common/lib/store-utils/url-params-store-sync.svelte.ts";

/**
 * Filter managers for the chips in a filter bar.
 *
 * The managers are derived from the filter param, so they follow the param, the metrics view specs
 * and the pinned/required config as each of them arrives or changes. Filters the param does not
 * cover get a manager as well: required and pinned filters that have no value yet, the filter added
 * from the filter menu, and the ones created by a chart interaction.
 */
export class ExpressionFilterManager implements UrlParamsStore {
  // True when the param cannot be represented as chips.
  public isComplexFilter = $state(false);
  public exprByMetricsView: Record<string, V1Expression>;
  // Temporary svelte4 store for legacy components
  public exprByMetricsViewStore: Readable<Record<string, V1Expression>>;
  public inList: string[];
  public hasSomeFilter: boolean;

  // Name of the filter added from the filter menu. Its chip opens as soon as it is rendered.
  public temporaryFilterName: string | undefined = $state(undefined);
  // Manager for the whole filter. Will contain parsed and validated expressions per metrics view.
  public topLevelJoiner: JoinerFilterManager;
  public sortedFilterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };
  public readonly filterManagersMap: Record<string, DimensionOrMeasureManager>;

  public curSetParams = $state(new URLSearchParams());

  // Shared with every manager below this one, so a chip edit reports itself
  // without the managers in between having to forward it. See `filter-events.ts`.
  private events = new EventEmitter<FilterEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  // Temporary lock in explore. Once we move whereFilter out of explore, we can remove this.
  public updating = false;

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
    private readonly singleParamFormMv = false,
  ) {
    this.topLevelJoiner = $state(
      JoinerFilterManager.parse(
        this.metricsViewsProvider,
        createAndExpression([]),
        [],
        this.yamlConfigProvider,
        this.events,
      ) as JoinerFilterManager,
    );

    this.sortedFilterManagers = $derived.by(() =>
      getSortFilterManagers(this.topLevelJoiner, this.yamlConfigProvider),
    );
    this.inList = $derived(
      this.sortedFilterManagers.dimensions
        .filter((dfm) => dfm.mode === DimensionFilterMode.InList)
        .map((dfm) => dfm.name),
    );
    this.filterManagersMap = $derived.by(() =>
      Object.fromEntries([
        ...this.sortedFilterManagers.dimensions.map(
          (dfm) => [dfm.name, dfm] as [string, DimensionFilterManager],
        ),
        ...this.sortedFilterManagers.measures.map(
          (mfm) => [mfm.name, mfm] as [string, MeasureFilterManager],
        ),
      ]),
    );

    this.exprByMetricsView = $derived(this.topLevelJoiner.expr);
    this.exprByMetricsViewStore = toStore(() => this.exprByMetricsView);
    this.hasSomeFilter = $derived(
      Object.keys(this.exprByMetricsView).length > 0,
    );
  }

  public createListener() {}

  public clone() {
    const cloned = new ExpressionFilterManager(
      this.metricsViewsProvider,
      this.yamlConfigProvider,
    );

    const newUrlSearch = new URLSearchParams();
    this.applyFilterToParams(newUrlSearch);
    cloned.setUrlParams(newUrlSearch);

    return cloned;
  }

  public setUrlParams(searchParams: URLSearchParams) {
    const expandedUrlParams = expandCompressedParams(searchParams);

    const relevantUrlParams = new URLSearchParams();
    expandedUrlParams.forEach((value, key) => {
      if (
        key === ExploreStateURLParams.Filters ||
        key.startsWith(ExploreStateURLParams.Filters + ".")
      ) {
        relevantUrlParams.append(key, value);
      }
    });

    const { expr, inList, advanced } = mergeFilterParams(relevantUrlParams);

    this.temporaryFilterName = undefined;
    this.topLevelJoiner = JoinerFilterManager.parse(
      this.metricsViewsProvider,
      maybeWrapAndExpression(expr) ?? createAndExpression([]),
      inList,
      this.yamlConfigProvider,
      this.events,
    ) as JoinerFilterManager;
    this.isComplexFilter = advanced;

    this.curSetParams = normalizeUrlParams(
      relevantUrlParams,
      this.metricsViewsProvider.metricsViewNames,
      this.singleParamFormMv,
    );
  }

  public setParamForMetricsView(mvName: string, param: string) {
    const paramKey = getParamKeyForMv(mvName, this.singleParamFormMv);
    const paramsAreEqual = this.curSetParams.get(paramKey) === param;
    if (paramsAreEqual) return;

    const newParams = new URLSearchParams(this.curSetParams);
    newParams.set(paramKey, param);
    this.setUrlParams(newParams);
  }

  public setExprForMetricsView(
    mvName: string,
    expr: V1Expression | undefined,
    dimensionsWithInlistFilter?: string[],
  ) {
    this.setParamForMetricsView(
      mvName,
      expr
        ? convertExpressionToFilterParam(expr, dimensionsWithInlistFilter)
        : "",
    );
  }

  public applyFilterToParams(searchParams: URLSearchParams) {
    Object.entries(this.topLevelJoiner.param).forEach(([key, value]) => {
      const paramKey = getParamKeyForMv(key, this.singleParamFormMv);
      if (value) {
        searchParams.set(paramKey, value);
      } else {
        searchParams.delete(paramKey);
      }
    });

    // A singular `f=...` is a legacy canvas filter. `setUrlParams` folds it into every metrics view,
    // so the per metrics view params written above already carry it and it can be dropped.
    if (!this.singleParamFormMv && this.metricsViewsProvider.ready) {
      searchParams.delete(ExploreStateURLParams.Filters);
    }
  }

  /** Adds a chip for a filter that has no value yet. The chip opens as soon as it is rendered. */
  public addNewFilter(name: string) {
    if (this.filterManagersMap[name]) return;

    if (this.metricsViewsProvider.dimensionSpecs[name]) {
      const dfm = DimensionFilterManager.createForMetricsViews(
        this.metricsViewsProvider,
        name,
        { events: this.events },
      );
      if (!dfm) return;

      this.topLevelJoiner.maybeAddDimensionFilter(dfm);
      this.temporaryFilterName = name;
    } else if (this.metricsViewsProvider.measureSpecs[name]) {
      const mfm = MeasureFilterManager.createForMetricsViews(
        this.metricsViewsProvider,
        name,
        { events: this.events },
      );
      if (!mfm) return;

      this.topLevelJoiner.maybeAddMeasureFilter(mfm);
      this.temporaryFilterName = name;
    }
  }

  // TODO: add return type based on callback type?
  public dimensionFilterAction(
    name: string,
    callback: (dimensionFilterManager: DimensionFilterManager) => any,
    // Anything other than the filter bar reaches the managers through here,
    // so this is where a caller names itself as the cause of the change.
    source?: FilterChangeSource,
  ) {
    const dimensionFilterManager =
      this.filterManagersMap[name] ??
      DimensionFilterManager.createForMetricsViews(
        this.metricsViewsProvider,
        name,
        { events: this.events },
      );
    if (!(dimensionFilterManager instanceof DimensionFilterManager)) return;

    const ret = dimensionFilterManager.runWithSource(source, () =>
      callback(dimensionFilterManager),
    );
    // Only add dimension filter if the action resulted in a valid expression
    if (dimensionFilterManager.expr) {
      this.topLevelJoiner.maybeAddDimensionFilter(dimensionFilterManager);
    }
    return ret;
  }

  public measureFilterAction(
    name: string,
    callback: (measureFilterManager: MeasureFilterManager) => any,
    source?: FilterChangeSource,
  ) {
    const measureFilterManager =
      this.filterManagersMap[name] ??
      MeasureFilterManager.createForMetricsViews(
        this.metricsViewsProvider,
        name,
        { events: this.events },
      );
    if (!(measureFilterManager instanceof MeasureFilterManager)) return;

    const ret = measureFilterManager.runWithSource(source, () =>
      callback(measureFilterManager),
    );
    // Only add measure filter if the action resulted in a valid expression
    if (measureFilterManager.expr) {
      this.topLevelJoiner.maybeAddMeasureFilter(measureFilterManager);
    }
    return ret;
  }

  public clear() {
    this.temporaryFilterName = undefined;
    this.topLevelJoiner.clear();
    // Clearing goes through the param rather than the managers, so it has to report itself.
    this.events.emit("filter-changed", { source: undefined });
  }

  /**
   * Filters that apply to `metricsViewName`, other than the one on `name`.
   * Used to narrow the values shown in a dimension filter by the rest of the filter bar.
   */
  public getOtherDimensionsFilter(name: string, metricsViewName: string) {
    const exprs = this.sortedFilterManagers.dimensions
      .filter(
        (dfm) =>
          !!dfm.expr &&
          dfm.name !== name &&
          !!this.metricsViewsProvider.dimensionSpecs[dfm.name]?.[
            metricsViewName
          ],
      )
      .map((d) => d.expr as V1Expression);
    return exprs.length === 0 ? undefined : createAndExpression(exprs);
  }

  public createLocalFilterStore(metricsViewName: string) {
    return new ExpressionFilterManager(
      new MetricsViewsProvider(this.metricsViewsProvider.runtimeClient, [
        metricsViewName,
      ]),
      this.yamlConfigProvider,
    );
  }

  public getExprStoreForFirstMetricsView() {
    // The name is read inside the store, since the metrics views only arrive once the specs load.
    return toStore(() => {
      const mvName = this.metricsViewsProvider.metricsViewNames[0];
      return {
        expr: this.topLevelJoiner.expr[mvName],
        dimensionOnlyExpr: this.topLevelJoiner.dimensionOnlyExpr[mvName],
      };
    });
  }

  public getExprStoreForMetricsView(mvName: string) {
    return toStore(() => ({
      expr: this.topLevelJoiner.expr[mvName],
      dimensionOnlyExpr: this.topLevelJoiner.dimensionOnlyExpr[mvName],
    }));
  }
}

function normalizeUrlParams(
  urlParams: URLSearchParams,
  metricsViews: string[],
  singleParamFormMv: boolean,
) {
  const singularParam = urlParams.get(ExploreStateURLParams.Filters);
  if (!singularParam || singleParamFormMv) return urlParams;

  const newUrlParams = new URLSearchParams();
  metricsViews.forEach((mvName) =>
    newUrlParams.set(
      getParamKeyForMv(mvName, singleParamFormMv),
      singularParam,
    ),
  );
  return newUrlParams;
}

export function getParamKeyForMv(mvName: string, singleParamFormMv: boolean) {
  return singleParamFormMv
    ? ExploreStateURLParams.Filters
    : `${ExploreStateURLParams.Filters}.${mvName}`;
}
