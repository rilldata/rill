import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { readable, derived, type Readable } from "svelte/store";
import { Theme } from "./theme";
import type {
  RpcStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { CanvasResponse } from "../canvas/selector";
import type { ExploreValidSpecResponse } from "../explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { getRuntimeServiceGetResourceQueryKey } from "@rilldata/web-common/runtime-client";

export function useTheme(instanceId: string, name: string) {
  return useResource(instanceId, name, ResourceKind.Theme);
}

function extractThemeInfo(
  data: CanvasResponse | ExploreValidSpecResponse | undefined,
) {
  if (!data) return { themeName: undefined, embeddedTheme: undefined };

  if ("canvas" in data) {
    return {
      themeName: data.canvas?.theme,
      embeddedTheme: data.canvas?.embeddedTheme,
    };
  }

  if ("explore" in data) {
    return {
      themeName: data.explore?.theme,
      embeddedTheme: data.explore?.embeddedTheme,
    };
  }

  return { themeName: undefined, embeddedTheme: undefined };
}

function getThemeSpecFromCache(
  queryKey: readonly unknown[],
): Theme | undefined {
  const data = queryClient.getQueryData<{ resource?: V1Resource }>(queryKey);
  const spec = data?.resource?.theme?.spec;
  return spec ? new Theme(spec) : undefined;
}

export function createResolvedThemeStore(
  urlThemeName: Readable<string | undefined | null>,
  query: Readable<
    QueryObserverResult<CanvasResponse | ExploreValidSpecResponse, RpcStatus>
  >,
  instanceId: string,
): Readable<Theme | undefined> {
  const themeInput = derived([urlThemeName, query], ([$url, $query]) => {
    const { themeName, embeddedTheme } = extractThemeInfo($query.data);
    const resolvedName = $url || themeName?.trim();

    return {
      name: resolvedName || undefined,
      embedded: embeddedTheme,
    };
  });

  return readable<Theme | undefined>(undefined, (set) => {
    let cleanupQuerySubscription: (() => void) | undefined;

    const unsubscribe = themeInput.subscribe(({ name, embedded }) => {
      // Clean up previous query subscription
      cleanupQuerySubscription?.();
      cleanupQuerySubscription = undefined;

      // Case 1: Embedded theme (inline theme definition)
      if (embedded) {
        set(new Theme(embedded));
        return;
      }

      // Case 2: Named theme (reference to theme resource)
      if (name) {
        const queryKey = getRuntimeServiceGetResourceQueryKey(instanceId, {
          "name.name": name,
          "name.kind": ResourceKind.Theme,
        });
        const queryKeyStr = JSON.stringify(queryKey);

        // Set initial value from cache
        const cachedTheme = getThemeSpecFromCache(queryKey);
        if (cachedTheme) {
          set(cachedTheme);
        }

        // Subscribe to query cache updates for this theme
        cleanupQuerySubscription = queryClient
          .getQueryCache()
          .subscribe((event) => {
            const eventKeyStr = event?.query.queryKey
              ? JSON.stringify(event.query.queryKey)
              : null;

            if (eventKeyStr === queryKeyStr) {
              const theme = getThemeSpecFromCache(queryKey);
              if (theme) {
                set(theme);
              }
            }
          });

        return;
      }

      // Case 3: No theme
      set(undefined);
    });

    return () => {
      unsubscribe();
      cleanupQuerySubscription?.();
    };
  });
}
