import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

export enum DashboardShortcutAction {
  ToggleHelp = "toggle-help",
  OpenFilter = "open-filter",
  OpenMetricPicker = "open-metric-picker",
  OpenDimensionPicker = "open-dimension-picker",
  ToggleComparison = "toggle-comparison",
  ToggleView = "toggle-view",
  OpenDashboardList = "open-dashboard-list",
  GoHome = "go-home",
}

export type DashboardShortcutView = "canvas" | "explore" | "pivot";
export const DASHBOARD_SHORTCUT_VIEWS: readonly DashboardShortcutView[] = [
  "canvas",
  "explore",
  "pivot",
];

export interface DashboardShortcut {
  views: readonly DashboardShortcutView[];
  id: string;
  keys: string[];
  description: string;
  context?: string;
}

export function getDashboardShortcuts({
  root = document,
  explainEnabled = false,
  view = "explore",
}: {
  root?: ParentNode;
  explainEnabled?: boolean;
  view?: DashboardShortcutView;
} = {}): DashboardShortcut[] {
  const hasCharts = !!root.querySelector(
    '[data-dashboard-shortcut-context="chart"]',
  );
  const shortcuts: DashboardShortcut[] = [
    {
      id: "toggle-help",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["?"],
      description: m.dashboard_shortcut_toggle_help(),
    },
    {
      id: "open-filter",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["/"],
      description: m.dashboard_shortcut_open_filter(),
    },
    {
      id: "open-metric-picker",
      views: ["explore"],
      keys: [","],
      description: m.dashboard_shortcut_open_metric_picker(),
    },
    {
      id: "open-dimension-picker",
      views: ["explore"],
      keys: ["."],
      description: m.dashboard_shortcut_open_dimension_picker(),
    },
    {
      id: "toggle-comparison",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["C"],
      description: m.dashboard_shortcut_toggle_comparison(),
    },
    {
      id: "toggle-view",
      views: ["explore", "pivot"],
      keys: ["P"],
      description: m.dashboard_shortcut_toggle_view(),
    },
    {
      id: "open-dashboard-list",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["D"],
      description: m.dashboard_shortcut_open_dashboard_list(),
    },
    {
      id: "go-home",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["H"],
      description: m.dashboard_shortcut_go_home(),
    },
    {
      id: "inspect-cell",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: [m.dashboard_shortcuts_key_space()],
      description: m.dashboard_shortcut_inspect_cell(),
    },
    {
      id: "lock-inspector",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: ["L"],
      description: m.dashboard_shortcut_lock_inspector(),
    },
    {
      id: "clear-selection",
      views: DASHBOARD_SHORTCUT_VIEWS,
      keys: [m.dashboard_shortcuts_key_escape()],
      description: m.dashboard_shortcut_clear_selection(),
    },
    {
      id: "select-all",
      views: ["explore"],
      keys: ["⌘/Ctrl", "A"],
      description: m.dashboard_shortcut_select_all(),
      context: m.dashboard_shortcut_select_all_context(),
    },
    {
      id: "pan-left",
      views: ["explore"],
      keys: ["←"],
      description: m.dashboard_shortcut_pan_left(),
      context: m.dashboard_shortcut_pan_left_context(),
    },
    {
      id: "pan-right",
      views: ["explore"],
      keys: ["→"],
      description: m.dashboard_shortcut_pan_right(),
      context: m.dashboard_shortcut_pan_right_context(),
    },
    {
      id: "zoom",
      views: ["explore"],
      keys: ["Z"],
      description: m.dashboard_shortcut_zoom(),
    },
    {
      id: "undo-zoom",
      views: ["explore"],
      keys: ["⌘/Ctrl", "Z"],
      description: m.dashboard_shortcut_undo_zoom(),
    },
    {
      id: "explain",
      views: ["explore"],
      keys: ["E"],
      description: m.dashboard_shortcut_explain(),
    },
    {
      id: "pivot-navigate",
      views: ["pivot"],
      keys: ["↑ / ↓"],
      description: m.dashboard_shortcut_pivot_navigate(),
    },
    {
      id: "pivot-sections",
      views: ["pivot"],
      keys: ["Tab"],
      description: m.dashboard_shortcut_pivot_sections(),
    },
    {
      id: "pivot-tags",
      views: ["pivot"],
      keys: ["← / →"],
      description: m.dashboard_shortcut_pivot_tags(),
    },
    {
      id: "pivot-select",
      views: ["pivot"],
      keys: [`Enter / ${m.dashboard_shortcuts_key_space()}`],
      description: m.dashboard_shortcut_pivot_select(),
      context: m.dashboard_shortcut_pivot_select_context(),
    },
    {
      id: "pivot-escape",
      views: ["pivot"],
      keys: [m.dashboard_shortcuts_key_escape()],
      description: m.dashboard_shortcut_pivot_escape(),
    },
  ];
  return shortcuts.filter((shortcut) => {
    if (!shortcut.views.includes(view)) return false;
    switch (shortcut.id) {
      case DashboardShortcutAction.ToggleHelp:
      case DashboardShortcutAction.OpenFilter:
      case DashboardShortcutAction.OpenMetricPicker:
      case DashboardShortcutAction.OpenDimensionPicker:
      case DashboardShortcutAction.ToggleComparison:
      case DashboardShortcutAction.ToggleView:
      case DashboardShortcutAction.OpenDashboardList:
      case DashboardShortcutAction.GoHome:
        return !!root.querySelector(
          `[data-dashboard-shortcut="${shortcut.id}"]:not(:disabled):not([aria-disabled="true"])`,
        );
      case "pivot-navigate":
      case "pivot-sections":
      case "pivot-tags":
      case "pivot-select":
      case "pivot-escape":
        return !!root.querySelector(
          '[data-dashboard-shortcut-context="pivot-sidebar"]',
        );
      case "select-all":
        return !!root.querySelector(
          '[data-dashboard-shortcut-context="dimension-selection"]',
        );
      case "explain":
        return hasCharts && explainEnabled;
      case "pan-left":
      case "pan-right":
      case "zoom":
      case "undo-zoom":
        return hasCharts;
      default:
        return true;
    }
  });
}

