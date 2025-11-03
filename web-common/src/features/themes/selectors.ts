import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { readable, derived, type Readable } from "svelte/store";
import { Theme } from "./theme";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { CanvasResponse } from "../canvas/selector";
import type { ExploreValidSpecResponse } from "../explores/selectors";

export function useTheme(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Theme);
}

export function createResolvedThemeStore(
  urlThemeName: Readable<string | undefined | null>,
  query: Readable<
    QueryObserverResult<CanvasResponse | ExploreValidSpecResponse, RpcStatus>
  >,
  instanceId: string,
): Readable<Theme | undefined> {
  const inputs = derived([urlThemeName, query], ([$url, $query]) => {
    const res = $query.data;
    const themeName =
      res && "canvas" in res
        ? res.canvas?.theme
        : res && "explore" in res
          ? res.explore?.theme
          : undefined;
    const embeddedTheme =
      res && "canvas" in res
        ? res.canvas?.embeddedTheme
        : res && "explore" in res
          ? res.explore?.embeddedTheme
          : undefined;

    const name = $url || themeName?.trim() || undefined;
    return { name, embedded: embeddedTheme };
  });

  return readable<Theme | undefined>(undefined, (set) => {
    let stopInner: (() => void) | undefined;
    let token = 0;

    const stop = inputs.subscribe(({ name, embedded }) => {
      if (stopInner) {
        stopInner();
        stopInner = undefined;
      }

      if (name) {
        const t = ++token;
        const q = useTheme(instanceId, name);

        stopInner = q.subscribe((resp) => {
          if (t !== token) return;
          const spec = resp?.data?.theme?.spec;
          if (spec) set(new Theme(spec));
        });

        return;
      }

      if (embedded) {
        set(new Theme(embedded));
        return;
      }

      set(undefined);
    });

    return () => {
      stop();
      if (stopInner) stopInner();
    };
  });
}
