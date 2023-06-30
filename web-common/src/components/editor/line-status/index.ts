import type { EditorView } from "@codemirror/basic-setup";
import { createLineNumberGutter, createStatusLineGutter } from "./gutter";
import { createLineStatusHighlighter } from "./highlighter";
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