const KEY_ACTIONS = new Map([
  ["?", DashboardShortcutAction.ToggleHelp],
  ["/", DashboardShortcutAction.OpenFilter],
  [",", DashboardShortcutAction.OpenMetricPicker],
  [".", DashboardShortcutAction.OpenDimensionPicker],
  ["c", DashboardShortcutAction.ToggleComparison],
  ["p", DashboardShortcutAction.ToggleView],
  ["d", DashboardShortcutAction.OpenDashboardList],
  ["h", DashboardShortcutAction.GoHome],
]);

function isEditableTarget(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    (!!target.closest("input, textarea, select, [contenteditable]") ||
      (target instanceof HTMLElement && target.isContentEditable))
  );
}

const OPEN_MODAL_SELECTOR =
  '[role="dialog"][aria-modal="true"]:not([data-state="closed"]):not([hidden]), [role="alertdialog"][aria-modal="true"]:not([data-state="closed"]):not([hidden])';

export function hasOpenDashboardModal(
  root: ParentNode = document,
  except?: Element,
): boolean {
  return Array.from(root.querySelectorAll(OPEN_MODAL_SELECTOR)).some(
    (modal) => modal !== except,
  );
}

export function getDashboardShortcutAction(
  event: KeyboardEvent,
  {
    helpOpen = false,
    root = document,
  }: { helpOpen?: boolean; root?: ParentNode } = {},
): DashboardShortcutAction | undefined {
  if (
    event.defaultPrevented ||
    event.repeat ||
    event.metaKey ||
    event.ctrlKey ||
    event.altKey ||
    isEditableTarget(event.target)
  ) {
    return undefined;
  }

  const action = KEY_ACTIONS.get(event.key.toLowerCase());
  if (helpOpen) {
    return action === DashboardShortcutAction.ToggleHelp ? action : undefined;
  }
  if (hasOpenDashboardModal(root)) {
    return undefined;
  }
  return action;
}

export function performDashboardShortcut(
  action: DashboardShortcutAction,
  root: ParentNode = document,
): boolean {
  const target = root.querySelector<HTMLElement>(
    `[data-dashboard-shortcut="${action}"]:not(:disabled):not([aria-disabled="true"])`,
  );
  if (!target) return false;

  target.click();
  return true;
}
