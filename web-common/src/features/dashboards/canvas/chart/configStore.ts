import { writable } from "svelte/store";
import type { ChartTypeConfig } from "../types";

export const chartConfig = writable<ChartTypeConfig>({
  data: {},
});

export function updateAxis(axis: "x" | "y" | "color", fieldName: string) {
  chartConfig.update((config) => {
    const type = axis === "y" ? "quantitative" : "nominal";
    return {
      ...config,
      data: {
        ...config.data,
        [axis]: { field: fieldName, type },
      },
    };
  });
}
