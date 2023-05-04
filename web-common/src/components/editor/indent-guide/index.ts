import { StateEffect, StateField } from "@codemirror/state";
import { Decoration, EditorView, ViewPlugin } from "@codemirror/view";

const updateIndentGuides = StateEffect.define({});

const indentGuidesField = StateField.define({
  create: () => Decoration.none,
  update(deco, tr) {
    if (tr.effects.some((effect) => effect.is(updateIndentGuides))) {
      const newDeco = [];
      const config = tr.state.facet(EditorView.theme);
      const indentUnit = tr.state.facet(EditorView.indentUnit);

      for (let pos = tr.start, end = tr.end; pos < end; ) {
        const { lineBreak, to } = tr.state.doc.lineAt(pos);
        const lineContent = tr.state.sliceDoc(pos, to);
        const indent = /^ */.exec(lineContent)[0].length;

        if (indent > 0) {
          const width = (indent / indentUnit) * config.indentGuideWidth;
          newDeco.push(Decoration.line({}).range(pos, to));
        }
        pos = lineBreak + 1;
      }

      deco = Decoration.set(newDeco);
    }
    return deco.map(tr.changes);
  },
  provide: (f) => EditorView.decorations.from(f),
});

export const indentGuides = ViewPlugin.fromClass(
  class {
    constructor(view) {
      this.update(view);
    }

    update(view) {
      const config = view.state.facet(EditorView.theme);
      view.dom.style.setProperty(
        "--indent-guide-width",
        config.indentGuideWidth + "px"
      );
      view.dom.style.setProperty(
        "--indent-guide-color",
        config.indentGuideColor
      );
    }
  },
  {
    decorations: (v) => v.state.field(indentGuidesField),
  }
);

export const indentGuidesTheme = EditorView.theme({
  indentGuideWidth: 4,
  indentGuideColor: "rgba(169, 169, 169, 0.4)",
});
