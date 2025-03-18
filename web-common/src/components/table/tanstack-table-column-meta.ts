import type { RowData } from "@tanstack/svelte-table";
import type { ComponentType, SvelteComponent } from "svelte";

declare module "@tanstack/svelte-table" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    widthPercent?: number;
    marginLeft?: string;
    icon?: ComponentType<SvelteComponent>;
  }
}
