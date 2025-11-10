import {
  ResourceKind,
  useResource,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import { derived, type Readable } from "svelte/store";
import { Theme } from "./theme";
import type { RpcStatus } from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { CanvasResponse } from "../canvas/selector";
import type { ExploreValidSpecResponse } from "../explores/selectors";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

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

  // Create a derived store that reactively fetches the theme based on the theme name
  return derived(themeInput, ({ name, embedded }, set) => {
    // Case 1: Embedded theme (inline theme definition)
    if (embedded) {
      set(new Theme(embedded));
      return;
    }

    // Case 2: Named theme (reference to theme resource)
    if (name) {
      const themeQuery = useResource(
        instanceId,
        name,
        ResourceKind.Theme,
        undefined,
        queryClient,
      );
      return themeQuery.subscribe(($themeQuery) => {
        if ($themeQuery.data?.theme?.spec) {
          set(new Theme($themeQuery.data.theme.spec));
        } else {
          set(undefined);
        }
      });
    }

    // Case 3: No theme
    set(undefined);
  });
}
