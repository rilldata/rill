import type { EditorView } from "@codemirror/view";
import { createLineStatusHighlighter } from "./highlighter";
import { createLineNumberGutter } from "./line-number-gutter";
import { createStatusLineGutter } from "./line-status-gutter";
import {
  type LineStatus,
  lineStatusesStateField,
  updateLineStatuses as updateLineStatusesEffect,
} from "./state";

/** convenience function for quickly updating line statuses elsewhere in components.
 * It assumes access to the CodeMirror editor view, which our editor components should expose as a
 * bindable prop and which tends to be accessible from edit events dispatched via
 * the dispatch-events extension in the editor component directory.
 */
export function setLineStatuses(
  lineStatuses: LineStatus[],
  view: EditorView,
  wait = true,
) {
  const transaction = updateLineStatusesEffect.of({
    lineStatuses: lineStatuses,
  });

  if (wait) {
    queueMicrotask(() => {
      view.dispatch({
        effects: [transaction],
      });
    });
  } else {
    view.dispatch({
      effects: [transaction],
    });
  }
}

/** Creates a special gutter that enables usage of line statuses.
 * Utilize this in an editor component to enable line statuses,
 * and set the line statuses via the setLineStatuses function.
 *
 * It's comprised of
 * - a state field for tracking line statuses
 * - a gutter for displaying line statuses
 * - a custom gutter for displaying line numbers that also changes color
 * depending on the line status
 * - a line bg highlighter that also changes color depending on the line status
 *
 * The lineStatusesStateField is the field that triggers updates in the other
 * extensions in this set.
 */
export function lineStatus() {
  return [
    lineStatusesStateField,
    createStatusLineGutter(),
    createLineNumberGutter(),
    createLineStatusHighlighter(),
  ];
}
