import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/DimensionFilterManager.svelte.ts";
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
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/manager/MeasureFilterManager.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

type FilterManager = DimensionFilterManager | MeasureFilterManager;

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
  /**
   * Names of the filters that get a chip even though the param does not cover them and they are
   * neither required nor pinned: the one added from the filter menu and the ones created by a chart
   * interaction.
   */
  private localFilterNames: string[] = $state([]);

  public dimensionFilterManagers: DimensionFilterManager[];
  public measureFilterManagers: MeasureFilterManager[];
  /** True when the param cannot be represented as chips. */
  public isComplexFilter: boolean;
  public expr: V1Expression;

  private parsedParam: ReturnType<typeof convertFilterParamToExpression>;
  /** Managers for the filters in the param, keyed by name and in param order. */
  private paramFilterManagers: {
    dimension: Map<string, DimensionFilterManager>;
    measure: Map<string, MeasureFilterManager>;
  };
  private filterManagers: {
    dimension: DimensionFilterManager[];
    measure: MeasureFilterManager[];
  };
  /**
   * Managers for the filters the param does not cover. They are kept across rebuilds so that a
   * rebuild cannot drop an edit that has not reached the param yet; `setExprParam` drops the ones
   * the param has caught up with.
   */
  private readonly outOfParamManagers = new Map<string, FilterManager>();

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
    this.paramFilterManagers = $derived.by(() =>
      this.buildParamFilterManagers(),
    );
    this.filterManagers = $derived.by(() => this.buildFilterManagers());
    this.dimensionFilterManagers = $derived(this.filterManagers.dimension);
    this.measureFilterManagers = $derived(this.filterManagers.measure);

    this.expr = $derived.by(() => {
      const dimExprs = this.dimensionFilterManagers
        .filter((dfm) => !!dfm.expr)
        .map((d) => d.expr as V1Expression);
      const mesExprs = this.measureFilterManagers
        .filter((mfm) => !!mfm.expr)
        .map((m) => m.expr as V1Expression);
      return createAndExpression(dimExprs.concat(mesExprs));
    });
  }

  public setExprParam(exprParam: string) {
    this.exprParam = exprParam;

    // Once the param covers a filter its param manager takes over, so drop the manager that was
    // kept for it. Keeping it would restore the value it used to have when the filter is cleared.
    for (const name of this.outOfParamManagers.keys()) {
      if (this.isInParam(name)) this.outOfParamManagers.delete(name);
    }

    const localFilterNames = this.localFilterNames.filter(
      (name) => !this.isInParam(name),
    );
    // Assigning unconditionally would rebuild every manager on any url change.
    if (localFilterNames.length !== this.localFilterNames.length) {
      this.localFilterNames = localFilterNames;
    }

    if (this.temporaryFilterName && this.isInParam(this.temporaryFilterName)) {
      this.temporaryFilterName = undefined;
    }
  }

  /** Adds a chip for a filter that has no value yet. The chip opens as soon as it is rendered. */
  public addTemporaryFilter(name: string) {
    if (!this.getFilterManager(name)) return;

    this.temporaryFilterName = name;
    this.addLocalFilterName(name);
  }

  public dimensionFilterAction(
    name: string,
    callback: (dimensionFilterManager: DimensionFilterManager) => any,
  ) {
    const dimensionFilterManager = this.getFilterManager(name);
    if (!(dimensionFilterManager instanceof DimensionFilterManager)) return;

    const ret = callback(dimensionFilterManager);
    // A filter this action created needs a chip until the param catches up with it.
    if (dimensionFilterManager.expr) this.addLocalFilterName(name);
    return ret;
  }

  public clear() {
    this.exprParam = "";
    this.localFilterNames = [];
    this.temporaryFilterName = undefined;
    this.outOfParamManagers.clear();
  }

  public getOtherDimensionsFilter(name: string) {
    const exprs = this.dimensionFilterManagers
      .filter((dfm) => !!dfm.expr && dfm.name !== name)
      .map((d) => d.expr as V1Expression);
    return exprs.length === 0 ? undefined : createAndExpression(exprs);
  }

  /** Managers for the filters in the param, keyed by name and in param order. */
  private buildParamFilterManagers() {
    const dimension = new Map<string, DimensionFilterManager>();
    const measure = new Map<string, MeasureFilterManager>();

    const { expr, dimensionsWithInlistFilter } = this.parsedParam;
    if (!expr || this.isComplexFilter) return { dimension, measure };

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
          ident,
          firstValueExpr,
        );
        if (measureFilterManager)
          measure.set(measureName, measureFilterManager);
      } else {
        const dimensionFilterManager = this.createDimensionFilterManager(
          ident,
          e,
          dimensionsWithInlistFilter.includes(ident),
        );
        if (dimensionFilterManager)
          dimension.set(ident, dimensionFilterManager);
      }
    });

    return { dimension, measure };
  }

  /**
   * Managers in chip order: required filters first, then pinned ones, then the rest of the param
   * filters in param order, and finally the filters the param does not cover yet.
   */
  private buildFilterManagers() {
    const dimension: DimensionFilterManager[] = [];
    const measure: MeasureFilterManager[] = [];
    const added = new Set<string>();

    const add = (name: string) => {
      if (added.has(name)) return;

      const filterManager = this.getFilterManager(name);
      if (!filterManager) return;

      added.add(name);
      if (filterManager instanceof MeasureFilterManager) {
        measure.push(filterManager);
      } else {
        dimension.push(filterManager);
      }
    };

    // Required and pinned filters lead the chips and have a chip even without a value.
    const { requiredFilters, pinnedFilters } = this.yamlConfigProvider;
    const configuredFilters = new Set([
      ...Object.keys(requiredFilters),
      ...Object.keys(pinnedFilters),
    ]);
    for (const name of configuredFilters) {
      if (requiredFilters[name] || pinnedFilters[name]) add(name);
    }

    this.paramFilterManagers.dimension.forEach((_, name) => add(name));
    this.paramFilterManagers.measure.forEach((_, name) => add(name));
    this.localFilterNames.forEach((name) => add(name));

    return { dimension, measure };
  }

  /**
   * Manager for a filter by name: the param one when the param covers the filter, else one that is
   * kept until the param does. Returns undefined when no metrics view defines the filter.
   */
  private getFilterManager(name: string): FilterManager | undefined {
    const paramFilterManager =
      this.paramFilterManagers.dimension.get(name) ??
      this.paramFilterManagers.measure.get(name);
    if (paramFilterManager) return paramFilterManager;

    const outOfParamManager = this.outOfParamManagers.get(name);
    if (outOfParamManager) return outOfParamManager;

    const filterManager =
      this.createDimensionFilterManager(name) ??
      this.createMeasureFilterManager(name);
    if (filterManager) this.outOfParamManagers.set(name, filterManager);
    return filterManager;
  }

  private isInParam(name: string) {
    return (
      this.paramFilterManagers.dimension.has(name) ||
      this.paramFilterManagers.measure.has(name)
    );
  }

  private addLocalFilterName(name: string) {
    if (this.isInParam(name) || this.localFilterNames.includes(name)) return;
    this.localFilterNames = [...this.localFilterNames, name];
  }

  /** Returns undefined when no metrics view defines the dimension. */
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
      dimensionSpecs,
      initExpr,
      isInList,
    );
  }

  /** Returns undefined when no metrics view defines the measure. */
  private createMeasureFilterManager(
    name: string,
    initDimension?: string,
    initExpr?: V1Expression,
  ) {
    const measureSpecs = this.metricsViewsProvider.measureSpecs[name];
    if (!measureSpecs) return undefined;

    return new MeasureFilterManager(
      name,
      getMeasureDisplayName(Object.values(measureSpecs)[0]),
      measureSpecs,
      initDimension,
      initExpr,
    );
  }
}
