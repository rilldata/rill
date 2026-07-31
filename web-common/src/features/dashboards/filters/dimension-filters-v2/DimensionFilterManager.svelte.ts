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

export class DimensionFilterManager {
  public expr: V1Expression | undefined = $state(undefined);

  public mode = $state(DimensionFilterMode.Select);
  public selectedValues = $state<string[]>([]);
  public inputText = $state("");
  public exclude = $state(false);

  private oldMode: DimensionFilterMode;

  public constructor(
    public readonly name: string,
    public readonly label: string,
    initExpr: V1Expression = createInExpression(name, []),
    isInList: boolean = false,
    // Filter dropdown doesnt immediately apply changes.
    // This marks this manager as ephemeral, it will not notify about mode changes.
    private readonly ephemeral: boolean = false,
  ) {
    this.reconcile(initExpr, isInList ? [name] : []);
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
    this.expr = undefined;
    this.commit();
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
            message: "Converted filter type to Select",
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
  }
}

export function sanitizeSearchText(searchText: string) {
  return searchText.replace(/^%/, "").replace(/%$/, "");
}
