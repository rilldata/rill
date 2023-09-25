import { createLoadingCell } from "@rilldata/web-common/features/dashboards/pivot/util";
import { getRowHeaderKeysFromPos } from "./query-keys";
import { get } from "svelte/store";

export function createMouseDownHandler({ queryClient, config, cache, getPos }) {
  return function handleMouseDown(event: MouseEvent, table) {
    if (event.target.hasAttribute("data-expandable")) {
      const meta = table.getMeta(event.target.parentNode);
      config.update((c) => {
        const existingKeys = getRowHeaderKeysFromPos(
          getPos(),
          JSON.stringify(get(config))
        );

        let action = {
          type: "",
          idx: -1,
        };
        if (meta.value.isExpanded) {
          c.expanded = c.expanded.filter((idx) => idx !== meta.value.idx);
          action.type = "collapse";
          action.idx = meta.value.idx;
        } else {
          c.expanded.push(meta.value.idx);
          action.type = "expand";
          action.idx = meta.value.idx;
        }

        // Attempt to use setQueryData to optimistically update with loaders
        const nextKeys = existingKeys.map((existingKey) => [
          ...existingKey.slice(0, 1),
          JSON.stringify(c),
          ...existingKey.slice(2),
        ]);

        const prevData = existingKeys.map((key) =>
          queryClient.getQueryData(key)
        );

        const nextData = prevData.map((cache: any) => {
          let data = structuredClone(cache.data);
          if (action.type === "collapse") {
            const parentRow = data.find((r) => r[0]?.idx === action.idx);
            if (parentRow) {
              parentRow[0].isExpanded = false;
            }
            data = data.filter((r) => r[1]?.parentIdx !== action.idx);
          } else if (action.type === "expand") {
            const targetIdx = data.findIndex((r) => r[0]?.idx === action.idx);
            if (targetIdx > -1) {
              data[targetIdx][0].isExpanded = true;
              data.splice(targetIdx + 1, 0, ["", createLoadingCell()]);
            }
          }
          return {
            ...cache,
            data,
          };
        });
        // Prepopulate for optimistic update
        const keysToPrepopulate = nextKeys.filter((key) => !cache.find(key));
        keysToPrepopulate.forEach((key, i) => {
          queryClient.setQueryData(key, nextData[i]);
        });
        // Invalidate so optimistic updates get replaced with real data
        setTimeout(() => {
          keysToPrepopulate.forEach((key) => {
            queryClient.invalidateQueries(key);
          });
        });

        return c;
      });
    }
  };
}
