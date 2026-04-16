import { get, writable } from "svelte/store";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";
import { explicitLocalStorageStore } from "@rilldata/web-common/lib/store-utils/local-storage.ts";

export type ThemeMode = "light" | "dark" | "system";

function isEmbedEnvironment(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.location.pathname.includes("/-/embed");
  } catch {
    return false;
  }
}

const THEME_LOCAL_STORAGE_KEY = "rill:theme";
const THEME_SESSION_STORAGE_KEY = "rill:embed:theme-mode";

class ThemeControl {
  public current = writable<"light" | "dark">("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
  private preferenceStore = isEmbedEnvironment()
    ? sessionStorageStore<ThemeMode>(THEME_SESSION_STORAGE_KEY, "light")
    : explicitLocalStorageStore<ThemeMode>(THEME_LOCAL_STORAGE_KEY, "light");

  public subscribe = this.current.subscribe;
  public preference = { subscribe: this.preferenceStore.subscribe };

  constructor() {
    this.init();
  }

  init = () => {
    const currentPreference = get(this.preferenceStore);

    if (
      currentPreference === "dark" ||
      (currentPreference === "system" && this.darkQuery.matches)
    ) {
      this.setDark();
    }

    this.darkQuery.addEventListener("change", ({ matches }) => {
      if (get(this.preferenceStore) !== "system") return;

      if (matches) {
        this.setDark();
      } else {
        this.removeDark();
      }
    });
  };

  public set: Record<ThemeMode, () => void> = {
    light: () => {
      this.preferenceStore.set("light");
      this.removeDark();
    },
    dark: () => {
      this.preferenceStore.set("dark");
      this.setDark();
    },
    system: () => {
      this.preferenceStore.set("system");

      if (this.darkQuery.matches) {
        this.setDark();
      } else {
        this.removeDark();
      }
    },
  };

  private setDark() {
    this.current.set("dark");
    document.documentElement.classList.add("dark");
  }

  private removeDark() {
    this.current.set("light");
    document.documentElement.classList.remove("dark");
  }
}

export const themeControl = new ThemeControl();

/**
 * Returns true if the user needs to select a theme — i.e. no theme has been
 * persisted to localStorage yet. Always returns false in the embed context,
 * which manages its own ephemeral theme preference.
 */
export function isThemeSelectionNeeded(): boolean {
  if (isEmbedEnvironment()) return false;
  try {
    return !localStorage.getItem(THEME_LOCAL_STORAGE_KEY);
  } catch {
    return false;
  }
}
