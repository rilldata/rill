import { RangeSetBuilder } from "@codemirror/rangeset";
import {
  Decoration,
  DecorationSet,
  ViewPlugin,
  ViewUpdate,
} from "@codemirror/view";
import { levels, lineStatusStateField } from "../line-status";
function bgDeco(view) {
  const lineStatuses = view.state.field(lineStatusStateField);

  const builder = new RangeSetBuilder<Decoration>();

  for (const { line, level } of lineStatuses) {
    if (line !== null && line !== 0 && view.state.doc.length) {
      const startPos = view.state.doc.line(line).from;
      const { to, from } = view.state.doc.lineAt(startPos);

      builder.add(
        from,
        to,
        // FIXME: this should be Decoration.line, but it appears to clobber
        // the line text if I use it. Something must be wrong with the updates.
        Decoration.mark({
          attributes: {
            style: `background-color: ${
              levels?.[level]?.bgColor || levels.error.bgColor
            }`,
          },
        })
      );
    }
  }
  return builder.finish();
}

export function createLineStatusHighlighter() {
  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;
      hints: DecorationSet;

      constructor(view) {
        this.decorations = bgDeco(view);
      }

      update(update: ViewUpdate) {
        this.decorations = bgDeco(update.view);
      }
    },
    {
      decorations: (v) => v.decorations,
    }
  );
}
