import { createStatusLineGutter } from "../gutter";
import { createLineStatusHighlighter } from "../highlighter";
import { lineStatusesStateField, updateLineStatuses } from "../line-status";

/** creates a special gutter that enables usage of line statuses. */
export function createLineStatusSystem() {
  return {
    /** closes the line status state over a function that dispatches a transaction.
     */
    createUpdater(lineStatuses) {
      return (view) => {
        const transaction = updateLineStatuses.of({
          lineStatuses: lineStatuses,
        });
        view.dispatch({
          effects: [transaction],
        });
      };
    },
    extension: [
      lineStatusesStateField,
      createStatusLineGutter(),
      createLineStatusHighlighter(),
    ],
  };
}
