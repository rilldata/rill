import { page } from "$app/state";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { DimensionFilterMode } from "@rilldata/web-common/features/dashboards/filters/dimension-filters/constants.ts";
import type {
  MetricsViewSpecDimension,
  V1Expression,
} from "@rilldata/web-common/runtime-client";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import {
  createInExpression,
  createLikeExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils.ts";

export type DimensionFilterManagerInit = {
  mode: DimensionFilterMode;
  selectedValues?: string[];
  inputText?: string;
  exclude?: boolean;
  pinned?: boolean;
  required?: boolean;
};

export class DimensionFilterManager {
  public expr: V1Expression | undefined = $state(undefined);

  public mode: DimensionFilterMode;
  public selectedValues: string[];
  public inputText: string;
  public exclude: boolean;
  public pinned: boolean;
  public required: boolean;

  public constructor(
    public readonly name: string,
    public readonly label: string,
    public readonly dimensions: Map<string, MetricsViewSpecDimension>,
    public readonly editing: boolean, // TODO: maybe separate editing, pinned & required?
    init: DimensionFilterManagerInit,
  ) {
    this.mode = $state(init.mode);
    this.selectedValues = $state(init.selectedValues ?? []);
    this.inputText = $state(init.inputText ?? "");
    this.exclude = $state(init.exclude ?? false);
    this.pinned = $state(init.pinned ?? false);
    this.required = $state(init.required ?? false);
    this.buildExpression();
  }

  public toggleMultipleValues(
    dimensionValues: string[],
    isExclusiveFilter?: boolean,
    exclude?: boolean,
  ) {
    if (exclude !== undefined) this.exclude = exclude;

    if (this.mode !== DimensionFilterMode.Select) {
      this.inputText = "";
      eventBus.emit("notification", {
        message: "Converted filter type to Select",
        link: {
          text: m.common_undo(),
          href: page.url.href,
        },
      });
      this.mode = DimensionFilterMode.Select;
      this.selectedValues = dimensionValues;
      return this.buildExpression();
    }

    if (isExclusiveFilter) {
      this.selectedValues = dimensionValues;
      return this.buildExpression();
    }

    dimensionValues.forEach((v) => {
      const removedIndex = this.toggleValue(v, false);
      if (removedIndex === -1) return;

      // TODO: decrement pinIndex if the removed value was before the pinned value
    });

    this.buildExpression();
  }

  public toggleValue(dimensionValue: string, isExclusiveFilter: boolean) {
    const inIdx = this.selectedValues.findIndex((v) => v === dimensionValue);
    let retIdx = inIdx;

    if (inIdx === -1) {
      if (isExclusiveFilter) {
        this.selectedValues = [dimensionValue];
        retIdx = -1;
      } else {
        this.selectedValues = [...this.selectedValues, dimensionValue];
      }
    } else {
      this.selectedValues = this.selectedValues.splice(inIdx, 1);
    }

    return retIdx;
  }

  public setInList(values: string[], exclude?: boolean) {
    if (exclude !== undefined) this.exclude = exclude;

    this.mode = DimensionFilterMode.InList;
    this.selectedValues = values;
    this.inputText = "";
    this.buildExpression();
  }

  public setContainsText(searchText: string, exclude?: boolean) {
    if (exclude !== undefined) this.exclude = exclude;

    this.mode = DimensionFilterMode.Contains;
    this.selectedValues = [];
    this.inputText = searchText;
    this.buildExpression();
  }

  public toggleExclude() {
    this.exclude = !this.exclude;
  }

  public togglePinned() {
    this.pinned = !this.pinned;
  }

  public toggleRequired() {
    this.required = !this.required;
  }

  public clear() {
    this.selectedValues = [];
    this.inputText = "";
    this.expr = undefined;
    this.buildExpression();
  }

  private buildExpression() {
    switch (this.mode) {
      case DimensionFilterMode.Select:
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
  }
}
