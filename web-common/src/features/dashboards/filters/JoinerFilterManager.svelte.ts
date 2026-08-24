import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { DimensionFilterManager } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/DimensionFilterManager.svelte.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import { MeasureFilterManager } from "@rilldata/web-common/features/dashboards/filters/measure-filters/MeasureFilterManager.svelte.ts";
import type {
  MetricsViewName,
  MetricsViewsProvider,
} from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import {
  createAndExpression,
  createOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import type { YAMLConfigProvider } from "@rilldata/web-common/features/dashboards/providers/YAMLConfigProvider.svelte.ts";
import type { FilterEventEmitter } from "@rilldata/web-common/features/dashboards/filters/filter-events.ts";

export type DimensionOrMeasureManager =
  | DimensionFilterManager
  | MeasureFilterManager;
export type AnyManager = DimensionOrMeasureManager | JoinerFilterManager;
export type AllManagers = {
  dimensionManagers: DimensionFilterManager[];
  measureManagers: MeasureFilterManager[];
  joinerManagers: JoinerFilterManager[];
};

/**
 * Manager for an AND/OR expression.
 */
export class JoinerFilterManager {
  public expr: Record<MetricsViewName, V1Expression>;
  public dimensionOnlyExpr: Record<MetricsViewName, V1Expression>;
  // String representation of the filter expression. Used to check duplicate expressions across metrics views.
  public param: Record<MetricsViewName, string>;

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
    private readonly metricsViewsProvider: MetricsViewsProvider,
    public type:
      | typeof V1Operation.OPERATION_AND
      | typeof V1Operation.OPERATION_OR,
    managers: AllManagers,
    public readonly inList: string[],
  ) {
    this.managers = $state(managers);

    this.expr = $derived(
      this.buildExpression(
        this.managers,
        this.metricsViewsProvider.metricsViewNames,
        false,
      ),
    );
    this.dimensionOnlyExpr = $derived(
      this.buildExpression(
        this.managers,
        this.metricsViewsProvider.metricsViewNames,
        true,
      ),
    );

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
    this.param = $derived.by(() =>
      Object.fromEntries(
        this.metricsViewsProvider.metricsViewNames.map((mvName) => [
          mvName,
          this.expr[mvName]
            ? convertExpressionToFilterParam(
                this.expr[mvName],
                this.inListDimensions,
              )
            : "",
        ]),
      ),
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
    expr: V1Expression | undefined,
    dimensionsWithInlistFilter: string[],
    yamlConfigProvider: YAMLConfigProvider | undefined,
    events: FilterEventEmitter | undefined,
  ): AnyManager | undefined {
    const op = expr?.cond?.op;
    if (op === V1Operation.OPERATION_AND || op === V1Operation.OPERATION_OR) {
      const dimensionManagers: DimensionFilterManager[] = [];
      const measureManagers: MeasureFilterManager[] = [];
      const joinerManagers: JoinerFilterManager[] = [];
      const added = new Set<string>();

      const add = (manager: AnyManager, force: boolean) => {
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
        } else if (manager instanceof JoinerFilterManager) {
          joinerManagers.push(manager);
        }
      };

      expr?.cond?.exprs?.forEach((e) => {
        const manager = JoinerFilterManager.parse(
          metricsViewsProvider,
          e,
          dimensionsWithInlistFilter,
          // Deeper parse should not add required/pinned filters. So pass undefined here.
          undefined,
          events,
        );
        if (!manager) return;

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
                { events },
              ) ??
              MeasureFilterManager.createForMetricsViews(
                metricsViewsProvider,
                requiredFilter,
                { events },
              );
            if (manager) add(manager, false);
          });
      }

      return new JoinerFilterManager(
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
      if (
        !Object.keys(metricsViewsProvider.dimensionSpecs[ident] ?? {}).length
      ) {
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
          { initExpr: firstValueExpr, events },
        );
      } else {
        // Everything else is a dimension filter for now.
        return DimensionFilterManager.createForMetricsViews(
          metricsViewsProvider,
          ident,
          {
            initExpr: expr,
            isInList: dimensionsWithInlistFilter.includes(ident),
            events,
          },
        );
      }
    }
  }

  public maybeAddDimensionFilter(
    dimensionFilterManager: DimensionFilterManager,
  ) {
    const dimensionName = dimensionFilterManager.name;
    const alreadyAdded = !!this.managerLookup[dimensionName];
    const dimensionNotInMV = !Object.keys(
      this.metricsViewsProvider.dimensionSpecs[dimensionName] ?? {},
    ).length;
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
    const measureNotInMV = !Object.keys(
      this.metricsViewsProvider.measureSpecs[measureName] ?? {},
    ).length;
    if (alreadyAdded || measureNotInMV) return;

    this.managers = {
      ...this.managers,
      measureManagers: [...this.managers.measureManagers, measureFilterManager],
    };
  }

  private buildExpression(
    managers: AllManagers,
    metricsViewNames: string[],
    dimensionOnly: boolean,
  ) {
    const exprsByMv: Record<string, V1Expression[]> = {};

    const addExprForMv = (expr: V1Expression | undefined, mvName: string) => {
      if (!expr) return;
      exprsByMv[mvName] ??= [];
      exprsByMv[mvName].push(expr);
    };
    // A manager is shared across the metrics views, but the identifier it filters on is not:
    // a filter only goes to the metrics views that define it, since the rest cannot resolve it.
    const addExpr = (
      expr: V1Expression | undefined,
      definedBy: (mvName: string) => boolean,
    ) => {
      if (!expr) return;
      metricsViewNames
        .filter(definedBy)
        .forEach((mvName) => addExprForMv(expr, mvName));
    };

    managers.dimensionManagers.forEach((dfm) =>
      addExpr(
        dfm.expr,
        (mvName) =>
          !!this.metricsViewsProvider.dimensionSpecs[dfm.name]?.[mvName],
      ),
    );
    if (!dimensionOnly) {
      // A measure filter is a subquery grouped by a dimension, so both have to be defined.
      managers.measureManagers.forEach((mfm) =>
        addExpr(
          mfm.expr,
          (mvName) =>
            !!this.metricsViewsProvider.measureSpecs[mfm.name]?.[mvName] &&
            !!this.metricsViewsProvider.dimensionSpecs[mfm.dimension]?.[mvName],
        ),
      );
    }
    managers.joinerManagers.forEach((jfm) => {
      if (dimensionOnly) {
        Object.entries(jfm.dimensionOnlyExpr).forEach(([mvName, expr]) =>
          addExprForMv(expr, mvName),
        );
      } else {
        Object.entries(jfm.expr).forEach(([mvName, expr]) =>
          addExprForMv(expr, mvName),
        );
      }
    });

    const joinerCreator =
      this.type === V1Operation.OPERATION_AND
        ? createAndExpression
        : createOrExpression;

    return Object.fromEntries(
      Object.entries(exprsByMv).map(([mvName, exprs]) => [
        mvName,
        joinerCreator(exprs),
      ]),
    );
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
