import { getProtoFromDashboardState } from "@rilldata/web-common/features/dashboards/proto-state/toProto";
import type { MetricsExplorerStoreType } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  enablePatches,
  applyPatches,
  type Patch,
  produce,
  enableMapSet,
} from "immer";
import { get, type Readable, type Updater, writable } from "svelte/store";

enablePatches();
enableMapSet();

export class ImmerLayer {
  private updaters: ((patches: Patch[]) => void)[] = [];
  private name: string;

  public constructor(
    private readonly updater: (
      this: void,
      updater: Updater<MetricsExplorerStoreType>,
    ) => void,
  ) {}

  public setName(name: string) {
    this.name = name;
    this.updaters = [];
  }

  public wrapAction(
    name: string,
    callback: (metricsExplorer: MetricsExplorerEntity) => void,
  ) {
    this.updater((state) => {
      if (!state.entities[name]) {
        return state;
      }

      return produce(
        state,
        (d) => {
          callback(d.entities[name]);
          d.entities[name].proto = getProtoFromDashboardState(d.entities[name]);
        },
        (patches) => this.broadcast(patches),
      );
    });
  }

  public adhoc(
    state: MetricsExplorerStoreType,
    callback: (metricsExplorerStore: MetricsExplorerStoreType) => void,
  ) {
    return produce(
      state,
      (d) => {
        callback(d);
      },
      (patches) => this.broadcast(patches),
    );
  }

  public wrapStore(
    dashboard: Readable<MetricsExplorerEntity>,
    keys: Array<keyof MetricsExplorerEntity>,
  ): Readable<{ root: MetricsExplorerEntity }> {
    const state = get(dashboard);
    const keysLookup = new Set(keys);

    const { update, subscribe } = writable({ root: state });
    this.updaters.push((patches: Patch[]) => {
      const filteredPatches = patches.filter(
        (p) =>
          p.path.length === 1 ||
          keysLookup.has(p.path[1] as keyof MetricsExplorerEntity),
      );
      if (!filteredPatches.length) return;
      update((s) => applyPatches(s, filteredPatches));
    });

    return { subscribe };
  }

  private broadcast(patches: Patch[]) {
    patches = patches
      .filter((p) => p.path[1] == this.name)
      .map((p) => ({
        ...p,
        path: ["root", ...p.path.slice(2)],
      }));

    this.updaters.forEach((u) => u(patches));
  }
}
