import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { type V1Expression } from "@rilldata/web-common/runtime-client";
import {
  convertExpressionToFilterParam,
  convertFilterParamToExpression,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import {
  type MetricsViewName,
  MetricsViewsProvider,
} from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { toStore, type Readable } from "svelte/store";
import {
  type AnyManager,
  type DimensionOrMeasureManager,
  MetricsViewFilterManager,
} from "@rilldata/web-common/features/dashboards/filters/MetricsViewFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

type ExpressionFilterManagerEvents = {
  "state-changed": void;
};

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
  // Filter manager per metrics view. Will contain parsed and validated expression.
  public managerByMetricsView: Record<
    MetricsViewName,
    MetricsViewFilterManager
  >;
  public filterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };
  public readonly filterManagersMap: Record<
    string,
    DimensionOrMeasureManager[]
  >;

  private events = new EventEmitter<ExpressionFilterManagerEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  // Raw filter param per metrics view.
  private paramByMetricsView = $state<Record<MetricsViewName, string>>({});
  // Filter param based on managers.
  private paramByManager: Record<MetricsViewName, string>;

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
  ) {
    this.managerByMetricsView = $derived.by(() => {
      console.log("managerByMetricsView calc");
      const existingManagers = new Map<string, AnyManager>();
      return Object.fromEntries(
        Object.entries(this.paramByMetricsView)
          .map(([mvName, param]) => {
            const { expr, dimensionsWithInlistFilter } =
              parseFilterParam(param);
            return [
              mvName,
              MetricsViewFilterManager.parse(
                this.metricsViewsProvider,
                mvName,
                maybeWrapAndExpression(expr) ?? createAndExpression([]),
                dimensionsWithInlistFilter,
                existingManagers,
              ),
            ];
          })
          .filter(([, manager]) => !!manager) as [
          string,
          MetricsViewFilterManager,
        ][],
      );
    });
    this.paramByManager = $derived(
      Object.fromEntries(
        Object.entries(this.managerByMetricsView).map(([mvName, manager]) => [
          mvName,
          manager.param,
        ]),
      ),
    );
    this.isComplexFilter = $derived(
      Object.values(this.managerByMetricsView).some((m) => m.isComplexFilter),
    );

    this.filterManagers = $derived.by(() =>
      this.buildFilterManagers(this.buildParamFilterManagers()),
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

    this.exprByMetricsView = $derived(
      Object.fromEntries(
        Object.entries(this.managerByMetricsView)
          .map(([mvName, manager]) => [mvName, manager.expr])
          .filter(([, manager]) => !!manager) as [string, V1Expression][],
      ),
    );
    this.exprByMetricsViewStore = toStore(() => this.exprByMetricsView);
    this.hasSomeFilter = $derived(
      Object.keys(this.exprByMetricsView).length > 0,
    );

    // Init with empty url params
    this.setUrlParams(new URLSearchParams(), true);
  }

  public createListener() {
    $effect(() => {
      const paramsAreEqual = recordsMatch(
        this.paramByManager,
        this.paramByMetricsView,
        this.metricsViewsProvider.metricsViewNames,
      );
      if (!paramsAreEqual) {
        this.events.emit("state-changed");
      }
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
    const newParamByMetricsView = {} as Record<MetricsViewName, string>;

    const singularFilter = searchParams.get(ExploreStateURLParams.Filters);
    this.metricsViewsProvider.metricsViewNames.forEach((name: string) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${name}`;
      newParamByMetricsView[name] =
        searchParams.get(paramKey) ?? singularFilter ?? "";
    });

    // Managers can be mutated in place without going through the param, so the new param also has
    // to match what the managers hold. Otherwise a param that reverts such a mutation is a no-op.
    const unchanged =
      recordsMatch(
        this.paramByMetricsView,
        newParamByMetricsView,
        this.metricsViewsProvider.metricsViewNames,
      ) &&
      recordsMatch(
        this.paramByManager,
        newParamByMetricsView,
        this.metricsViewsProvider.metricsViewNames,
      );
    if (unchanged && !force) return;

    this.paramByMetricsView = newParamByMetricsView;
    this.temporaryFilterName = undefined;
  };

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
  }

  /** Adds a chip for a filter that has no value yet. The chip opens as soon as it is rendered. */
  public addNewFilter(name: string) {
    if (this.filterManagersMap[name]) return;

    if (this.metricsViewsProvider.dimensionSpecs[name]) {
      const dfm = this.createDimensionFilterManager(name);
      console.log("Adding dimension filter for", name);
      if (!dfm) return;
      Object.keys(this.metricsViewsProvider.dimensionSpecs[name]).forEach(
        (mv) => this.managerByMetricsView[mv]?.maybeAddDimensionFilter(dfm),
      );
      this.temporaryFilterName = name;
    } else if (this.metricsViewsProvider.measureSpecs[name]) {
      const mfm = this.createMeasureFilterManager(name);
      console.log("Adding measure filter for", name);
      if (!mfm) return;
      Object.keys(this.metricsViewsProvider.measureSpecs[name]).forEach((mv) =>
        this.managerByMetricsView[mv]?.maybeAddMeasureFilter(mfm),
      );
      this.temporaryFilterName = name;
    }
  }

  public dimensionFilterAction(
    name: string,
    callback: (dimensionFilterManager: DimensionFilterManager) => any,
  ) {
    const dimensionFilterManager =
      this.filterManagersMap[name] ?? this.createDimensionFilterManager(name);
    if (!(dimensionFilterManager instanceof DimensionFilterManager)) return;

    const ret = callback(dimensionFilterManager);
    // Only add dimension filter if the action resulted in a valid expression
    if (dimensionFilterManager.expr) {
      Object.values(this.managerByMetricsView).forEach((m) =>
        m.maybeAddDimensionFilter(dimensionFilterManager),
      );
    }
    return ret;
  }

  public measureFilterAction(
    name: string,
    callback: (measureFilterManager: MeasureFilterManager) => any,
  ) {
    const measureFilterManager =
      this.filterManagersMap[name] ?? this.createMeasureFilterManager(name);
    if (!(measureFilterManager instanceof MeasureFilterManager)) return;

    const ret = callback(measureFilterManager);
    // Only add measure filter if the action resulted in a valid expression
    if (measureFilterManager.expr) {
      Object.values(this.managerByMetricsView).forEach((m) =>
        m.maybeAddMeasureFilter(measureFilterManager),
      );
    }
    return ret;
  }

  public clear() {
    this.temporaryFilterName = undefined;
    this.setUrlParams(new URLSearchParams());
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

  public createDimensionOrMeasureManagerForName(name: string) {
    return (
      this.createDimensionFilterManager(name) ??
      this.createMeasureFilterManager(name)
    );
  }

  public createDimensionOrMeasureManagerForExpr(
    expr: V1Expression,
    dimensionsWithInlistFilter: string[] = [],
  ) {
    const ident = expr?.cond?.exprs?.[0]?.ident ?? "";
    // Both filter types are keyed off a dimension: a dimension filter by the dimension it filters
    // and a measure filter by the dimension its subquery groups by.
    if (!this.metricsViewsProvider.dimensionSpecs[ident]) return undefined;

    const firstValueExpr = expr?.cond?.exprs?.[1];

    if (firstValueExpr?.subquery) {
      // Having a subquery means this is a measure filter.
      const measureName = firstValueExpr.subquery.measures?.[0];
      if (!measureName) return undefined;

      return this.createMeasureFilterManager(measureName, firstValueExpr);
    } else {
      // Everything else is a dimension filter for now.
      return this.createDimensionFilterManager(
        ident,
        expr,
        dimensionsWithInlistFilter.includes(ident),
      );
    }
  }

  public getExprStoreForFirstMetricsView() {
    return toStore(() => {
      const exprs = Object.values(this.managerByMetricsView)[0];
      return {
        expr: exprs?.expr,
        dimensionOnlyExpr: exprs?.dimensionOnlyExpr,
      };
    });
  }

  public getExprStoreForMetricsView(mvName: string) {
    return toStore(() => {
      const mvManager = this.managerByMetricsView[mvName];
      if (!mvManager) return undefined;
      return {
        expr: mvManager.expr,
        dimensionOnlyExpr: mvManager.dimensionOnlyExpr,
      };
    });
  }

  /** Managers for the filters in the param, keyed by name and in param order. */
  private buildParamFilterManagers() {
    const dimensionsMap = {} as Record<string, DimensionFilterManager[]>;
    const measuresMap = {} as Record<string, MeasureFilterManager[]>;

    // Go through exprs of parsed url params for each metrics views.
    Object.entries(this.managerByMetricsView).forEach(([, mvManager]) => {
      mvManager.managers.dimensionManagers.forEach((manager) => {
        dimensionsMap[manager.name] ??= [];
        dimensionsMap[manager.name].push(manager);
      });
      mvManager.managers.measureManagers.forEach((manager) => {
        measuresMap[manager.name] ??= [];
        measuresMap[manager.name].push(manager);
      });
    });

    return { measuresMap, dimensionsMap };
  }

  /**
   * Managers in chip order: required filters first, then pinned ones, then the rest of the param
   * filters in param order, and finally the filters the param does not cover yet.
   */
  private buildFilterManagers({
    dimensionsMap,
    measuresMap,
  }: {
    dimensionsMap: Record<string, DimensionFilterManager[]>;
    measuresMap: Record<string, MeasureFilterManager[]>;
  }) {
    const dimensions: DimensionFilterManager[] = [];
    const measures: MeasureFilterManager[] = [];
    const added = new Set<string>();

    const add = (name: string) => {
      if (added.has(name)) return;

      const filterManager =
        dimensionsMap[name]?.[0] ??
        measuresMap[name]?.[0] ??
        this.createDimensionOrMeasureManagerForName(name);
      if (!filterManager) return;

      added.add(name);
      if (filterManager instanceof MeasureFilterManager) {
        measures.push(filterManager);
      } else {
        dimensions.push(filterManager);
      }
    };

    // Required and pinned filters lead the chips and have a chip even without a value.
    const { requiredFilters, pinnedFilters } = this.yamlConfigProvider;
    Object.keys(requiredFilters).forEach((name) => add(name));
    Object.keys(pinnedFilters).forEach((name) => add(name));

    // Next comes dimensions from params
    sortFilterIterator(Object.keys(dimensionsMap)).forEach((name) => add(name));

    // Next comes measures from params
    sortFilterIterator(Object.keys(measuresMap)).forEach((name) => add(name));

    return { dimensions, measures };
  }

  private createDimensionFilterManager(
    name: string,
    initExpr?: V1Expression,
    isInList?: boolean,
  ) {
    const dimensionSpecs = this.metricsViewsProvider.dimensionSpecs[name];
    if (!dimensionSpecs) return undefined;

    return new DimensionFilterManager(
      name,
      getDimensionDisplayName(Object.values(dimensionSpecs)[0]),
      initExpr,
      isInList,
    );
  }

  private createMeasureFilterManager(name: string, initExpr?: V1Expression) {
    const measureSpecs = this.metricsViewsProvider.measureSpecs[name];
    if (!measureSpecs) return undefined;

    return new MeasureFilterManager(
      name,
      getMeasureDisplayName(Object.values(measureSpecs)[0]),
      initExpr,
    );
  }
}

/**
 * The parser throws on a malformed param.
 * Converting the url to explore state reports it as an error, so drop it here instead of throwing.
 */
function parseFilterParam(param: string) {
  try {
    return convertFilterParamToExpression(param);
  } catch {
    return { expr: undefined, dimensionsWithInlistFilter: [] };
  }
}

function sortFilterIterator(arr: string[]) {
  return [...arr].sort((a, b) => (a > b ? 1 : -1));
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
