import { goto } from "$app/navigation";
import { page } from "$app/stores";
import {
  type DashboardInfo,
  getDashboardFromEmbedRoute,
} from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
import {
  fetchResources,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { buildValidatedExploreUrl } from "@rilldata/web-common/features/dashboards/state-managers/loaders/build-validated-explore-url.ts";
import { clearExploreSessionStore } from "@rilldata/web-common/features/dashboards/state-managers/loaders/explore-web-view-store.ts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import type { PageContentResized } from "@rilldata/web-common/lib/event-bus/events.ts";
import { Throttler } from "@rilldata/web-common/lib/throttler.ts";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";
import {
  emitNotification,
  registerRPCMethod,
} from "@rilldata/web-common/lib/rpc";
import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
import { getEmbedThemeStoreInstance } from "@rilldata/web-common/features/embeds/embed-theme";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
import {
  dashboardChatActions,
  dashboardChatOpen,
} from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";

const STATE_CHANGE_THROTTLE_TIMEOUT = 200;
const RESIZE_THROTTLE_TIMEOUT = 200;
const AI_PANE_CHANGE_THROTTLE_TIMEOUT = 200;

// The resource kinds an embed can render as a dashboard.
const DashboardResourceKinds = new Set<string>([
  ResourceKind.Explore,
  ResourceKind.Canvas,
]);

type SetValidStateParams = {
  state: string;
  // When false (the default) the cleaned state is applied even if some params were invalid.
  // When true the state is only applied if there are no validation errors.
  failOnError?: boolean;
};

type NavigateToDashboardParams = {
  name: string;
  // State is optional. Uses default behaviour of navigation when not specified.
  state?: string;
  // When false (the default) the cleaned state is applied even if some params were invalid.
  // When true the state is only applied if there are no validation errors.
  failOnError?: boolean;
};

export default function initEmbedPublicAPI(client: RuntimeClient): () => void {
  const embedThemeStore = getEmbedThemeStoreInstance();

  const embedStore = EmbedStore.getInstance();
  const themeModeFromUrl = embedStore?.themeMode;
  if (themeModeFromUrl) {
    if (
      themeModeFromUrl === "dark" ||
      themeModeFromUrl === "light" ||
      themeModeFromUrl === "system"
    ) {
      themeControl.set[themeModeFromUrl]();
    }
  } else {
    themeControl.set.light();
  }

  registerRPCMethod("getState", () => {
    const { url } = get(page);
    return { state: removeEmbedParams(url.searchParams) };
  });

  registerRPCMethod("setState", (state: string) => {
    if (typeof state !== "string") {
      throw new Error("Expected state to be a string");
    }
    const currentUrl = new URL(get(page).url);
    currentUrl.search = state;
    void goto(currentUrl, { replaceState: true });
    return true;
  });

  registerRPCMethod("setValidState", async (params: SetValidStateParams) => {
    if (
      typeof params !== "object" ||
      params === null ||
      typeof params.state !== "string"
    ) {
      throw new Error(
        "Expected params to be an object with a string `state` property",
      );
    }
    const { state, failOnError = false } = params;

    const pageState = get(page);
    const activeDashboard = getDashboardFromEmbedRoute(
      pageState.route.id,
      pageState.params,
    );

    return applyDashboardState(client, activeDashboard, state, pageState.url, {
      failOnError,
      // The dashboard stays the same, so the state change should not add a history entry.
      replaceState: true,
    });
  });

  registerRPCMethod("getThemeMode", () => {
    return { themeMode: get(themeControl.preference) };
  });

  registerRPCMethod("setThemeMode", (themeMode: string) => {
    if (
      themeMode !== "dark" &&
      themeMode !== "light" &&
      themeMode !== "system"
    ) {
      throw new Error(
        'Expected themeMode to be one of "dark", "light", or "system"',
      );
    }
    if (themeMode === "dark") {
      themeControl.set.dark();
    } else if (themeMode === "light") {
      themeControl.set.light();
    } else {
      themeControl.set.system();
    }
    return true;
  });

  registerRPCMethod("getTheme", () => {
    const theme = get(embedThemeStore);
    return { theme: theme || "default" };
  });

  registerRPCMethod("setTheme", (theme: string | null) => {
    if (theme !== null && typeof theme !== "string") {
      throw new Error("Expected theme to be a string or null");
    }
    const themeValue = !theme || theme === "default" ? null : theme;
    embedThemeStore.set(themeValue);
    return true;
  });

  registerRPCMethod("getAiPane", () => {
    return { open: get(dashboardChatOpen) };
  });

  registerRPCMethod("setAiPane", (open: boolean) => {
    if (typeof open !== "boolean") {
      throw new Error("Expected open to be a boolean");
    }
    if (open) {
      dashboardChatActions.openChat();
    } else {
      dashboardChatActions.closeChat();
    }
    return true;
  });

  // The embed suppresses all navigation when configured with `navigation=false`,
  // so fail loudly rather than accepting a call that would silently do nothing.
  function assertNavigationEnabled() {
    if (!embedStore?.navigationEnabled) {
      throw new Error("Navigation is disabled for this embed");
    }
  }

  registerRPCMethod("navigateBack", () => {
    assertNavigationEnabled();
    window.history.back();
    return true;
  });
  registerRPCMethod("navigateForward", () => {
    assertNavigationEnabled();
    window.history.forward();
    return true;
  });
  registerRPCMethod(
    "navigateToDashboard",
    async (params: NavigateToDashboardParams) => {
      assertNavigationEnabled();
      if (
        typeof params !== "object" ||
        params === null ||
        typeof params.name !== "string"
      ) {
        throw new Error(
          "Expected params to be an object with a string `name` property",
        );
      }

      const { name, state, failOnError = false } = params;
      if (state !== undefined && typeof state !== "string") {
        throw new Error("Expected `state` to be a string");
      }

      const resources = await fetchResources(queryClient, client, true);
      const targetKind = resources.find(
        (r) =>
          r.meta?.name?.name === name &&
          DashboardResourceKinds.has(r.meta?.name?.kind ?? ""),
      )?.meta?.name?.kind as ResourceKind | undefined;
      if (!targetKind) {
        throw new Error(`Dashboard "${name}" not found`);
      }
      const dashboard: DashboardInfo = { name, kind: targetKind };

      const targetUrl = new URL(get(page).url);
      targetUrl.pathname = `/-/embed/${
        targetKind === ResourceKind.Canvas ? "canvas" : "explore"
      }/${encodeURIComponent(name)}`;
      // Params of the dashboard being navigated away from should never carry over.
      targetUrl.search = "";

      if (state === undefined) {
        // Navigate without any state so that the dashboard falls back to its default behaviour.
        void goto(targetUrl);
        return { success: true, appliedState: "", errors: [] };
      }

      return applyDashboardState(client, dashboard, state, targetUrl, {
        failOnError,
        // Navigating to another dashboard should be undoable via `navigateBack`.
        replaceState: false,
      });
    },
  );

  emitNotification("ready");

  const stateChangeThrottler = new Throttler(
    STATE_CHANGE_THROTTLE_TIMEOUT,
    STATE_CHANGE_THROTTLE_TIMEOUT,
  );
  // Keep this at the end so that RPC methods are already available and "ready" has been fired.
  const unsubscribe = page.subscribe(({ url }) => {
    // Throttle the state change event.
    // This avoids too many events being fired when state is changed quickly.
    // This also avoids early events being fired just before dashboard is ready but is routed to.
    stateChangeThrottler.throttle(() => {
      emitNotification("stateChange", {
        state: removeEmbedParams(url.searchParams),
      });
    });
  });

  const resizeThrottler = new Throttler(
    RESIZE_THROTTLE_TIMEOUT,
    RESIZE_THROTTLE_TIMEOUT,
  );
  function onResize(event: PageContentResized) {
    // Throttle the resize event.
    // This avoids too many events being fired when size changes quickly, especially when page is loading.
    resizeThrottler.throttle(() => {
      emitNotification("resized", {
        width: event.width,
        height: event.height,
      });
    });
  }
  const resizeUnsub = eventBus.on("page-content-resized", onResize);
  onResize({
    width: document.body.scrollWidth,
    height: document.body.scrollHeight,
  });

  // Subscribe to AI pane state changes
  const aiPaneChangeThrottler = new Throttler(
    AI_PANE_CHANGE_THROTTLE_TIMEOUT,
    AI_PANE_CHANGE_THROTTLE_TIMEOUT,
  );
  const aiPaneUnsubscribe = dashboardChatOpen.subscribe((isOpen) => {
    aiPaneChangeThrottler.throttle(() => {
      emitNotification("aiPaneChanged", {
        open: isOpen,
      });
    });
  });

  return () => {
    unsubscribe();
    resizeUnsub();
    aiPaneUnsubscribe();
  };
}

type ApplyDashboardStateResult = {
  success: boolean;
  // The state that was actually applied. Omitted when the state was not applied.
  appliedState?: string;
  errors: string[];
};

/**
 * Applies `state` to `targetUrl` and navigates there.
 *
 * For explore dashboards the state is first validated against the explore and metrics view specs,
 * so that invalid params are dropped and the applied url is canonicalized the same way it would be
 * if the user had navigated there directly. Upfront validation is not supported for anything else
 * (e.g. canvas or the dashboard listing), where the state is applied as-is.
 */
async function applyDashboardState(
  client: RuntimeClient,
  dashboard: DashboardInfo | null,
  state: string,
  targetUrl: URL,
  {
    failOnError,
    replaceState,
  }: { failOnError: boolean; replaceState: boolean },
): Promise<ApplyDashboardStateResult> {
  const url = new URL(targetUrl);

  if (dashboard?.kind !== ResourceKind.Explore) {
    url.search = state;
    void goto(url, { replaceState });
    return { success: true, appliedState: state, errors: [] };
  }

  const { url: validatedParams, errors } = await buildValidatedExploreUrl(
    client,
    dashboard.name,
    new URLSearchParams(state),
    targetUrl,
  );
  const errorMessages = errors.map((error) => error.message);

  if (errors.length > 0 && failOnError) {
    return { success: false, errors: errorMessages };
  }

  // Clear any prior embed session state for this explore before navigating.
  // buildValidatedExploreUrl intentionally ignores session storage,
  // but applying the url via goto triggers handleURLChange which re-merges session storage for empty / view-only urls.
  // Without this, resetting the dashboard (e.g. setValidState({ state: "" })) would restore the previous session filters
  // instead of the validated state we just computed and returned.
  clearExploreSessionStore(dashboard.name, EmbedStorageNamespacePrefix);

  url.search = validatedParams.toString();
  void goto(url, { replaceState });
  return {
    success: true,
    appliedState: url.search.replace(/^\?/, ""),
    errors: errorMessages,
  };
}

const EmbedParams = [
  "instance_id",
  "runtime_host",
  "access_token",
  "resource",
  "type",
  "kind",
  "navigation",
  "hide_navigation_bar",
  "theme",
  "theme_mode",
  "external_user_id",
];
export function removeEmbedParams(searchParams: URLSearchParams) {
  const cleanedParams = new URLSearchParams(searchParams);
  EmbedParams.forEach((param) => cleanedParams.delete(param));
  const search = cleanedParams.toString();
  return search;
}
