import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilterManager.svelte.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  forEachIdentifier,
  isExpressionUnsupported,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilterManager.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

type FilterManager = MeasureFilterManager | DimensionFilterManager;

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

  public filterManagers: {
    measures: MeasureFilterManager[];
    dimensions: DimensionFilterManager[];
  };

  private readonly parsedParam: ReturnType<
    typeof convertFilterParamToExpression
  >;
  private readonly filterManagersMap: Record<string, FilterManager>;

  public constructor(
    public readonly metricsViewsProvider: MetricsViewsProvider,
    public readonly yamlConfigProvider: YAMLConfigProvider,
  ) {
    this.parsedParam = $derived.by(() => {
      const { expr, dimensionsWithInlistFilter } =
        convertFilterParamToExpression(this.exprParam);
      // A param with a single filter parses to that filter alone, so wrap it to get the AND of
      // individual filters the rest of the class works with.
      const needsWrapper =
        !!expr &&
        expr.cond?.op !== V1Operation.OPERATION_AND &&
        expr.cond?.op !== V1Operation.OPERATION_OR;
      return {
        expr: needsWrapper ? createAndExpression([expr]) : expr,
        dimensionsWithInlistFilter,
      };
    });
    this.isComplexFilter = $derived(
      this.parsedParam.expr
        ? isExpressionUnsupported(this.parsedParam.expr)
        : false,
    );
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
  }

  public setExprParam(exprParam: string) {
    if (exprParam === this.exprParam) return;
    console.log("new filter", exprParam);
    this.temporaryFilterName = undefined;
    this.exprParam = exprParam;
  }

  /** Adds a chip for a filter that has no value yet. The chip opens as soon as it is rendered. */
  public addNewFilter(name: string) {
    if (
      !this.metricsViewsProvider.dimensionSpecs[name] &&
      !this.metricsViewsProvider.measureSpecs[name]
    ) {
      return;
    }

    console.log("new filter", name);
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
    this.exprParam = "";
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

  /** Managers for the filters in the param, keyed by name and in param order. */
  private buildParamFilterManagers() {
    const measuresMap = new Map<string, MeasureFilterManager>();
    const dimensionsMap = new Map<string, DimensionFilterManager>();

    const { expr, dimensionsWithInlistFilter } = this.parsedParam;
    if (!expr || this.isComplexFilter) return { measuresMap, dimensionsMap };

    forEachIdentifier(expr, (e, ident) => {
      // Both filter types are keyed off a dimension: a dimension filter by the dimension it filters
      // and a measure filter by the dimension its subquery groups by.
      if (!this.metricsViewsProvider.dimensionSpecs[ident]) return;

      const firstValueExpr = e?.cond?.exprs?.[1];

      if (firstValueExpr?.subquery) {
        const measureName = firstValueExpr.subquery.measures?.[0];
        if (!measureName) return;

        const measureFilterManager = this.createMeasureFilterManager(
          measureName,
          firstValueExpr,
        );
        if (measureFilterManager)
          measuresMap.set(measureName, measureFilterManager);
      } else {
        const dimensionFilterManager = this.createDimensionFilterManager(
          ident,
          e,
          dimensionsWithInlistFilter.includes(ident),
        );
        if (dimensionFilterManager)
          dimensionsMap.set(ident, dimensionFilterManager);
      }
    });

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
