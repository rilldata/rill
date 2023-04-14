import { RangeSetBuilder } from "@codemirror/rangeset";
import { StateEffect, StateField } from "@codemirror/state";
import {
  Decoration,
  DecorationSet,
  GutterMarker,
  ViewPlugin,
  ViewUpdate,
  gutter,
} from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import StatusGutterMarkerComponent from "./StatusGutterMarker.svelte";

const updateLineState = StateEffect.define<{
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

export const lineState = StateField.define({
  create: () => [],
  update: (lines, tr) => {
    // Handle transactions with the updateLineState effect
    for (const effect of tr.effects) {
      if (effect.is(updateLineState)) {
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
  const lineStates = view.state.field(lineState);

  const builder = new RangeSetBuilder<Decoration>();

  for (const { line, message, level } of lineStates) {
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

export function createLineStatusDecoration() {
  return ViewPlugin.fromClass(
    class {
      decorations: DecorationSet;
      // gutterDecorations: DecorationSet;

      constructor(view) {
        this.decorations = bgDeco(view);
        // this.gutterDecorations = gutterDeco(view);
      }

      update(update: ViewUpdate) {
        this.decorations = bgDeco(update.view);
        // this.gutterDecorations = gutterDeco(update.view);
      }
    },
    {
      decorations: (v) => v.decorations,
    }
  );
}

// export function createLineStatusGutterDecoration() {
//   return ViewPlugin.fromClass(
//     class {
//       decorations: DecorationSet;

//       constructor(view) {
//         this.decorations = gutterDeco(view);
//       }

//       update(update: ViewUpdate) {
//         this.decorations = gutterDeco(update.view);
//       }
//     },
//     {
//       decorations: (v) => v.decorations,
//     }
//   );
// }

/** create a DOM node that contains the gutter container, and map a Svelte component
 * to it.
 */
class StatusGutterMarker extends GutterMarker {
  element: HTMLElement;
  component: SvelteComponent;
  constructor(level) {
    super();

    this.element = document.createElement("div");
    this.component = new StatusGutterMarkerComponent({
      target: this.element,
      props: { level },
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

class NumberMarker extends GutterMarker {
  number: number;
  level: string;
  constructor(number, level) {
    super();
    this.number = number;
    this.level = level;
  }
  toDOM() {
    const el = document.createElement("div");
    el.textContent = String(this.number);
    el.style.textAlign = "right";
    el.style.paddingRight = "8px";
    el.style.paddingLeft = "8px";
    if (this.level === "error") {
      el.style.backgroundColor = "hsla(1, 100%, 80%, .5)";
    }
    return el;
  }
}

export function createLineStatusFactory() {
  return {
    update(state) {
      return (view) => {
        const transaction = updateLineState.of({ lineState: state });
        view.dispatch({
          effects: [transaction],
        });
      };
    },
    field: lineState,
    extension: [
      gutter({
        lineMarker(view, line) {
          const lineStates = view.state
            .field(lineState)
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
          const matchFromAndTo = lineStates.some((lineState) => {
            return lineState.from === line.from && lineState.to === line.to;
          });

          return matchFromAndTo ? new StatusGutterMarker("error") : null;
        },
        initialSpacer: () => new StatusGutterMarker(null),
        lineMarkerChange(update) {
          return update.transactions.some((tr) => {
            return tr.effects.some((effect) => effect.is(updateLineState));
          });
        },
      }),
      gutter({
        lineMarker(view, line) {
          //
          const number = view.state.doc.lineAt(line.from).number;

          // get line status
          const lineStates = view.state.field(lineState);
          const hasStatus = lineStates
            .filter((line) => line.line !== null)
            .map((line) => {
              return {
                ...line,
                from: view.state.doc.line(line.line).from,
                to: view.state.doc.line(line.line).to,
              };
            })
            .some((lineState) => {
              return lineState.from === line.from && lineState.to === line.to;
            });
          return new NumberMarker(number, hasStatus ? "error" : undefined);
        },
        initialSpacer: (view) => {
          // get largest line number
          const number = view.state.doc.lineAt(view.state.doc.length).number;
          return new NumberMarker(number, null);
        },
      }),

      createLineStatusDecoration(),
    ],
  };
}
