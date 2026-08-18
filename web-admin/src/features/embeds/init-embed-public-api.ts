import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { getDashboardFromEmbedRoute } from "@rilldata/web-admin/features/embeds/embed-route-utils.ts";
import { EmbedStorageNamespacePrefix } from "@rilldata/web-admin/features/embeds/constants.ts";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
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

const STATE_CHANGE_THROTTLE_TIMEOUT = 200;
const RESIZE_THROTTLE_TIMEOUT = 200;
const AI_PANE_CHANGE_THROTTLE_TIMEOUT = 200;

type SetValidStateParams = {
  state: string;
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

    // Upfront validation is only supported for explore dashboards.
    // For anything else (e.g. canvas) fall back to applying the state as-is.
    if (activeDashboard?.kind !== ResourceKind.Explore) {
      const currentUrl = new URL(pageState.url);
      currentUrl.search = state;
      void goto(currentUrl, { replaceState: true });
      return { success: true, appliedState: state, errors: [] };
    }

    const { url, errors } = await buildValidatedExploreUrl(
      client,
      activeDashboard.name,
      new URLSearchParams(state),
      pageState.url,
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
    clearExploreSessionStore(activeDashboard.name, EmbedStorageNamespacePrefix);

    const currentUrl = new URL(pageState.url);
    currentUrl.search = url.toString();
    void goto(currentUrl, { replaceState: true });
    return {
      success: true,
      appliedState: currentUrl.search.replace(/^\?/, ""),
      errors: errorMessages,
    };
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
