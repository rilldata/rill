import {
  DimensionFilterManager,
  type DimensionFilterManagerInit,
} from "@rilldata/web-common/features/dashboards/filters/manager/dimension-filter-manager.svelte.ts";
import type { V1Expression } from "@rilldata/web-admin/client";
import {
  convertExpressionToFilterParam,
  convertFilterParamToExpression,
} from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import {
  createAndExpression,
  forEachIdentifier,
  getValuesInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import {
  type MetricsViewSpecDimension,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";

export class FilterManagerManager {
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

      const op = e.cond?.op;
      let init: DimensionFilterManagerInit | undefined = undefined;
      if (op === V1Operation.OPERATION_IN || op === V1Operation.OPERATION_NIN) {
        const isInListMode = dimensionsWithInlistFilter.includes(ident);
        init = {
          mode: isInListMode
            ? DimensionFilterMode.InList
            : DimensionFilterMode.Select,
          selectedValues: getValuesInExpression(e),
          exclude: op === V1Operation.OPERATION_NIN,
          inputText: undefined,
        };
      } else if (
        op === V1Operation.OPERATION_LIKE ||
        op === V1Operation.OPERATION_NLIKE
      ) {
        init = {
          mode: DimensionFilterMode.Contains,
          selectedValues: [],
          inputText: e.cond?.exprs?.[1]?.val?.toString?.() ?? "",
          exclude: op === V1Operation.OPERATION_NLIKE,
        };
      }

      if (!init) return;

      newDimensionFilterManagers.push(
        new DimensionFilterManager(
          ident,
          getDimensionDisplayName(dim),
          new Map<string, MetricsViewSpecDimension>([[metricsViewName, dim]]),
          false,
          init,
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
      { mode: DimensionFilterMode.Select },
    );
    this.dimensionFilterManagers = [
      ...this.dimensionFilterManagers,
      this.temporaryFilter,
    ];
  }

  public clear() {
    this.dimensionFilterManagers = [];
    this.temporaryFilter = undefined;
    console.log("clear", [...this.dimensionFilterManagers]);
  }
}
