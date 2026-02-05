import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { MetricsViewSpecDimension } from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";
import type { YAMLMap } from "yaml";

export const types: ItemType[] = ["measures", "dimensions"];

export type ItemType = "measures" | "dimensions";

export type MenuOption = { value: string; label: string; type?: string };

export const editingItemData = writable<{
  index: number;
  type: ItemType;
} | null>(null);

export type Confirmation = {
  action: "cancel" | "delete" | "switch";
  type?: ItemType;
  model?: string;
  database?: string;
  connector?: string;
  schema?: string;
  index?: number;
  field?: string;
};

export class YAMLDimension {
  column: string;
  expression: string;
  name: string;
  display_name: string;
  description: string;
  unnest: boolean | undefined;
  resourceName: string;
  type: "time" | "geo" | "categorical" | undefined;

  constructor(
    item?: YAMLMap<string, string>,
    dimension?: MetricsViewSpecDimension,
  ) {
    this.column = item?.get("column") ?? "";
    this.expression = item?.get("expression") ?? "";
    this.name = item?.get("name") ?? "";
    this.display_name = item?.get("display_name") ?? item?.get("label") ?? "";
    this.description = item?.get("description") ?? "";
    this.unnest =
      item?.get("unnest") === undefined
        ? undefined
        : item?.get("unnest") === "true";
    this.resourceName = dimension?.name ?? "";
    this.type = item?.get("type") as "time" | "geo" | "categorical" | undefined;
  }
}

export class YAMLMeasure {
  expression: string;
  name: string;
  display_name: string;
  description: string;
  valid_percent_of_total: boolean;
  format_d3: string;
  format_preset: FormatPreset | "";
  type: "simple" | "derived" | "time_comparison" | undefined;

  constructor(item?: YAMLMap<string, string>) {
    this.expression = item?.get("expression") ?? "";
    this.name = item?.get("name") ?? "";
    this.display_name = item?.get("display_name") ?? item?.get("label") ?? "";
    this.description = item?.get("description") ?? "";
    this.valid_percent_of_total = Boolean(item?.get("valid_percent_of_total"));
    this.format_d3 = item?.get("format_d3") ?? "";
    this.format_preset =
      (item?.get("format_preset") as unknown as FormatPreset) ??
      (this.format_d3 ? "" : FormatPreset.NONE);
    this.type = item?.get("type") as
      | "simple"
      | "derived"
      | "time_comparison"
      | undefined;
  }
}

export const ROW_HEIGHT = 40;
