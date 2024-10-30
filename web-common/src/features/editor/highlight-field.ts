import { StateField, StateEffect } from "@codemirror/state";
import { type DecorationSet, Decoration } from "@codemirror/view";
import { EditorView } from "@codemirror/view";
import type { Reference } from "../models/utils/get-table-references";

const highlightField = StateField.define<DecorationSet>({
  create() {
    return Decoration.none;
  },
  update(underlines, tr) {
    underlines = underlines.map(tr.changes);
    underlines = underlines.update({
      filter: () => false,
    });

    for (const e of tr.effects)
      if (e.is(addHighlight)) {
        underlines = underlines.update({
          add: [highlightMark.range(e.value.from, e.value.to)],
        });
      }
    return underlines;
  },
  provide: (f) => EditorView.decorations.from(f),
});

const addHighlight = StateEffect.define<{ from: number; to: number }>({
  map: ({ from, to }, change) => ({
    from: change.mapPos(from),
    to: change.mapPos(to),
  }),
});
const highlightMark = Decoration.mark({ class: "cm-underline" });

export function underlineSelection(editor: EditorView, refs: Reference[]) {
  const selections = refs.map((ref) => {
    return {
      from: ref.referenceIndex,
      to: ref.referenceIndex + ref.reference.length,
    };
  });

  const effects: StateEffect<unknown>[] = selections.map(({ from, to }) =>
    addHighlight.of({ from, to }),
  );

  if (!editor?.state.field(highlightField, false))
    effects.push(StateEffect.appendConfig.of([highlightField]));
  editor?.dispatch({ effects });
}
