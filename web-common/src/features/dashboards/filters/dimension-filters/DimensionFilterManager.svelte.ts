import { page } from "$app/state";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import {
  type V1Expression,
  V1Operation,
} from "@rilldata/web-common/runtime-client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import {
  createInExpression,
  createLikeExpression,
  getValuesInExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";
import { convertExpressionToFilterParam } from "@rilldata/web-common/features/dashboards/url-state/filters/converters.ts";
import type { MetricsViewsProvider } from "@rilldata/web-common/features/metrics-views/providers/MetricsViewsProvider.svelte.ts";
import { getDimensionDisplayName } from "@rilldata/web-common/features/dashboards/filters/getDisplayName.ts";
import type {
  FilterChangeSource,
  FilterEventEmitter,
} from "@rilldata/web-common/features/dashboards/filters/filter-events.ts";

export class DimensionFilterManager {
  public expr: V1Expression | undefined = $state(undefined);
  // String representation of the filter expression. Used to check duplicate expressions across metrics views.
  public param: string = $state("");

  public mode = $state(DimensionFilterMode.Select);
  public selectedValues = $state<string[]>([]);
  public inputText = $state("");
  public exclude = $state(false);

  private oldMode: DimensionFilterMode;
  // Cause of the change currently being applied. See `runWithSource`.
  private activeSource: FilterChangeSource;

  public constructor(
    public readonly name: string,
    public readonly label: string,
    initExpr: V1Expression = createInExpression(name, []),
    isInList: boolean = false,
    // Filter dropdown doesnt immediately apply changes.
    // This marks this manager as ephemeral, it will not notify about mode changes.
    private readonly ephemeral: boolean = false,
    // Emitter shared by every manager under an `ExpressionFilterManager`.
    // A manager built outside one, the ephemeral clone for the dropdown for example, has none.
    private readonly events: FilterEventEmitter | undefined = undefined,
  ) {
    this.reconcile(initExpr, isInList ? [name] : []);
  }

  public static createForMetricsViews(
    metricsViewsProvider: MetricsViewsProvider,
    name: string,
    {
      metricsViewName,
      initExpr,
      isInList,
      events,
    }: {
      metricsViewName?: string;
      initExpr?: V1Expression;
      isInList?: boolean;
      events?: FilterEventEmitter;
    } = {},
  ) {
    const dimensionSpecs = metricsViewsProvider.dimensionSpecs[name];
    if (!dimensionSpecs) return undefined;
    const dimensionSpec = metricsViewName
      ? dimensionSpecs[metricsViewName]
      : Object.values(dimensionSpecs)[0];
    if (!dimensionSpec) return undefined;

    return new DimensionFilterManager(
      name,
      getDimensionDisplayName(dimensionSpec),
      initExpr,
      isInList,
      false, // not ephemeral
      events,
    );
  }

  public reconcile(expr: V1Expression, inList: string[]) {
    let initMode: DimensionFilterMode = DimensionFilterMode.Select;
    let initSelectedValues: string[] = [];
    let initInputText: string = "";
    let initExclude: boolean = false;

    const op = expr.cond?.op;
    if (op === V1Operation.OPERATION_IN || op === V1Operation.OPERATION_NIN) {
      initMode = inList.includes(this.name)
        ? DimensionFilterMode.InList
        : DimensionFilterMode.Select;
      initSelectedValues = getValuesInExpression(expr);
      initExclude = op === V1Operation.OPERATION_NIN;
    } else if (
      op === V1Operation.OPERATION_LIKE ||
      op === V1Operation.OPERATION_NLIKE
    ) {
      initMode = DimensionFilterMode.Contains;
      initInputText = sanitizeSearchText(
        expr.cond?.exprs?.[1]?.val?.toString?.() ?? "",
      );
      initExclude = op === V1Operation.OPERATION_NLIKE;
    }

    this.mode = initMode;
    this.oldMode = initMode;
    this.selectedValues = initSelectedValues;
    this.inputText = initInputText;
    this.exclude = initExclude;
    this.commit(false);
  }

