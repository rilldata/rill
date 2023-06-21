import { EditorView } from "@codemirror/basic-setup";

export function bindEditorEventsToDispatcher(
  dispatch: (event: string, data?: unknown) => void,
  stateFieldUpdaters: ((view: EditorView) => void)[]
) {
  return EditorView.updateListener.of((viewUpdate) => {
    const view = viewUpdate.view;
    const state = viewUpdate.state;
    const cursor = state.selection.main.head;
    const line = state.doc.lineAt(cursor);
    // dispatch current cursor location
    dispatch("cursor", {
      line: line.number,
      column: cursor - line.from,
    });
    if (viewUpdate.focusChanged && viewUpdate.view.hasFocus) {
      dispatch("receive-focus");
    }
    if (viewUpdate.docChanged) {
      dispatch("update", { content: state.doc.toString() });
      stateFieldUpdaters.forEach((updater) => {
        updater(view);
      });
    }
  });
}
