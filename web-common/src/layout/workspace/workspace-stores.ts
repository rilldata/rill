import { writable } from "svelte/store";
import {
  DEFAULT_INSPECTOR_WIDTH,
  DEFAULT_PREVIEW_TABLE_HEIGHT,
} from "../config";
import { page } from "$app/stores";
import { derived } from "svelte/store";
import { debounce } from "@rilldata/web-common/lib/create-debouncer";

type WorkspaceLayout = {
  inspector: {
    width: number;
    visible: boolean;
  };
  table: {
    height: number;
    visible: boolean;
  };
};

class WorkspaceLayoutStore {
  private inspectorVisible = writable<boolean>(true);
  private inspectorWidth = writable<number>(DEFAULT_INSPECTOR_WIDTH);
  private tableHeight = writable<number>(DEFAULT_PREVIEW_TABLE_HEIGHT);
  private tableVisible = writable<boolean>(true);

  constructor(key: string) {
    const history = localStorage.getItem(key);

    if (history) {
      const parsed = JSON.parse(history) as WorkspaceLayout;
      this.inspectorVisible.set(parsed.inspector.visible);
      this.inspectorWidth.set(parsed.inspector.width);
      this.tableHeight.set(parsed.table.height);
      this.tableVisible.set(parsed.table.visible);
    }

    const debouncer = debounce(
      (v: WorkspaceLayout) => localStorage.setItem(key, JSON.stringify(v)),
      750,
    );

    this.subscribe((v) => {
      debouncer(v);
    });
  }

  subscribe = derived(
    [
      this.inspectorVisible,
      this.inspectorWidth,
      this.tableHeight,
      this.tableVisible,
    ],
    ([$inspectorVisible, $inspectorWidth, $tableHeight, $tableVisible]) => {
      const layout: WorkspaceLayout = {
        inspector: {
          visible: $inspectorVisible,
          width: $inspectorWidth,
        },
        table: {
          height: $tableHeight,
          visible: $tableVisible,
        },
      };
      return layout;
    },
  ).subscribe;

  get inspector() {
    return {
      visible: this.inspectorVisible,
      width: this.inspectorWidth,
      open: () => this.inspectorVisible.set(true),
      close: () => this.inspectorVisible.set(false),
      toggle: () => this.inspectorVisible.update((v) => !v),
    };
  }

  get table() {
    return {
      height: this.tableHeight,
      visible: this.tableVisible,
      open: () => this.tableVisible.set(true),
      close: () => this.tableVisible.set(false),
      toggle: () => this.tableVisible.update((v) => !v),
    };
  }
}

class Workspaces {
  private workspaces = writable(new Map<string, WorkspaceLayoutStore>());

  subscribe = derived([page, this.workspaces], ([$page, $inspectors]) => {
    const context = $page.route.id ?? crypto.randomUUID();
    const assetId = $page.params.name;

    const key = `${context}:${assetId}`;

    let store = $inspectors.get(key);
    if (!store) {
      store = new WorkspaceLayoutStore(key);
      $inspectors.set(key, store);
    }
    return store;
  }).subscribe;

  set = this.workspaces.set;
  update = this.workspaces.update;
}

export const workspaces = new Workspaces();
