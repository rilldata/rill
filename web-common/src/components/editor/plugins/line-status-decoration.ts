import { RangeSetBuilder } from "@codemirror/rangeset";
import { StateEffect, StateField } from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  GutterMarker,
  ViewPlugin,
  ViewUpdate,
  WidgetType,
  gutter,
} from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import StatusGutterMarkerComponent from "../gutter/StatusGutterMarker.svelte";
import LineStatusHint from "../hints/LineStatusHint.svelte";
const updateLineStatus = StateEffect.define<{
  lineState: Array<{ line: number; message: string; level: string }>;
}>({
  map: (value, mapping) => {
    return {
      lineState: value.lineState
        .filter((line) => line.line !== null)
        .map((line) => ({
          line: mapping.mapPos(line.line),
          message: line.message,
          level: line.level,
        })),
    };
  },
});

export const lineStatusStateField = StateField.define({
  create: () => [],
  update: (lines, tr) => {
    // Handle transactions with the updateLineState effect
    for (const effect of tr.effects) {
      if (effect.is(updateLineStatus)) {
        // Clear the existing errors and set the new errors
        return effect.value.lineState.slice();
      }
    }

    return lines;
  },
  compare: (a, b) => a === b,
});

const levels = {
  error: {
    bgColor: "rgba(255,0,0,.1)",
  },
};

function bgDeco(view) {
  const lineStatuses = view.state.field(lineStatusStateField);

  const builder = new RangeSetBuilder<Decoration>();

  for (const { line, message, level } of lineStatuses) {
    if (line !== null) {
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

class HintTextWidget extends WidgetType {
  text: string;
  element: HTMLElement;
  component: SvelteComponent;
  constructor(text) {
    super();
    this.element = document.createElement("span");
    this.component = new LineStatusHint({
      target: this.element,
      props: { text: text.split("at line")[0].trim() },
    });
  }
  eq(other: HintTextWidget) {
    return other.text == this.text;
  }

  toDOM() {
    return this.element;
  }

  ignoreEvent() {
    return false;
  }
}

function textWidget(view) {
  const lineStatuses = view.state.field(lineStatusStateField);
  // loop through these.
  const widgets = [];
  for (const lineStatus of lineStatuses) {
    if (lineStatus.line !== null) {
      const startPos = view.state.doc.line(lineStatus.line).from;
      const { to } = view.state.doc.lineAt(startPos);
      const widget = Decoration.widget({
        widget: new HintTextWidget(lineStatus.message),
        side: 1,
      });
      widgets.push(widget.range(to));
    }
  }
  return Decoration.set(widgets);
}

export function createLineStatusHints() {
  return ViewPlugin.fromClass(
    class {
      hints: DecorationSet;

      constructor(view) {
        this.hints = textWidget(view);
      }

      update(update: ViewUpdate) {
        this.hints = textWidget(update.view);
      }
    },
    {
      decorations: (v) => v.hints,
    }
  );
}

export function createLineStatusDecoration() {
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

/** create a DOM node that contains the gutter container, and map a Svelte component
 * to it.
 */
class StatusGutterMarker extends GutterMarker {
  element: HTMLElement;
  component: SvelteComponent;

  constructor(line, level, message, active = false) {
    super();

    this.element = document.createElement("div");
    this.component = new StatusGutterMarkerComponent({
      target: this.element,
      props: { line, level, message, active },
    });
  }
  eq() {
    return false;
  }
  toDOM() {
    return this.element;
  }
  destroy() {
    this.component.$destroy();
  }
}

export function createLineStatusFactory() {
  return {
    update(state) {
      return (view) => {
        const transaction = updateLineStatus.of({ lineState: state });
        view.dispatch({
          effects: [transaction],
        });
      };
    },
    field: lineStatusStateField,
    extension: [
      gutter({
        lineMarker(view, line) {
          const lineStates = view.state
            .field(lineStatusStateField)
            .filter((line) => {
              return line.line !== null;
            })
            .map((line) => {
              return {
                ...line,
                from: view.state.doc.line(line.line).from,
                to: view.state.doc.line(line.line).to,
              };
            });
          const matchFromAndTo = lineStates.find((lineState) => {
            return lineState.from === line.from && lineState.to === line.to;
          });

          const currentLine = view.state.doc.lineAt(
            view.state.selection.main.head
          ).number;

          const thisLine = view.state.doc.lineAt(line.from).number;

          return new StatusGutterMarker(
            thisLine, // line number
            matchFromAndTo?.level,
            matchFromAndTo?.message,
            currentLine === thisLine
          );
        },
        initialSpacer: () =>
          new StatusGutterMarker(90, "error", "no message needed."),

        lineMarkerChange(update) {
          return update.transactions.some((tr) => {
            return tr.effects.some((effect) => effect.is(updateLineStatus));
          });
        },
      }),
      // not ready yet
      //createLineStatusHints(),
      createLineStatusDecoration(),
    ],
  };
}
