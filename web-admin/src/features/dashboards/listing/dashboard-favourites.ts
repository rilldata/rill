import { writable, type Writable } from "svelte/store";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";

function getDashboardFavouritesKey(org: string, project: string) {
  return `rill:app:${org}:${project}:dashboard:favourites`;
}

export class DashboardFavourites {
  public favourites: Writable<string[]>;

  public constructor() {
    this.favourites = writable<string[]>([]);
  }

  public setOrgAndProject(org: string, project: string) {
    this.favourites = localStorageStore(
      getDashboardFavouritesKey(org, project),
      [],
    );
  }

  public toggleDashboard(name: string) {
    this.favourites.update((f) => {
      const existingIdx = f.indexOf(name);
      if (existingIdx === -1) {
        return [...f, name];
      } else {
        return [...f.slice(0, existingIdx), ...f.slice(existingIdx + 1)];
      }
    });
  }

  public moveDashboardUp(name: string) {
    this.favourites.update((f) => {
      const existingIdx = f.indexOf(name);
      if (existingIdx <= 0) return f;
      return [
        ...f.slice(0, existingIdx - 1),
        name,
        f[existingIdx - 1],
        ...f.slice(existingIdx + 1),
      ];
    });
  }

  public moveDashboardDown(name: string) {
    this.favourites.update((f) => {
      const existingIdx = f.indexOf(name);
      if (existingIdx === -1 || existingIdx === f.length - 1) return f;
      return [
        ...f.slice(0, existingIdx),
        f[existingIdx + 1],
        name,
        ...f.slice(existingIdx + 1),
      ];
    });
  }
}
