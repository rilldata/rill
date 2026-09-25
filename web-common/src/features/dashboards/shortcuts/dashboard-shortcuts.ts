export enum DashboardShortcutAction {
  ToggleHelp = "toggle-help",
  OpenFilter = "open-filter",
  OpenMetricPicker = "open-metric-picker",
  OpenDimensionPicker = "open-dimension-picker",
  InspectCell = "inspect-cell",
  LockInspector = "lock-inspector",
  ClearSelection = "clear-selection",
  SelectAll = "select-all",
  PanLeft = "pan-left",
  PanRight = "pan-right",
  Zoom = "zoom",
  UndoZoom = "undo-zoom",
  Explain = "explain",
}

export interface DashboardShortcut {
  action: DashboardShortcutAction;
  keys: string[];
  description: string;
}

export const DASHBOARD_SHORTCUTS: DashboardShortcut[] = [
  {
    action: DashboardShortcutAction.ToggleHelp,
    keys: ["?"],
    description: "Show/hide the keyboard-shortcuts menu",
  },
  {
    action: DashboardShortcutAction.OpenFilter,
    keys: ["/"],
    description: "Open the Filter menu",
  },
  {
    action: DashboardShortcutAction.OpenMetricPicker,
    keys: [","],
    description: "Open the metric picker",
  },
  {
    action: DashboardShortcutAction.OpenDimensionPicker,
    keys: ["."],
    description: "Open the dimension picker",
  },
  {
    action: DashboardShortcutAction.InspectCell,
    keys: ["Space"],
    description: "Show/hide the cell inspector",
  },
  {
    action: DashboardShortcutAction.LockInspector,
    keys: ["L"],
    description: "Lock/unlock the cell inspector",
  },
  {
    action: DashboardShortcutAction.ClearSelection,
    keys: ["Esc"],
    description: "Close the inspector or clear a chart selection",
  },
  {
    action: DashboardShortcutAction.SelectAll,
    keys: ["⌘/Ctrl", "A"],
    description: "Select all dimension values",
  },
  {
    action: DashboardShortcutAction.PanLeft,
    keys: ["←"],
    description: "Pan the chart backward",
  },
  {
    action: DashboardShortcutAction.PanRight,
    keys: ["→"],
    description: "Pan the chart forward",
  },
  {
    action: DashboardShortcutAction.Zoom,
    keys: ["Z"],
    description: "Zoom into the selected range",
  },
  {
    action: DashboardShortcutAction.UndoZoom,
    keys: ["⌘/Ctrl", "Z"],
    description: "Undo chart zoom",
  },
  {
    action: DashboardShortcutAction.Explain,
    keys: ["E"],
    description: "Explain the selected range",
  },
];

const KEY_ACTIONS = new Map([
  ["?", DashboardShortcutAction.ToggleHelp],
  ["/", DashboardShortcutAction.OpenFilter],
  [",", DashboardShortcutAction.OpenMetricPicker],
  [".", DashboardShortcutAction.OpenDimensionPicker],
]);

function isEditableTarget(target: EventTarget | null): boolean {
  return (
    target instanceof Element &&
    (!!target.closest("input, textarea, select, [contenteditable]") ||
      (target instanceof HTMLElement && target.isContentEditable))
  );
}

export function getDashboardShortcutAction(
  event: KeyboardEvent,
): DashboardShortcutAction | undefined {
  if (
    event.repeat ||
    event.metaKey ||
    event.ctrlKey ||
    event.altKey ||
    isEditableTarget(event.target)
  ) {
    return undefined;
  }

  return KEY_ACTIONS.get(event.key);
}

export function performDashboardShortcut(
  action: DashboardShortcutAction,
  root: ParentNode = document,
): boolean {
  const target = root.querySelector<HTMLElement>(
    `[data-dashboard-shortcut="${action}"]`,
  );
  if (!target) return false;

  target.click();
  return true;
}
