import { get, writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { featureFlags } from "../feature-flags";

class ThemeControl {
  private preferenceStore = localStorageStore<"light" | "dark" | "system">(
    "rill:theme",
    "light",
  );
  private current = writable<"light" | "dark" | "system">("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");

  constructor() {
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
  }

  public subscribe = this.current.subscribe;
  public _preference = { subscribe: this.preferenceStore.subscribe };

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
    if (get(featureFlags.darkMode) === false) return;
    this.current.set("dark");
    document.documentElement.classList.add("dark");
  }
  private removeDark() {
    this.current.set("light");
    document.documentElement.classList.remove("dark");
  }
}

export const themeControl = new ThemeControl();
