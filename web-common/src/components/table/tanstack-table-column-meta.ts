import type { RowData } from "tanstack-table-8-svelte-5";
import type { ComponentType, SvelteComponent } from "svelte";

declare module "tanstack-table-8-svelte-5" {
  // eslint-disable-next-line @typescript-eslint/no-unused-vars
  interface ColumnMeta<TData extends RowData, TValue> {
    widthPercent?: number;
    marginLeft?: string;
    icon?: ComponentType<SvelteComponent>;
    /** Full path of dimension name/value pairs from root to this header */
    dimensionPath?: Record<string, string>;
    tooltipFormatter?: (value: unknown) => string | null | undefined;
  }
}
