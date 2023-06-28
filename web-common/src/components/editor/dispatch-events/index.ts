import { EditorView } from "@codemirror/basic-setup";
import type { ViewUpdate } from "@codemirror/view";
export interface UpdateDetails {
  content: string;
  viewUpdate: ViewUpdate;
}

export const DEFAULT_EDITOR_UPDATE_DEBOUNCE_MS = 300;

export function bindEditorEventsToDispatcher(
  dispatch: (event: string, data?: unknown) => void
) {
  return EditorView.updateListener.of((viewUpdate: ViewUpdate) => {
    const state = viewUpdate.state;

    if (viewUpdate.focusChanged && viewUpdate.view.hasFocus) {
      dispatch("receive-focus");
    }
    if (viewUpdate.docChanged) {
      /** we will pass in the content directly as well as the viewUpdate more broadly.
       * The viewUpdate can be used to look at transactions at the parent component level.
       */
      dispatch("update", {
        content: state.doc.toString(),
        viewUpdate,
      } as UpdateDetails);
    }
  });
}
