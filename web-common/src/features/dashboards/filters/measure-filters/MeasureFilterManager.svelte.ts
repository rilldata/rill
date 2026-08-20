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
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { getMeasureDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type {
  FilterChangeSource,
  FilterEventEmitter,
} from "@rilldata/web-common/features/dashboards/filters/filter-events.ts";

export class MeasureFilterManager {
  public expr: V1Expression | undefined = $state(undefined);
  // String representation of the filter expression. Used to check duplicate expressions across metrics views.
  public param: string = $state("");

  public dimension = $state("");
  public operation = $state(MeasureFilterOperation.LessThan);
  public type = $state(MeasureFilterType.Value);
  public value1 = $state("");
  public value2 = $state("");

  // Cause of the change currently being applied. See `runWithSource`.
  private activeSource: FilterChangeSource;

  public constructor(
    public readonly name: string,
    public readonly label: string,
    initExpr: V1Expression | undefined = undefined,
    // Emitter shared by every manager under an `ExpressionFilterManager`.
    // A manager built outside one has none.
    private readonly events: FilterEventEmitter | undefined = undefined,
  ) {
    this.reconcile(initExpr);
  }

  public static createForMetricsViews(
    metricsViewsProvider: MetricsViewsProvider,
    name: string,
    {
      metricsViewName,
      initExpr,
      events,
    }: {
      metricsViewName?: string;
      initExpr?: V1Expression;
      events?: FilterEventEmitter;
    } = {},
  ) {
    const measureSpecs = metricsViewsProvider.measureSpecs[name];
    if (!measureSpecs) return undefined;
    const measureSpec = metricsViewName
      ? measureSpecs[metricsViewName]
      : Object.values(measureSpecs)[0];
    if (!measureSpec) return undefined;

    return new MeasureFilterManager(
      name,
      getMeasureDisplayName(measureSpec),
      initExpr,
      events,
    );
  }

  public reconcile(expr: V1Expression | undefined) {
    const dimension = expr?.subquery?.dimension;

    const unwrappedHavingFilter = removeWrapperAndOrExpression(
      expr?.subquery?.having,
    );
    const mappedMeasureFilter = mapExprToMeasureFilter(unwrappedHavingFilter);

    this.dimension = dimension ?? "";
    this.operation =
      mappedMeasureFilter?.operation ?? MeasureFilterOperation.LessThan;
    this.type = mappedMeasureFilter?.type ?? MeasureFilterType.Value;
    this.value1 = mappedMeasureFilter?.value1 ?? "";
    this.value2 = mappedMeasureFilter?.value2 ?? "";
    this.commit(false);
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
    this.value1 = "";
    this.value2 = "";
    this.dimension = "";
    this.commit();
  }

  /**
   * Runs `fn` with `source` recorded as the cause of whatever it changes.
   * The filter bar mutates the managers directly, so only the callers that are not
   * the filter bar have a source to record.
   */
  public runWithSource<T>(source: FilterChangeSource, fn: () => T): T {
    this.activeSource = source;
    try {
      return fn();
    } finally {
      this.activeSource = undefined;
    }
  }

  public commit(notify: boolean = true) {
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
    this.param = this.expr ? convertExpressionToFilterParam(this.expr, []) : "";

    // `notify` is false while the filter param is being parsed, which rebuilds every manager.
    // Reporting those would make the filter that is already applied look like a fresh change.
    if (notify) {
      this.events?.emit("filter-changed", { source: this.activeSource });
    }
  }
}
