import { get, writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { featureFlags } from "../feature-flags";

type Theme = "light" | "dark" | "system";

class ThemeControl {
  private current = writable<Theme>("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");
  private preferenceStore = localStorageStore<Theme>("rill:theme", "light");

  public subscribe = this.current.subscribe;
  public preference = { subscribe: this.preferenceStore.subscribe };

  constructor() {
    this.init().catch((error) => {
      console.error("Failed to initialize theme control:", error);
    });
  }

  init = async () => {
    const currentPreference = get(this.preferenceStore);

    await featureFlags.ready;

    if (
      (get(featureFlags.darkMode) && currentPreference === "dark") ||
      (currentPreference === "system" && this.darkQuery.matches)
    ) {
      this.setDark();
    }

    this.darkQuery.addEventListener("change", ({ matches }) => {
      if (get(this.preferenceStore) !== "system") return;

      if (matches && get(featureFlags.darkMode)) {
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
}

export const themeControl = new ThemeControl();
