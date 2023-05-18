import { RangeSetBuilder } from "@codemirror/rangeset";
import {
  Decoration,
  EditorView,
  ViewPlugin,
  ViewUpdate,
} from "@codemirror/view";
import { lineStatusesStateField, updateLineStatuses } from "../line-status";

const lineBackground = (level) =>
  Decoration.line({
    class: `cm-line-${level}`,
    // attributes: { style: "background-color: #FEF2F2" },
  });

function errorLinesDecoration(view) {
  const lineStatuses = view.state.field(lineStatusesStateField);

  const builder = new RangeSetBuilder<Decoration>();

  for (const { line, level } of lineStatuses) {
    if (line !== null && line > 0 && line <= view.state.doc.lines) {
      const from = view.state.doc.line(line).from;
      builder.add(from, from, lineBackground(level));
    }
  }
  return builder.finish();
}

export function createLineStatusHighlighter() {
  return ViewPlugin.fromClass(
    class {
      decorations;

      constructor(view: EditorView) {
        this.decorations = errorLinesDecoration(view);
      }

      update(update: ViewUpdate) {
        if (
          update.transactions.some((tr) => {
            return tr.effects.some((effect) => effect.is(updateLineStatuses));
          })
        ) {
          this.decorations = errorLinesDecoration(update.view);
        }
      }
    },
    { decorations: (v) => v.decorations }
  );
}
