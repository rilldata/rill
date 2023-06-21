import { createStatusLineGutter } from "./gutter";
import { createLineStatusHighlighter } from "./highlighter";
import { lineStatusesStateField, updateLineStatuses } from "./state";

/** creates a special gutter that enables usage of line statuses. */
export function createLineStatusSystem() {
  let debounceTimer: ReturnType<typeof setTimeout>;
  return {
    /** closes the line status state over a function that dispatches a transaction.
     */
    createUpdater(lineStatuses) {
      return (view) => {
        const transaction = updateLineStatuses.of({
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
    },
    extension: [
      lineStatusesStateField,
      createStatusLineGutter(),
      createLineStatusHighlighter(),
    ],
  };
}
