import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilterManager.svelte.ts";
import { type V1Expression } from "@rilldata/web-common/runtime-client";
import {
  convertExpressionToFilterParam,
  convertFilterParamToExpression,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  forEachIdentifier,
  isExpressionUnsupported,
  maybeWrapAndExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilterManager.svelte.ts";
import { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

type FilterManager = MeasureFilterManager | DimensionFilterManager;
type ParsedUrlParamsByMV = Record<
  string,
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
  /** Filter param the managers are derived from. */
  public exprParam: string = $state("");
  /** Name of the filter added from the filter menu. Its chip opens as soon as it is rendered. */
  public temporaryFilterName: string | undefined = $state(undefined);

  /** True when the param cannot be represented as chips. */
  public isComplexFilter: boolean;
  public exprByMetricsView: Record<string, V1Expression>;
  public hasSomeFilter: boolean;

  public filterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };
  public readonly filterManagersMap: Record<string, FilterManager>;

  private prevUrlParams = $state("");
  private parsedUrlParams = $state<ParsedUrlParamsByMV>({});

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
  ) {
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
      isSomeComplexFilter ||= expr ? isExpressionUnsupported(expr) : false;
      newParsedUrlParams[name] = {
        expr: maybeWrapAndExpression(expr),
        dimensionsWithInlistFilter,
      };
    });

    const newParam = newParams.join("&");
    if (this.prevUrlParams === newParam) return;
    this.prevUrlParams = newParam;
    this.parsedUrlParams = newParsedUrlParams;
    this.isComplexFilter = isSomeComplexFilter;
  };

  public setExprForMetricsView(mvName: string, expr: V1Expression | undefined) {
    const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
    const filterParam = expr ? convertExpressionToFilterParam(expr) : "";
    this.setUrlParams(new URLSearchParams(`${paramKey}=${filterParam}`));
  }

  public setParamForMetricsView(mvName: string, param: string) {
    const paramKey = `${ExploreStateURLParams.Filters}.${mvName}`;
    this.setUrlParams(new URLSearchParams(`${paramKey}=${param}`));
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

  /** Managers for the filters in the param, keyed by name and in param order. */
  private buildParamFilterManagers() {
    const measuresMap = new Map<string, MeasureFilterManager>();
    const dimensionsMap = new Map<string, DimensionFilterManager>();

    if (this.isComplexFilter) return { measuresMap, dimensionsMap };

    // Go through exprs of parsed url params for each metrics views.
    Object.entries(this.parsedUrlParams).forEach(
      ([, { expr, dimensionsWithInlistFilter }]) => {
        if (!expr) return;

        forEachIdentifier(expr, (e, ident) => {
          // Both filter types are keyed off a dimension: a dimension filter by the dimension it filters
          // and a measure filter by the dimension its subquery groups by.
          if (!this.metricsViewsProvider.dimensionSpecs[ident]) return;

          const firstValueExpr = e?.cond?.exprs?.[1];

          if (firstValueExpr?.subquery) {
            // Having a subquery means this is a measure filter.
            const measureName = firstValueExpr.subquery.measures?.[0];
            if (!measureName) return;

            const measureFilterManager = this.createMeasureFilterManager(
              measureName,
              firstValueExpr,
            );
            if (measureFilterManager)
              measuresMap.set(measureName, measureFilterManager);
          } else {
            // Everything else is a dimension filter for now.
            const dimensionFilterManager = this.createDimensionFilterManager(
              ident,
              e,
              dimensionsWithInlistFilter.includes(ident),
            );
            if (dimensionFilterManager)
              dimensionsMap.set(ident, dimensionFilterManager);
          }
        });
      },
    );

    return { measuresMap, dimensionsMap };
  }

  /**
   * Managers in chip order: required filters first, then pinned ones, then the rest of the param
   * filters in param order, and finally the filters the param does not cover yet.
   */
  private buildFilterManagers({
    measuresMap,
    dimensionsMap,
  }: {
    measuresMap: Map<string, MeasureFilterManager>;
    dimensionsMap: Map<string, DimensionFilterManager>;
  }) {
    const dimensions: DimensionFilterManager[] = [];
    const measures: MeasureFilterManager[] = [];
    const added = new Set<string>();

    const add = (name: string) => {
      if (added.has(name)) return;

      const filterManager =
        dimensionsMap.get(name) ??
        measuresMap.get(name) ??
        this.createDimensionFilterManager(name) ??
        this.createMeasureFilterManager(name);
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
        !dimensionsMap.has(this.temporaryFilterName) &&
        this.temporaryFilterName in this.metricsViewsProvider.dimensionSpecs,
    );
    sortFilterIterator(
      dimensionsMap.keys(),
      shouldAddTempDimension ? this.temporaryFilterName : undefined,
    ).forEach((name) => add(name));

    // Next comes measures from params
    const shouldAddTempMeasure =
      this.temporaryFilterName &&
      !measuresMap.has(this.temporaryFilterName) &&
      this.temporaryFilterName in this.metricsViewsProvider.measureSpecs;
    sortFilterIterator(
      measuresMap.keys(),
      shouldAddTempMeasure ? this.temporaryFilterName : undefined,
    ).forEach((name) => add(name));

    return { measures, dimensions };
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

  private buildExpressionsByMetricsView() {
    const exprsByMetricsView = Object.fromEntries(
      this.metricsViewsProvider.metricsViewNames.map((m) => [
        m,
        [] as V1Expression[],
      ]),
    );

    this.filterManagers.dimensions.forEach((dfm) => {
      if (!dfm.expr) return;
      Object.keys(
        this.metricsViewsProvider.dimensionSpecs[dfm.name] ?? {},
      ).forEach((mv) => {
        exprsByMetricsView[mv]?.push(dfm.expr!);
      });
    });

    this.filterManagers.measures.forEach((mfm) => {
      if (!mfm.expr) return;
      Object.keys(
        this.metricsViewsProvider.measureSpecs[mfm.name] ?? {},
      ).forEach((mv) => {
        exprsByMetricsView[mv]?.push(mfm.expr!);
      });
    });

    return Object.fromEntries(
      Object.entries(exprsByMetricsView)
        .filter(([, exprs]) => exprs.length > 0)
        .map(([mv, exprs]) => [mv, createAndExpression(exprs)]),
    );
  }
}

function sortFilterIterator(
  iterator: MapIterator<string> | SetIterator<string>,
  tempName: string | undefined,
) {
  const newArr = [...iterator];
  if (tempName) newArr.push(tempName);
  return newArr.sort((a, b) => (a > b ? 1 : -1));
}
