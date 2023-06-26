import { createStatusLineGutter } from "./gutter";
import { createLineStatusHighlighter } from "./highlighter";
import {
  LineStatus,
  lineStatusesStateField,
  updateLineStatuses as updateLineStatusesEffect,
} from "./state";

export function setLineStatuses(lineStatuses: LineStatus[]) {
  let debounceTimer: ReturnType<typeof setTimeout>;
  return (view) => {
    const transaction = updateLineStatusesEffect.of({
      lineStatuses: lineStatuses,
    });
    // debounce this transaction to avoid unnecessary flickering updates.
    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      view.dispatch({
        effects: [transaction],
      });
    }, 100);
  };
}

/** creates a special gutter that enables usage of line statuses. */
export function createLineStatusSystem() {
  return {
    extension: [
      lineStatusesStateField,
      createStatusLineGutter(),
      createLineStatusHighlighter(),
    ],
  };
}