  public clone() {
    return new DimensionFilterManager(
      this.name,
      this.label,
      this.expr,
      this.mode === DimensionFilterMode.InList,
      true,
    );
  }

  public apply(dimensionManager: DimensionFilterManager) {
    this.mode = dimensionManager.mode;
    this.selectedValues = [...dimensionManager.selectedValues];
    this.inputText = dimensionManager.inputText;
    this.exclude = dimensionManager.exclude;
    this.commit();
  }

  public setSelectedValues(dimensionValues: string[], exclude: boolean) {
    this.mode = DimensionFilterMode.Select;
    this.selectedValues = dimensionValues;
    this.inputText = "";
    this.exclude = exclude;
    this.commit();
  }

  public toggleValue(dimensionValue: string, isExclusiveFilter: boolean) {
    const inIdx = this.selectedValues.findIndex((v) => v === dimensionValue);

    if (inIdx === -1) {
      if (isExclusiveFilter) {
        this.selectedValues = [dimensionValue];
      } else {
        this.selectedValues = [...this.selectedValues, dimensionValue];
      }
    } else {
      this.selectedValues = this.selectedValues.toSpliced(inIdx, 1);
    }
    this.commit();
  }

  public appendSelectedValues(dimensionValues: string[]) {
    const newValues = dimensionValues.filter(
      (v) => !this.selectedValues.includes(v),
    );
    this.selectedValues = [...this.selectedValues, ...newValues];
    this.commit();
    return newValues;
  }

  public removeSelectedValues(dimensionValues: string[]) {
    this.selectedValues = this.selectedValues.filter(
      (v) => !dimensionValues.includes(v),
    );
    this.commit();
  }

  public setInList(values: string[], exclude: boolean) {
    this.mode = DimensionFilterMode.InList;
    this.selectedValues = values;
    this.inputText = "";
    this.exclude = exclude;
    this.commit();
  }

  public setContainsText(searchText: string, exclude: boolean) {
    this.mode = DimensionFilterMode.Contains;
    this.selectedValues = [];
    this.inputText = searchText;
    this.exclude = exclude;
    this.commit();
  }

  public toggleExclude() {
    this.exclude = !this.exclude;
    this.commit();
  }

  public clear() {
    this.selectedValues = [];
    this.inputText = "";
    this.commit();
  }

  /**
   * Runs `fn` with `source` recorded as the cause of whatever it changes.
   * The filter bar mutates the managers directly, so only the callers that are not
   * the filter bar, a chart click to filter for example, have a source to record.
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
    switch (this.mode) {
      case DimensionFilterMode.Select:
        if (
          this.oldMode !== DimensionFilterMode.Select &&
          !this.ephemeral &&
          notify
        ) {
          eventBus.emit("notification", {
            message: m.filter_converted_to_select(),
            link: {
              text: m.common_undo(),
              href: page.url.href,
            },
          });
        }
      // eslint-disable-next-line no-fallthrough
      case DimensionFilterMode.InList:
        this.expr = this.selectedValues.length
          ? createInExpression(this.name, this.selectedValues, this.exclude)
          : undefined;
        break;

      case DimensionFilterMode.Contains:
        this.expr = this.inputText
          ? createLikeExpression(this.name, `%${this.inputText}%`, this.exclude)
          : undefined;
        break;
    }
    this.oldMode = this.mode;

    this.param = this.expr
      ? convertExpressionToFilterParam(
          this.expr,
          this.mode === DimensionFilterMode.InList ? [this.name] : [],
        )
      : "";

    // `notify` is false while the filter param is being parsed, which rebuilds every manager.
    // Reporting those would make the filter that is already applied look like a fresh change.
    if (notify) {
      this.events?.emit("filter-changed", { source: this.activeSource });
    }
  }
}

export function sanitizeSearchText(searchText: string) {
  return searchText.replace(/^%/, "").replace(/%$/, "");
}
