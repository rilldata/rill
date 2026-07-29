import type {
  MetricsViewSpecMeasure,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import {
  mapExprToMeasureFilter,
  mapMeasureFilterToExpr,
  type MeasureFilterEntry,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry.ts";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options.ts";
import {
  createSubQueryExpression,
  removeWrapperAndOrExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";

export class MeasureFilterManager {
  public expr: V1Expression | undefined = $state(undefined);

  public dimension: string;
  public operation: MeasureFilterOperation;
  public type: MeasureFilterType;
  public value1: string;
  public value2: string;

  public constructor(
    public readonly name: string,
    public readonly label: string,
    public readonly measures: Record<string, MetricsViewSpecMeasure>,
    initDimension: string = "",
    initExpr: V1Expression | undefined = undefined,
  ) {
    const unwrappedHavingFilter = removeWrapperAndOrExpression(
      initExpr?.subquery?.having,
    );
    const mappedMeasureFilter = mapExprToMeasureFilter(unwrappedHavingFilter);

    this.dimension = $state(initDimension);
    this.operation = $state(
      mappedMeasureFilter?.operation ?? MeasureFilterOperation.LessThan,
    );
    this.type = $state(mappedMeasureFilter?.type ?? MeasureFilterType.Value);
    this.value1 = $state(mappedMeasureFilter?.value1 ?? "");
    this.value2 = $state(mappedMeasureFilter?.value2 ?? "");
    this.commit();
  }

  public apply(dimension: string, newFilter: MeasureFilterEntry) {
    this.dimension = dimension;
    this.operation = newFilter.operation;
    this.type = newFilter.type;
    this.value1 = newFilter.value1;
    this.value2 = newFilter.value2;
    this.commit();
  }

  public clear() {
    this.expr = undefined;
    this.commit();
  }

  public commit() {
    const measureFilterExpr = mapMeasureFilterToExpr({
      measure: this.name,
      operation: this.operation,
      type: this.type,
      value1: this.value1,
      value2: this.value2,
    });
    const hasFilter = Boolean(this.dimension && measureFilterExpr);
    this.expr = hasFilter
      ? createSubQueryExpression(this.dimension, [this.name], measureFilterExpr)
      : undefined;
  }
}
