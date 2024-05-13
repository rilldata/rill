import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";
import { V1MetricsViewColumn } from "@rilldata/web-common/runtime-client";

export class VirtualizedTableColumnSizes {
  private readonly sizesCache = new Map<string, Map<string, number>>();

  public get(
    key: string,
    columns: (VirtualizedTableColumns | V1MetricsViewColumn)[],
    columnAccessor: keyof VirtualizedTableColumns,
    calculator: () => number[],
  ): number[] {
    const cache = this.sizesCache.get(key);
    if (!cache) {
      const sizes = calculator();
      this.sizesCache.set(
        key,
        new Map<string, number>(
          sizes.map((s, i) => [columns[i][columnAccessor], s]),
        ),
      );
      return sizes;
    }

    let missingSize = false;
    const sizes = columns.map((column) => {
      if (!cache.has(column[columnAccessor])) {
        missingSize = true;
        return 0;
      }
      return cache.get(column[columnAccessor]) as number;
    });
    if (!missingSize) return sizes;

    const newSizes = calculator();
    // retain sizes from cache
    newSizes.forEach((_, i) => {
      if (cache.has(columns[i][columnAccessor])) {
        newSizes[i] = cache.get(columns[i][columnAccessor]) as number;
      }
    });
    // reset the cache
    this.sizesCache.set(
      key,
      new Map<string, number>(
        newSizes.map((s, i) => [columns[i][columnAccessor], s]),
      ),
    );

    return newSizes;
  }

  public set(name: string, colName: string, value: number) {
    this.sizesCache.get(name)?.set(colName, value);
  }
}
