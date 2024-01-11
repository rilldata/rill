import { RangeSetBuilder } from "@codemirror/state";
import {
  Decoration,
  EditorView,
  ViewPlugin,
  ViewUpdate,
} from "@codemirror/view";
import { lineStatusesStateField, updateLineStatuses } from "./state";

const lineBackground = (level) =>
  Decoration.line({
    class: `cm-line-${level}`,
    style: {
      "font-style": "italic",
    },
  });

function errorLinesDecoration(view) {
  const builder = new RangeSetBuilder<Decoration>();

  // return early if the doc is empty.
  if (view.state.doc.toString().length === 0) {
    return builder.finish();
  }

  const lineStatuses = view.state.field(lineStatusesStateField);

  for (const { line, level } of lineStatuses) {
    if (line !== null && line > 0 && line <= view.state.doc.lines) {
      const from = view.state.doc.line(line).from;
      builder.add(from, from, lineBackground(level));
    }
  }
  return builder.finish();
}

/** creates a line background highlighter based on the line status. */
export function createLineStatusHighlighter() {
  return ViewPlugin.fromClass(
    class {
      decorations;

      constructor(view: EditorView) {
        this.decorations = errorLinesDecoration(view);
      }

      update(update: ViewUpdate) {
        if (
          // transaction was a line status update
          update.transactions.some((tr) => {
            return tr.effects.some((effect) => effect.is(updateLineStatuses));
          })
        ) {
          this.decorations = errorLinesDecoration(update.view);
        }
      }
    },
    { decorations: (v) => v.decorations },
  );
}
