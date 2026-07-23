import {
  DimensionFilterManager,
  type DimensionFilterManagerInit,
} from "@rilldata/web-common/features/dashboards/filters/manager/dimension-filter-manager.svelte.ts";
import type { V1Expression } from "@rilldata/web-admin/client";
import { convertFilterParamToExpression } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  createInExpression,
  forEachIdentifier,
  getValuesInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  type MetricsViewSpecDimension,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";

export class ExpressionFilterManager {
  public dimensionFilterManagers: DimensionFilterManager[] = $state([]);
  public expr: V1Expression;
  public exprParam: string = "";

  private temporaryFilter: DimensionFilterManager | undefined =
    $state(undefined);

  public constructor() {
    this.expr = $derived.by(() => {
      const exprs = this.dimensionFilterManagers
        .filter((dfm) => !!dfm.expr)
        .map((d) => d.expr as V1Expression);
      return createAndExpression(exprs);
    });
  }

  public setExprParam(
    exprParam: string,
    dimensionIdMap: Map<string, MetricsViewSpecDimension>,
    metricsViewName: string,
  ) {
    if (exprParam === this.exprParam) return;
    this.exprParam = exprParam;

    const { expr, dimensionsWithInlistFilter } =
      convertFilterParamToExpression(exprParam);
    // TODO: check complex filter
    if (!expr) {
      this.dimensionFilterManagers = [];
      return;
    }

    const addedDimension = new Set<string>();
    const newDimensionFilterManagers: DimensionFilterManager[] = [];
    forEachIdentifier(expr, (e, ident) => {
      const dim = dimensionIdMap.get(ident);
      if (!dim) return;
      addedDimension.add(ident);

      const isInListMode = dimensionsWithInlistFilter.includes(ident);
      newDimensionFilterManagers.push(
        new DimensionFilterManager(
          ident,
          getDimensionDisplayName(dim),
          new Map<string, MetricsViewSpecDimension>([[metricsViewName, dim]]),
          false,
          e,
          isInListMode,
        ),
      );
    });

    if (this.temporaryFilter) {
      if (!addedDimension.has(this.temporaryFilter.name)) {
        newDimensionFilterManagers.push(this.temporaryFilter);
      } else {
        this.temporaryFilter = undefined;
      }
    }

    this.dimensionFilterManagers = newDimensionFilterManagers;
  }

  public addTemporaryDimensionFilter(
    name: string,
    dimensionIdMap: Map<string, MetricsViewSpecDimension>,
    metricsViewName: string,
  ) {
    const dim = dimensionIdMap.get(name);
    if (!dim) return;

    this.temporaryFilter = new DimensionFilterManager(
      name,
      getDimensionDisplayName(dim),
      new Map<string, MetricsViewSpecDimension>([[metricsViewName, dim]]),
      false,
    );
    this.dimensionFilterManagers = [
      ...this.dimensionFilterManagers,
      this.temporaryFilter,
    ];
  }

  public clear() {
    this.dimensionFilterManagers = [];
    this.temporaryFilter = undefined;
  }

  public getOtherDimensionsFilter(name: string) {
    const exprs = this.dimensionFilterManagers
      .filter((dfm) => !!dfm.expr && dfm.name !== name)
      .map((d) => d.expr as V1Expression);
    return exprs.length === 0 ? undefined : createAndExpression(exprs);
  }
}
