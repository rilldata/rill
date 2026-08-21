import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { type V1Expression } from "@rilldata/web-common/runtime-client";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import {
  type MetricsViewName,
  MetricsViewsProvider,
} from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { fromStore, toStore, type Readable } from "svelte/store";
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
import { untrack } from "svelte";
import { mergeFilterParams } from "@rilldata/web-common/features/dashboards/filters/expr-utils.ts";

/**
 * Filter managers for the chips in a filter bar.
 *
 * The managers are derived from the filter param, so they follow the param, the metrics view specs
 * and the pinned/required config as each of them arrives or changes. Filters the param does not
 * cover get a manager as well: required and pinned filters that have no value yet, the filter added
 * from the filter menu, and the ones created by a chart interaction.
 */
export class ExpressionFilterManager {
  // True when the param cannot be represented as chips.
  public isComplexFilter: boolean;
  public exprByMetricsView: Record<string, V1Expression>;
  // Temporary svelte4 store for legacy components
  public exprByMetricsViewStore: Readable<Record<string, V1Expression>>;
  public inList: string[];
  public hasSomeFilter: boolean;

  // Name of the filter added from the filter menu. Its chip opens as soon as it is rendered.
  public temporaryFilterName: string | undefined = $state(undefined);
  // Manager for the whole filter. Will contain parsed and validated expressions per metrics view.
  public topLevelJoiner: JoinerFilterManager;
  public filterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };
  public readonly filterManagersMap: Record<string, DimensionOrMeasureManager>;

  // Shared with every manager below this one, so a chip edit reports itself
  // without the managers in between having to forward it. See `filter-events.ts`.
  private events = new EventEmitter<FilterEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  // Raw filter param per metrics view.
  private paramByMetricsView = $state<Record<MetricsViewName, string>>({});
  private mergedExprFromParams: ReturnType<typeof mergeFilterParams>;
  // Filter param based on managers.
  public paramByManager: Record<MetricsViewName, string>;
  // The param the url last seeded the managers with. See `writeBackToUrl`.
  private lastUrlParam = $state<Record<MetricsViewName, string>>({});

  // Temporary lock in explore. Once we move whereFilter out of explore, we can remove this.
  public updating = false;

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
  ) {
    this.mergedExprFromParams = $derived(
      mergeFilterParams(this.paramByMetricsView),
    );
    this.topLevelJoiner = $derived.by(() => {
      return JoinerFilterManager.parse(
        this.metricsViewsProvider,
        maybeWrapAndExpression(this.mergedExprFromParams.expr) ??
          createAndExpression([]),
        this.mergedExprFromParams.inList,
        this.yamlConfigProvider,
        this.events,
      ) as JoinerFilterManager;
    });
    this.paramByManager = $derived(
      Object.fromEntries(
        this.metricsViewsProvider.metricsViewNames.map((mvName) => [
          mvName,
          this.topLevelJoiner.param[mvName] ?? "",
        ]),
      ),
    );
    this.isComplexFilter = $derived(this.mergedExprFromParams.advanced);

    this.filterManagers = $derived.by(() =>
      this.buildFilterManagers(this.topLevelJoiner),
    );
    this.inList = $derived(
      this.filterManagers.dimensions
        .filter((dfm) => dfm.mode === DimensionFilterMode.InList)
        .map((dfm) => dfm.name),
    );
    this.filterManagersMap = $derived.by(() =>
      Object.fromEntries([
        ...this.filterManagers.dimensions.map(
          (dfm) => [dfm.name, dfm] as [string, DimensionFilterManager],
        ),
        ...this.filterManagers.measures.map(
          (mfm) => [mfm.name, mfm] as [string, MeasureFilterManager],
        ),
      ]),
    );

    this.exprByMetricsView = $derived(this.topLevelJoiner.expr);
    this.exprByMetricsViewStore = toStore(() => this.exprByMetricsView);
    this.hasSomeFilter = $derived(
      Object.keys(this.exprByMetricsView).length > 0,
    );

    // Init with empty url params
    this.setUrlParams(new URLSearchParams(), true);
  }

  public createListener() {
    // The param the listeners were last told about.
    // Changes are diffed against this rather than against `paramByMetricsView`, since a change that
    // goes through the param itself, `clear()` for example, moves both records at once.
    let emittedParam: Record<MetricsViewName, string> | undefined;

    $effect(() => {
      const param = { ...this.paramByManager };
      const metricsViewNames = this.metricsViewsProvider.metricsViewNames;

      // The managers only cover the metrics views once the specs have loaded and a param has been applied.
      // Reporting the empty state before that would clear the filter the url holds.
      if (
        Object.keys(param).length !== metricsViewNames.length ||
        !this.metricsViewsProvider.ready
      )
        return;

      const unchanged =
        !emittedParam || recordsMatch(emittedParam, param, metricsViewNames);
      emittedParam = param;
      if (unchanged) return;

      this.events.emit("state-changed");
    });
  }

  /**
   * Seeds the managers from a url the caller owns, and re-seeds them as the metrics view names arrive.
   *
   * The effect root is created here because the owner is not a component:
   * everything that reads a filter, a canvas component for example, has to see the url's filters
   * whether or not a filter bar is mounted. The caller owns the returned disposer.
   */
  public seedFromSearchParams(searchParams: Readable<URLSearchParams>) {
    return $effect.root(() => {
      // Subscribed inside the root so the disposer takes the subscription with it.
      const params = fromStore(searchParams);
      $effect(() => this.setUrlParams(params.current));
    });
  }

  /**
   * Writes what the managers hold back to the url.
   *
   * Diffed against the param the url last seeded rather than against the live url: a url the seed has
   * not caught up with yet, a back navigation for example, would otherwise read as a manager change
   * and navigate straight back to the state the managers still hold. Changes that bypass the chips,
   * `clear()` for example, leave the seed untouched and so are still written.
   */
  public writeBackToUrl(getUrl: () => URL, navigate: (url: URL) => void) {
    $effect(() => {
      // Every leaf expression is dropped until the specs land, so writing back before that would
      // strip the filter the url holds and lose it for good.
      if (!this.metricsViewsProvider.ready) return;

      const seeded = recordsMatch(
        this.lastUrlParam,
        this.paramByManager,
        this.metricsViewsProvider.metricsViewNames,
      );
      if (seeded) return;

      const newUrl = new URL(getUrl());
      this.applyFilterToParams(newUrl.searchParams);
      if (newUrl.search === getUrl().search) return;

      navigate(newUrl);
    });
  }

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

  public setUrlParams = (searchParams: URLSearchParams, force = false) => {
    const newParamByMetricsView = this.paramFromSearchParams(searchParams);
    // Recorded even when the param itself does not change, since `writeBackToUrl` diffs against it.
    this.lastUrlParam = newParamByMetricsView;
    this.applyParam(newParamByMetricsView, force);
  };

  /**
   * Folds what the managers hold back into the param.
   * For managers that are not synced with a url; the ones that are go through the url instead.
   */
  public foldManagersIntoParam() {
    const searchParams = new URLSearchParams();
    this.applyFilterToParams(searchParams);
    this.applyParam(this.paramFromSearchParams(searchParams), false);
  }

  private paramFromSearchParams(searchParams: URLSearchParams) {
    const newParamByMetricsView = {} as Record<MetricsViewName, string>;

    const singularFilter = searchParams.get(ExploreStateURLParams.Filters);
    // Tracked on purpose: the params are keyed by metrics view,
    // so they have to be re-applied once the names arrive.
    this.metricsViewsProvider.metricsViewNames.forEach((name: string) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${name}`;
      newParamByMetricsView[name] =
        searchParams.get(paramKey) ?? singularFilter ?? "";
    });

    return newParamByMetricsView;
  }

  private applyParam(
    newParamByMetricsView: Record<MetricsViewName, string>,
    force: boolean,
  ) {
    const metricsViewNames = this.metricsViewsProvider.metricsViewNames;

    // Managers can be mutated in place without going through the param,
    // so the new param also has to match what the managers hold.
    // Otherwise, a param that reverts such a mutation is a no-op.
    //
    // Read untracked: an effect feeding the url into this method must not depend on the state the method writes,
    // or every manager edit re-runs it and reverts itself back to the url.
    const unchanged = untrack(
      () =>
        recordsMatch(
          this.paramByMetricsView,
          newParamByMetricsView,
          metricsViewNames,
        ) &&
        recordsMatch(
          this.paramByManager,
          newParamByMetricsView,
          metricsViewNames,
        ),
    );
    if (unchanged && !force) return;

    this.paramByMetricsView = newParamByMetricsView;
    this.temporaryFilterName = undefined;
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

  public setParamForMetricsView(mvName: string, param: string) {
    if (this.paramByMetricsView[mvName] === param) return;

    this.paramByMetricsView = { ...this.paramByMetricsView, [mvName]: param };
    this.temporaryFilterName = undefined;
  }

  public applyFilterToParams(
    searchParams: URLSearchParams,
    singleParam: boolean = false,
  ) {
    Object.entries(this.paramByManager).forEach(([key, value]) => {
      const paramKey = singleParam
        ? ExploreStateURLParams.Filters
        : `${ExploreStateURLParams.Filters}.${key}`;
      if (value) {
        searchParams.set(paramKey, value);
      } else {
        searchParams.delete(paramKey);
      }
    });

    // A singular `f=...` is a legacy canvas filter. `setUrlParams` folds it into every metrics view,
    // so the per metrics view params written above already carry it and it can be dropped.
    // Until the names arrive there is nothing to fold it into, so it has to stay.
    const metricsViewsLoaded =
      this.metricsViewsProvider.metricsViewNames.length > 0 &&
      this.metricsViewsProvider.ready;
    if (!singleParam && metricsViewsLoaded) {
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
    // Applied to the param directly rather than through `setUrlParams`: clearing does not come from
    // the url, so recording it as the url's state would hide it from `writeBackToUrl`.
    this.applyParam(
      Object.fromEntries(
        this.metricsViewsProvider.metricsViewNames.map((name) => [name, ""]),
      ),
      false,
    );
    // Clearing goes through the param rather than the managers, so it has to report itself.
    this.events.emit("filter-changed", { source: undefined });
  }

  /**
   * Filters that apply to `metricsViewName`, other than the one on `name`.
   * Used to narrow the values shown in a dimension filter by the rest of the filter bar.
   */
  public getOtherDimensionsFilter(name: string, metricsViewName: string) {
    const exprs = this.filterManagers.dimensions
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

  /**
   * Managers in chip order: required filters first, then pinned ones, then the rest,
   * sorted by name within each of those groups.
   */
  private buildFilterManagers(joinerManager: JoinerFilterManager) {
    const rank = (manager: DimensionFilterManager | MeasureFilterManager) => {
      if (this.yamlConfigProvider.requiredFilters[manager.name]) return 0;
      if (this.yamlConfigProvider.pinnedFilters[manager.name]) return 1;
      return 2;
    };

    const sortFunc = (
      a: DimensionFilterManager | MeasureFilterManager,
      b: DimensionFilterManager | MeasureFilterManager,
    ) => rank(a) - rank(b) || a.name.localeCompare(b.name);

    // Sorted on a copy: the managers keep param order, which is the order the expression and the
    // param it writes back are built in.
    const dimensions = [...joinerManager.managers.dimensionManagers].sort(
      sortFunc,
    );
    const measures = [...joinerManager.managers.measureManagers].sort(sortFunc);

    return { dimensions, measures };
  }
}

export function recordsMatch(
  rec1: Record<string, string>,
  rec2: Record<string, string>,
  keys: string[],
) {
  const keysMatch =
    keys.length === Object.keys(rec1).length &&
    keys.length === Object.keys(rec2).length;
  if (!keysMatch) return false;
  return keys.every((key) => rec1[key] === rec2[key]);
}
