import { createLineNumberGutter, createStatusLineGutter } from "./gutter";
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

    if (debounceTimer) clearTimeout(debounceTimer);
    debounceTimer = setTimeout(() => {
      view.dispatch({
        effects: [transaction],
      });
    }, 500);
  };
}

/** creates a special gutter that enables usage of line statuses. */
export function lineStatus() {
  return [
    createStatusLineGutter(),
    //lineNumbers(),
    createLineNumberGutter(),
    // gutter({ class: "cool-gutter" }),

    lineStatusesStateField,
    createLineStatusHighlighter(),
  ];
}
