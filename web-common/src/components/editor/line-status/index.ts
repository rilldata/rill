import type { EditorView } from "@codemirror/basic-setup";
import { createLineStatusHighlighter } from "./highlighter";
import { createLineNumberGutter } from "./line-number-gutter";
import { createStatusLineGutter } from "./line-status-gutter";
import {
  LineStatus,
  lineStatusesStateField,
  updateLineStatuses as updateLineStatusesEffect,
} from "./state";

export function setLineStatuses(lineStatuses: LineStatus[], view: EditorView) {
  const transaction = updateLineStatusesEffect.of({
    lineStatuses: lineStatuses,
  });

  view.dispatch({
    effects: [transaction],
  });
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
