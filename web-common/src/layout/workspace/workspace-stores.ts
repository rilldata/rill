import { debounce } from "@rilldata/web-common/lib/create-debouncer";
import { derived, writable } from "svelte/store";
import {
  DEFAULT_INSPECTOR_WIDTH,
  DEFAULT_PREVIEW_TABLE_HEIGHT,
} from "../config";

// Views available in every workspace. Individual workspaces can support additional
// views by passing an extended union to `workspaces.get` (e.g. the metrics view
// workspace adds an "explore" view).
export type WorkspaceView = "code" | "split" | "viz";

type WorkspaceLayout<View extends string> = {
  inspector: {
    width: number;
    visible: boolean;
  };
  table: {
    height: number;
    visible: boolean;
  };
  view: View;
};

class WorkspaceLayoutStore<View extends string = WorkspaceView> {
  private inspectorVisible = writable<boolean>(true);
  private inspectorWidth = writable<number>(DEFAULT_INSPECTOR_WIDTH);
  private tableVisible = writable<boolean>(true);
  private tableHeight = writable<number>(DEFAULT_PREVIEW_TABLE_HEIGHT);
  public view = writable<View>("viz" as View);

  constructor(key: string) {
    const history = localStorage.getItem(key);

    if (history) {
      const parsed = JSON.parse(history) as WorkspaceLayout<View>;
      this.inspectorVisible.set(parsed?.inspector?.visible ?? true);
      this.inspectorWidth.set(
        parsed?.inspector?.width ?? DEFAULT_INSPECTOR_WIDTH,
      );
      this.tableHeight.set(
        parsed?.table?.height ?? DEFAULT_PREVIEW_TABLE_HEIGHT,
      );
      this.tableVisible.set(parsed?.table?.visible ?? true);
      if (parsed?.view) this.view.set(parsed.view);
    }

    const debouncer = debounce(
      (v: WorkspaceLayout<View>) =>
        localStorage.setItem(key, JSON.stringify(v)),
      500,
    );

    this.subscribe((v) => debouncer(v));
  }

  subscribe = derived(
    [
      this.inspectorVisible,
      this.inspectorWidth,
      this.tableHeight,
      this.tableVisible,
      this.view,
    ],
    ([
      $inspectorVisible,
      $inspectorWidth,
      $tableHeight,
      $tableVisible,
      $view,
    ]) => {
      const layout: WorkspaceLayout<View> = {
        inspector: {
          visible: $inspectorVisible,
          width: $inspectorWidth,
        },
        table: {
          height: $tableHeight,
          visible: $tableVisible,
        },
        view: $view,
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
  private workspaces = new Map<string, WorkspaceLayoutStore<string>>();

  get = <View extends string = WorkspaceView>(
    context: string,
  ): WorkspaceLayoutStore<View> => {
    let store = this.workspaces.get(context);

    if (!store) {
      store = new WorkspaceLayoutStore(context);
      this.workspaces.set(context, store);
    }

    return store as WorkspaceLayoutStore<View>;
  };
}

export const workspaces = new Workspaces();

// consumeViewSearchParam handles the `view` search param on file routes: links can
// append `?view=<view>` to open a file's workspace on a specific view (e.g.
// `?view=explore` for the explore editor of a metrics view file). It stores the view
// for the file and returns the URL to redirect to with the param removed, or null if
// the param is not present. Called from the files `+page.ts` load functions.
export function consumeViewSearchParam(
  url: URL,
  filePath: string,
): string | null {
  const view = url.searchParams.get("view");
  if (!view) return null;

  workspaces.get<string>(filePath).view.set(view);

  const clean = new URL(url);
  clean.searchParams.delete("view");
  return clean.pathname + clean.search;
}
