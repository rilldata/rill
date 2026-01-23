import { goto } from "$app/navigation";
import { page } from "$app/stores";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus.ts";
import type { PageContentResized } from "@rilldata/web-common/lib/event-bus/events.ts";
import { Throttler } from "@rilldata/web-common/lib/throttler.ts";
import { get } from "svelte/store";
import {
  emitNotification,
  registerRPCMethod,
} from "@rilldata/web-common/lib/rpc";
import { themeControl } from "@rilldata/web-common/features/themes/theme-control";
import { getEmbedThemeStoreInstance } from "@rilldata/web-common/features/embeds/embed-theme";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
import {
  chatOpen,
  sidebarActions,
} from "@rilldata/web-common/features/chat/layouts/sidebar/sidebar-store";

const STATE_CHANGE_THROTTLE_TIMEOUT = 200;
const RESIZE_THROTTLE_TIMEOUT = 200;

export default function initEmbedPublicAPI(): () => void {
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
    return { open: get(chatOpen) };
  });

  registerRPCMethod("setAiPane", (open: boolean) => {
    if (typeof open !== "boolean") {
      throw new Error("Expected open to be a boolean");
    }
    if (open) {
      sidebarActions.openChat();
    } else {
      sidebarActions.closeChat();
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

  return () => {
    unsubscribe();
    resizeUnsub();
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
  "theme",
  "theme_mode",
];
export function removeEmbedParams(searchParams: URLSearchParams) {
  const cleanedParams = new URLSearchParams(searchParams);
  EmbedParams.forEach((param) => cleanedParams.delete(param));
  const search = cleanedParams.toString();
  return decodeURIComponent(search);
}
