import type { V1Expression } from "@rilldata/web-common/runtime-client";
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

  public dimension = $state("");
  public operation = $state(MeasureFilterOperation.LessThan);
  public type = $state(MeasureFilterType.Value);
  public value1 = $state("");
  public value2 = $state("");

  public constructor(
    public readonly name: string,
    public readonly label: string,
    initExpr: V1Expression | undefined = undefined,
  ) {
    this.reconcile(initExpr);
  }

  public reconcile(expr: V1Expression | undefined) {
    const dimension = expr?.cond?.exprs?.[0]?.ident;
    const firstValueExpr = expr?.cond?.exprs?.[1];

    const unwrappedHavingFilter = removeWrapperAndOrExpression(
      firstValueExpr?.subquery?.having,
    );
    const mappedMeasureFilter = mapExprToMeasureFilter(unwrappedHavingFilter);

    this.dimension = dimension ?? "";
    this.operation =
      mappedMeasureFilter?.operation ?? MeasureFilterOperation.LessThan;
    this.type = mappedMeasureFilter?.type ?? MeasureFilterType.Value;
    this.value1 = mappedMeasureFilter?.value1 ?? "";
    this.value2 = mappedMeasureFilter?.value2 ?? "";
    this.commit();
  }

  public setMeasureFilter(dimension: string, newFilter: MeasureFilterEntry) {
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
