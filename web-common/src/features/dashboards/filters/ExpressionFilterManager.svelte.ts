import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilterManager.svelte.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import {
  convertExpressionToFilterParam,
  convertFilterParamToExpression,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  isExpressionUnsupported,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilterManager.svelte.ts";
import {
  type MetricsViewName,
  MetricsViewsProvider,
} from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";
import { writable, type Writable } from "svelte/store";
import {
  type AnyManager,
  type DimensionOrMeasureManager,
  MetricsViewFilterManager,
} from "@rilldata/web-common/features/dashboards/filters/MetricsViewFilterManager.svelte.ts";

type ParsedUrlParamsByMV = Record<
  MetricsViewName,
  ReturnType<typeof convertFilterParamToExpression>
>;

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
  public exprByMetricsViewStore: Writable<Record<string, V1Expression>> =
    writable({});
  public hasSomeFilter: boolean;

  // Name of the filter added from the filter menu. Its chip opens as soon as it is rendered.
  public temporaryFilterName: string | undefined = $state(undefined);
  public managerByMetricsView: Record<string, MetricsViewFilterManager>;
  public filterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };
  public readonly filterManagersMap: Record<
    string,
    DimensionOrMeasureManager[]
  >;

  // Raw parsed expression from url params. Validated expression is stored in exprByMetricsView.
  private prevUrlParams = $state("");
  private parsedUrlParams = $state<ParsedUrlParamsByMV>({});

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
  ) {
    this.managerByMetricsView = $derived.by(() => {
      const existingManagers = new Map<string, AnyManager>();
      return Object.fromEntries(
        Object.entries(this.parsedUrlParams)
          .map(([mvName, { expr, dimensionsWithInlistFilter }]) => [
            mvName,
            MetricsViewFilterManager.parse(
              this.metricsViewsProvider,
              mvName,
              maybeWrapAndExpression(expr),
              dimensionsWithInlistFilter,
              existingManagers,
            ),
          ])
          .filter(([, manager]) => !!manager) as [
          string,
          MetricsViewFilterManager,
        ][],
      );
    });

    this.filterManagers = $derived.by(() =>
      this.buildFilterManagers(this.buildParamFilterManagers()),
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

    this.exprByMetricsView = $derived.by(() =>
      this.buildExpressionsByMetricsView(),
    );
    this.hasSomeFilter = $derived(
      Object.keys(this.exprByMetricsView).length > 0,
    );

    $effect(() => {
      this.exprByMetricsViewStore.set(this.exprByMetricsView);
    });
  }

  public setUrlParams = (searchParams: URLSearchParams) => {
    const newParsedUrlParams = {} as ParsedUrlParamsByMV;
    const newParams = [] as string[];
    let isSomeComplexFilter = false;

    const singularFilter = searchParams.get(ExploreStateURLParams.Filters);
    this.metricsViewsProvider.metricsViewNames.forEach((name: string) => {
      const paramKey = `${ExploreStateURLParams.Filters}.${name}`;
      const filterString = searchParams.get(paramKey) ?? singularFilter ?? "";
      newParams.push(`${paramKey}=${filterString}`);

      const { expr, dimensionsWithInlistFilter } =
        convertFilterParamToExpression(filterString);
      const wrappedExpr = maybeWrapAndExpression(expr);
      isSomeComplexFilter ||= expr
        ? isExpressionUnsupported(wrappedExpr!)
        : false;
      newParsedUrlParams[name] = {
        expr: wrappedExpr,
        dimensionsWithInlistFilter,
      };
    });

    const newParam = newParams.join("&");
    if (this.prevUrlParams === newParam) return;
    this.prevUrlParams = newParam;
    this.parsedUrlParams = newParsedUrlParams;
    this.isComplexFilter = isSomeComplexFilter;
    this.temporaryFilterName = undefined;
  };

  public setExprForMetricsView(mvName: string, expr: V1Expression | undefined) {
    const prevURLSearch = new URLSearchParams(this.prevUrlParams);

    const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
    const filterParam = expr ? convertExpressionToFilterParam(expr) : "";
    prevURLSearch.set(paramKey, filterParam);

    this.setUrlParams(prevURLSearch);
  }

  public setParamForMetricsView(mvName: string, param: string) {
    const prevURLSearch = new URLSearchParams(this.prevUrlParams);

    const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
    prevURLSearch.set(paramKey, param);

    this.setUrlParams(prevURLSearch);
  }

  /** Adds a chip for a filter that has no value yet. The chip opens as soon as it is rendered. */
  public addNewFilter(name: string) {
    if (
      !this.metricsViewsProvider.dimensionSpecs[name] &&
      !this.metricsViewsProvider.measureSpecs[name]
    ) {
      return;
    }

    this.temporaryFilterName = name;
  }

  public dimensionFilterAction(
    name: string,
    callback: (dimensionFilterManager: DimensionFilterManager) => any,
  ) {
    const dimensionFilterManager =
      this.filterManagersMap[name] ?? this.createDimensionFilterManager(name);
    if (!(dimensionFilterManager instanceof DimensionFilterManager)) return;

    const ret = callback(dimensionFilterManager);
    if (dimensionFilterManager.expr && !(name in this.filterManagersMap)) {
      this.filterManagers = {
        ...this.filterManagers,
        dimensions: [...this.filterManagers.dimensions, dimensionFilterManager],
      };
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

  /** Managers for the filters in the param, keyed by name and in param order. */
  private buildParamFilterManagers() {
    const dimensionsMap = {} as Record<string, DimensionFilterManager[]>;
    const measuresMap = {} as Record<string, MeasureFilterManager[]>;

    // Go through exprs of parsed url params for each metrics views.
    Object.entries(this.managerByMetricsView).forEach(([, mvManager]) => {
      if (!mvManager.expr) return;

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
    const shouldAddTempDimension = Boolean(
      this.temporaryFilterName &&
        !(this.temporaryFilterName in dimensionsMap) &&
        this.temporaryFilterName in this.metricsViewsProvider.dimensionSpecs,
    );
    sortFilterIterator(
      Object.keys(dimensionsMap),
      shouldAddTempDimension ? this.temporaryFilterName : undefined,
    ).forEach((name) => add(name));

    // Next comes measures from params
    const shouldAddTempMeasure =
      this.temporaryFilterName &&
      !(this.temporaryFilterName in measuresMap) &&
      this.temporaryFilterName in this.metricsViewsProvider.measureSpecs;
    sortFilterIterator(
      Object.keys(measuresMap),
      shouldAddTempMeasure ? this.temporaryFilterName : undefined,
    ).forEach((name) => add(name));

    return { measures, dimensions };
  }

  /**
   * The filter to send with queries for each metrics view.
   *
   * The param contributes a whole expression per metrics view, so any nesting it has survives.
   * Chips the param does not cover yet, the ones from the filter menu or a chart interaction,
   * are added to every metrics view that defines them.
   */
  private buildExpressionsByMetricsView() {
    const exprsByMetricsView = {} as Record<string, V1Expression[]>;
    const addExpr = (metricsViewName: string, expr: V1Expression) => {
      exprsByMetricsView[metricsViewName] ??= [];
      // Flattening a top level AND keeps the result a single flat AND in the common case.
      if (expr.cond?.op === V1Operation.OPERATION_AND) {
        exprsByMetricsView[metricsViewName].push(...(expr.cond.exprs ?? []));
      } else {
        exprsByMetricsView[metricsViewName].push(expr);
      }
    };

    const managersInParam = new Set<DimensionOrMeasureManager>();
    Object.entries(this.managerByMetricsView).forEach(
      ([metricsViewName, mvManager]) => {
        Object.values(mvManager.managerLookup)
          .flat()
          .forEach((manager) => managersInParam.add(manager));
        if (mvManager.expr) addExpr(metricsViewName, mvManager.expr);
      },
    );

    [...this.filterManagers.dimensions, ...this.filterManagers.measures]
      .filter((manager) => !!manager.expr && !managersInParam.has(manager))
      .forEach((manager) => {
        const specs =
          this.metricsViewsProvider.dimensionSpecs[manager.name] ??
          this.metricsViewsProvider.measureSpecs[manager.name] ??
          {};
        Object.keys(specs).forEach((metricsViewName) =>
          addExpr(metricsViewName, manager.expr!),
        );
      });

    return Object.fromEntries(
      Object.entries(exprsByMetricsView)
        .filter(([, exprs]) => exprs.length > 0)
        .map(([metricsViewName, exprs]) => [
          metricsViewName,
          createAndExpression(exprs),
        ]),
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
}

function sortFilterIterator(arr: string[], tempName: string | undefined) {
  const newArr = [...arr];
  if (tempName) newArr.push(tempName);
  return newArr.sort((a, b) => (a > b ? 1 : -1));
}
