import type { ViewUpdate } from "@codemirror/view";
import { EditorView } from "@codemirror/view";
export interface UpdateDetails {
  content: string;
  viewUpdate: ViewUpdate;
}

/** Provides a way to bubble up different CodeMirror events (primarily docChanged)
 * to the parent component via a Svelte dispatcher.
 */
export function bindEditorEventsToDispatcher(
  dispatch: (event: string, data?: unknown) => void,
  whenFocused = false,
) {
  return EditorView.updateListener.of((viewUpdate: ViewUpdate) => {
    if (viewUpdate.focusChanged && viewUpdate.view.hasFocus) {
      dispatch("receive-focus");
    }
    if (viewUpdate.docChanged) {
      /** we will pass in the content directly as well as the viewUpdate more broadly.
       * The viewUpdate can be used to look at transactions at the parent component level.
       */
      if (whenFocused && !viewUpdate.view.hasFocus) return;
      dispatch("update", {
        content: viewUpdate.view.state.doc.toString(),
        viewUpdate,
      } as UpdateDetails);
    }
  });
}
