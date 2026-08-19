import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  createAndExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";

export type DimensionOrMeasureManager =
  | DimensionFilterManager
  | MeasureFilterManager;
export type AnyManager = DimensionOrMeasureManager | MetricsViewFilterManager;
export type AllManagers = {
  dimensionManagers: DimensionFilterManager[];
  measureManagers: MeasureFilterManager[];
  joinerManagers: MetricsViewFilterManager[];
};

/**
 * Manager for an AND/OR expression for a metrics view.
 */
export class MetricsViewFilterManager {
  public expr: V1Expression | undefined;
  public dimensionOnlyExpr: V1Expression | undefined;
  // String representation of the filter expression. Used to check duplicate expressions across metrics views.
  public param: string;

  public managers: AllManagers;
  public managerLookup: Record<
    string,
    (DimensionFilterManager | MeasureFilterManager)[]
  >;
  // Dimensions currently filtered in "In List" mode, which have their own param syntax.
  public inListDimensions: string[];
  public hasSomeFilter: boolean;
  public isComplexFilter: boolean;

  public constructor(
    public readonly metricsViewName: string,
    private readonly metricsViewsProvider: MetricsViewsProvider,
    public type:
      | typeof V1Operation.OPERATION_AND
      | typeof V1Operation.OPERATION_OR,
    managers: AllManagers,
    public readonly inList: string[],
  ) {
    this.managers = $state(managers);

    this.expr = $derived(this.buildExpression(this.managers));
    this.dimensionOnlyExpr = $derived(this.buildDimensionOnlyExpression());

    this.managerLookup = $derived(this.buildManagerLookup());
    // `inList` is what the param was parsed with. The chips can change the mode afterwards without
    // changing the expression, so the param follows the modes the managers hold right now.
    this.inListDimensions = $derived(
      Object.values(this.managerLookup)
        .flat()
        .filter(
          (manager) =>
            manager instanceof DimensionFilterManager &&
            manager.mode === DimensionFilterMode.InList,
        )
        .map((manager) => manager.name),
    );
    this.param = $derived(
      this.expr
        ? convertExpressionToFilterParam(this.expr, this.inListDimensions)
        : "",
    );

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
    yamlConfigProvider: YAMLConfigProvider | undefined,
  ): AnyManager | undefined {
    const op = expr?.cond?.op;
    if (op === V1Operation.OPERATION_AND || op === V1Operation.OPERATION_OR) {
      const dimensionManagers: DimensionFilterManager[] = [];
      const measureManagers: MeasureFilterManager[] = [];
      const joinerManagers: MetricsViewFilterManager[] = [];
      const added = new Set<string>();

      const add = (manager: AnyManager, force: boolean) => {
        if (manager.param && !existingManagers.has(manager.param)) {
          existingManagers.set(manager.param, manager);
        } else if (
          !manager.param &&
          !(manager instanceof MetricsViewFilterManager)
        ) {
          manager = existingManagers.get(manager.name) ?? manager;
          existingManagers.set(
            (manager as DimensionOrMeasureManager).name,
            manager,
          );
        }

        if (
          manager instanceof DimensionFilterManager &&
          (!added.has(manager.name) || force)
        ) {
          added.add(manager.name);
          dimensionManagers.push(manager);
        } else if (
          manager instanceof MeasureFilterManager &&
          (!added.has(manager.name) || force)
        ) {
          added.add(manager.name);
          measureManagers.push(manager);
        } else if (manager instanceof MetricsViewFilterManager) {
          joinerManagers.push(manager);
        }
      };

      expr?.cond?.exprs?.forEach((e) => {
        let manager = MetricsViewFilterManager.parse(
          metricsViewsProvider,
          metricsViewName,
          e,
          dimensionsWithInlistFilter,
          existingManagers,
          undefined,
        );
        if (!manager) return;
        // Prefer to share class for similar expressions.
        // This is a hack to make a single filter across the metrics view work.
        // TODO: find a better solution to share
        manager = existingManagers.get(manager.param) ?? manager;

        add(manager, true);
      });

      // Add required and pinned filters when yamlConfigProvider is provided.
      // yamlConfigProvider is only provided for the top level manager.
      // TODO: handle advanced filters
      if (yamlConfigProvider) {
        Object.keys(yamlConfigProvider.requiredFilters)
          .concat(Object.keys(yamlConfigProvider.pinnedFilters))
          .forEach((requiredFilter) => {
            const manager =
              DimensionFilterManager.createForMetricsViews(
                metricsViewsProvider,
                requiredFilter,
                metricsViewName,
              ) ??
              MeasureFilterManager.createForMetricsViews(
                metricsViewsProvider,
                requiredFilter,
                metricsViewName,
              );
            if (manager) add(manager, false);
          });
      }

      return new MetricsViewFilterManager(
        metricsViewName,
        metricsViewsProvider,
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
        return MeasureFilterManager.createForMetricsViews(
          metricsViewsProvider,
          measureName,
          metricsViewName,
          firstValueExpr,
        );
      } else {
        // Everything else is a dimension filter for now.
        return DimensionFilterManager.createForMetricsViews(
          metricsViewsProvider,
          ident,
          metricsViewName,
          expr,
          dimensionsWithInlistFilter.includes(ident),
        );
      }
    }
  }

  public maybeAddDimensionFilter(
    dimensionFilterManager: DimensionFilterManager,
  ) {
    const dimensionName = dimensionFilterManager.name;
    const alreadyAdded = !!this.managerLookup[dimensionName];
    const dimensionNotInMV =
      !this.metricsViewsProvider.dimensionSpecs[dimensionName]?.[
        this.metricsViewName
      ];
    if (alreadyAdded || dimensionNotInMV) return;

    this.managers = {
      ...this.managers,
      dimensionManagers: [
        ...this.managers.dimensionManagers,
        dimensionFilterManager,
      ],
    };
  }

  public maybeAddMeasureFilter(measureFilterManager: MeasureFilterManager) {
    const measureName = measureFilterManager.name;
    const alreadyAdded = !!this.managerLookup[measureName];
    const measureNotInMV =
      !this.metricsViewsProvider.measureSpecs[measureName]?.[
        this.metricsViewName
      ];
    if (alreadyAdded || measureNotInMV) return;

    this.managers = {
      ...this.managers,
      measureManagers: [...this.managers.measureManagers, measureFilterManager],
    };
  }

  private buildExpression(managers: AllManagers) {
    const dimensionExprs = managers.dimensionManagers
      .map((dfm) => dfm.expr)
      .filter(Boolean) as V1Expression[];
    const measureExprs = managers.measureManagers
      .map((mfm) => mfm.expr)
      .filter(Boolean) as V1Expression[];
    const joinerExprs = managers.joinerManagers
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

  private buildDimensionOnlyExpression() {
    const dimensionExprs = this.managers.dimensionManagers
      .map((dfm) => dfm.expr)
      .filter(Boolean) as V1Expression[];
    if (dimensionExprs.length === 0) return undefined;

    if (this.type === V1Operation.OPERATION_AND) {
      return createAndExpression(dimensionExprs);
    } else {
      return createOrExpression(dimensionExprs);
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
