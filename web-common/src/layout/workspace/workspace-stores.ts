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
  private tableVisible = writable<boolean>(true);
  private tableHeight = writable<number>(DEFAULT_PREVIEW_TABLE_HEIGHT);

  constructor(key: string) {
    const history = localStorage.getItem(key);

    if (history) {
      const parsed = JSON.parse(history) as WorkspaceLayout;
      this.inspectorVisible.set(parsed?.inspector?.visible ?? true);
      this.inspectorWidth.set(
        parsed?.inspector?.width ?? DEFAULT_INSPECTOR_WIDTH,
      );
      this.tableHeight.set(
        parsed?.table?.height ?? DEFAULT_PREVIEW_TABLE_HEIGHT,
      );
      this.tableVisible.set(parsed?.table?.visible ?? true);
    }

    const debouncer = debounce(
      (v: WorkspaceLayout) => localStorage.setItem(key, JSON.stringify(v)),
      750,
    );

    this.subscribe((v) => debouncer(v));
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
  private workspaces = new Map<string, WorkspaceLayoutStore>();

  subscribe = derived([page], ([$page]) => {
    const context = $page.route.id ?? crypto.randomUUID();
    const assetId = $page.params.name;

    const key = `${context}:${assetId}`;

    let store = this.workspaces.get(key);

    if (!store) {
      store = new WorkspaceLayoutStore(key);
      this.workspaces.set(key, store);
    }
    return store;
  }).subscribe;
}

export const workspaces = new Workspaces();
