import { get, writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";

class ThemeControl {
  private preference = localStorageStore<"light" | "dark" | "system">(
    "theme",
    "light",
  );
  private current = writable<"light" | "dark" | "system">("light");
  private darkQuery = window.matchMedia("(prefers-color-scheme: dark)");

  constructor() {
    this.darkQuery.addEventListener("change", ({ matches }) => {
      if (get(this.preference) !== "system") return;

      if (matches) {
        this.setDark();
      } else {
        this.removeDark();
      }
    });

    if (get(this.preference) === "dark" || this.darkQuery.matches) {
      this.setDark();
    }
  }

  public subscribe = this.current.subscribe;

  public set = {
    light: () => {
      this.preference.set("light");
      this.removeDark();
    },
    dark: () => {
      this.preference.set("dark");
      this.setDark();
    },
    system: () => {
      this.preference.set("system");

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
