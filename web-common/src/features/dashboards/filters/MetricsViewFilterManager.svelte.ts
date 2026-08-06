import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters-v2/DimensionFilterManager.svelte.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters-v2/MeasureFilterManager.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  getDimensionDisplayName,
  getMeasureDisplayName,
} from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import {
  createAndExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";

export type DimensionOrMeasureManager =
  | DimensionFilterManager
  | MeasureFilterManager;
export type AnyManager = DimensionOrMeasureManager | MetricsViewFilterManager;

/**
 * Manager for an AND/OR expression for a metrics view.
 */
export class MetricsViewFilterManager {
  public expr: V1Expression | undefined;
  // String representation of the filter expression. Used to check duplicate expressions across metrics views.
  public param: string;

  public managerLookup: Record<
    string,
    (DimensionFilterManager | MeasureFilterManager)[]
  >;
  public hasSomeFilter: boolean;
  public isComplexFilter: boolean;

  public constructor(
    public readonly metricsViewName: string,
    public type:
      | typeof V1Operation.OPERATION_AND
      | typeof V1Operation.OPERATION_OR,
    public readonly managers: {
      dimensionManagers: DimensionFilterManager[];
      measureManagers: MeasureFilterManager[];
      joinerManagers: MetricsViewFilterManager[];
    },
    private readonly inList: string[],
  ) {
    this.expr = $derived(this.buildExpressionsByMetricsView());
    this.param = $derived(
      this.expr ? convertExpressionToFilterParam(this.expr, this.inList) : "",
    );

    this.managerLookup = $derived(this.buildManagerLookup());
    this.hasSomeFilter = $derived(Object.values(this.managerLookup).length > 0);
    // Currently only single level AND joiner is supported where each dimension/measure has a single filter.
    // Anything else is considered complex filter, and we only show an editable text box.
    this.isComplexFilter = $derived(
      this.type === V1Operation.OPERATION_OR ||
        Object.values(this.managerLookup).some(
          (managers) => managers.length > 1,
        ) ||
        this.managers.joinerManagers.length > 0,
    );
  }

  public static parse(
    metricsViewsProvider: MetricsViewsProvider,
    metricsViewName: string,
    expr: V1Expression | undefined,
    dimensionsWithInlistFilter: string[],
    existingManagers: Map<string, AnyManager>,
  ): AnyManager | undefined {
    const op = expr?.cond?.op;
    if (op === V1Operation.OPERATION_AND || op === V1Operation.OPERATION_OR) {
      const dimensionManagers: DimensionFilterManager[] = [];
      const measureManagers: MeasureFilterManager[] = [];
      const joinerManagers: MetricsViewFilterManager[] = [];

      expr?.cond?.exprs?.forEach((e) => {
        let manager = MetricsViewFilterManager.parse(
          metricsViewsProvider,
          metricsViewName,
          e,
          dimensionsWithInlistFilter,
          existingManagers,
        );
        if (!manager) return;

        // Prefer to share class for similar expressions.
        // This is a hack to make a single filter across the metrics view work.
        // TODO: find a better solution to share
        if (existingManagers.has(manager.param)) {
          manager = existingManagers.get(manager.param)!;
        } else {
          existingManagers.set(manager.param, manager);
        }

        if (manager instanceof DimensionFilterManager) {
          dimensionManagers.push(manager);
        } else if (manager instanceof MeasureFilterManager) {
          measureManagers.push(manager);
        } else if (manager instanceof MetricsViewFilterManager) {
          joinerManagers.push(manager);
        }
      });

      return new MetricsViewFilterManager(
        metricsViewName,
        op,
        {
          dimensionManagers,
          measureManagers,
          joinerManagers,
        },
        dimensionsWithInlistFilter,
      );
    } else {
      const ident = expr?.cond?.exprs?.[0]?.ident ?? "";
      // Both filter types are keyed off a dimension: a dimension filter by the dimension it filters
      // and a measure filter by the dimension its subquery groups by.
      if (!metricsViewsProvider.dimensionSpecs[ident]?.[metricsViewName]) {
        return undefined;
      }

      const firstValueExpr = expr?.cond?.exprs?.[1];

      if (firstValueExpr?.subquery) {
        // Having a subquery means this is a measure filter.
        const measureName = firstValueExpr.subquery.measures?.[0];
        if (!measureName) return undefined;
        const measureSpec =
          metricsViewsProvider.measureSpecs[measureName]?.[metricsViewName];
        if (!measureSpec) return undefined;

        return new MeasureFilterManager(
          measureName,
          getMeasureDisplayName(measureSpec),
          firstValueExpr,
        );
      } else {
        // Everything else is a dimension filter for now.
        const dimensionSpec =
          metricsViewsProvider.dimensionSpecs[ident]?.[metricsViewName];
        if (!dimensionSpec) return undefined;

        return new DimensionFilterManager(
          ident,
          getDimensionDisplayName(dimensionSpec),
          expr,
          dimensionsWithInlistFilter.includes(ident),
        );
      }
    }
  }

  private buildExpressionsByMetricsView() {
    const dimensionExprs = this.managers.dimensionManagers
      .map((dfm) => dfm.expr)
      .filter(Boolean) as V1Expression[];
    const measureExprs = this.managers.measureManagers
      .map((mfm) => mfm.expr)
      .filter(Boolean) as V1Expression[];
    const joinerExprs = this.managers.joinerManagers
      .map((jm) => jm.expr)
      .filter(Boolean) as V1Expression[];
    const exprs = dimensionExprs.concat(measureExprs).concat(joinerExprs);
    if (exprs.length === 0) return undefined;

    if (this.type === V1Operation.OPERATION_AND) {
      return createAndExpression(exprs);
    } else {
      return createOrExpression(exprs);
    }
  }

  private buildManagerLookup() {
    const lookup = {} as Record<
      string,
      (DimensionFilterManager | MeasureFilterManager)[]
    >;
    this.managers.dimensionManagers.forEach((manager) => {
      lookup[manager.name] ??= [];
      lookup[manager.name].push(manager);
    });
    this.managers.measureManagers.forEach((manager) => {
      lookup[manager.name] ??= [];
      lookup[manager.name].push(manager);
    });
    this.managers.joinerManagers.forEach((manager) => {
      Object.entries(manager.managerLookup).forEach(([name, managers]) => {
        lookup[name] ??= [];
        lookup[name].push(...managers);
      });
    });
    return lookup;
  }
}
