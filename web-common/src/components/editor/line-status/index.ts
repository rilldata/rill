import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
import { createLineNumberGutter, createStatusLineGutter } from "./gutter";
import { createLineStatusHighlighter } from "./highlighter";
import {
  LineStatus,
  lineStatusesStateField,
  updateLineStatuses as updateLineStatusesEffect,
} from "./state";

export function setLineStatuses(lineStatuses: LineStatus[], debounce = true) {
  const debouncer = createDebouncer();
  return (view) => {
    const transaction = updateLineStatusesEffect.of({
      lineStatuses: lineStatuses,
    });

    debouncer(
      () => {
        view.dispatch({
          effects: [transaction],
        });
      },
      debounce ? 300 : 0
    );
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
