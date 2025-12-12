import { get, writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";

type Theme = "light" | "dark" | "system";

class ThemeControl {
  private current = writable<Theme>("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
  private preferenceStore = localStorageStore<Theme>("rill:theme", "light");
  private initialized = false;

  public subscribe = this.current.subscribe;
  public _preference = { subscribe: this.preferenceStore.subscribe };

  constructor() {
    try {
      this.init();
    } catch (error) {
      console.error("Failed to initialize theme control:", error);
    }
  }

  init = () => {
    if (this.initialized) return;
    this.initialized = true;

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

  public ensure = () => {
    void this.init();
  };
}

export const themeControl = new ThemeControl();
