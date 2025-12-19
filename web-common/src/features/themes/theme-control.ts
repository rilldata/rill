import { get, writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { sessionStorageStore } from "@rilldata/web-common/lib/store-utils/session-storage";

type Theme = "light" | "dark" | "system";

function isEmbedEnvironment(): boolean {
  if (typeof window === "undefined") return false;
  try {
    return window.location.pathname.includes("/-/embed");
  } catch {
    return false;
  }
}

class ThemeControl {
  private current = writable<Theme>("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
  private preferenceStore = isEmbedEnvironment()
    ? sessionStorageStore<Theme>("rill:embed:theme-mode", "light")
    : localStorageStore<Theme>("rill:theme", "light");

  public subscribe = this.current.subscribe;
  public preference = { subscribe: this.preferenceStore.subscribe };

  constructor() {
    const pref = get(this.preferenceStore);

    if (pref === "dark" || (pref === "system" && this.darkQuery.matches)) {
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
  }

  public set = {
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
