import { createStatusLineGutter } from "../gutter";
import { createLineStatusHighlighter } from "../highlighter";
import { lineStatusStateField, updateLineStatus } from "../line-status";

// class HintTextWidget extends WidgetType {
//   text: string;
//   element: HTMLElement;
//   component: SvelteComponent;
//   constructor(text) {
//     super();
//     this.element = document.createElement("span");
//     this.component = new LineStatusHint({
//       target: this.element,
//       props: { text: text.split("at line")[0].trim() },
//     });
//   }
//   eq(other: HintTextWidget) {
//     return other.text == this.text;
//   }

//   toDOM() {
//     return this.element;
//   }

//   ignoreEvent() {
//     return false;
//   }
// }

// function textWidget(view) {
//   const lineStatuses = view.state.field(lineStatusStateField);
//   // loop through these.
//   const widgets = [];
//   for (const lineStatus of lineStatuses) {
//     if (lineStatus.line !== null) {
//       const startPos = view.state.doc.line(lineStatus.line).from;
//       const { to } = view.state.doc.lineAt(startPos);
//       const widget = Decoration.widget({
//         widget: new HintTextWidget(lineStatus.message),
//         side: 1,
//       });
//       widgets.push(widget.range(to));
//     }
//   }
//   return Decoration.set(widgets);
// }

// export function createLineStatusHints() {
//   return ViewPlugin.fromClass(
//     class {
//       hints: DecorationSet;

//       constructor(view) {
//         this.hints = textWidget(view);
//       }

//       update(update: ViewUpdate) {
//         this.hints = textWidget(update.view);
//       }
//     },
//     {
//       decorations: (v) => v.hints,
//     }
//   );
// }

/** creates a special gutter that enables usage of line statuses. */
export function createLineStatusSystem() {
  return {
    /** closes the line status state over a function that dispatches a transaction.
     */
    createUpdater(state) {
      return (view) => {
        const transaction = updateLineStatus.of({ lineState: state });
        view.dispatch({
          effects: [transaction],
        });
      };
    },
    extension: [
      lineStatusStateField,
      createStatusLineGutter(),
      createLineStatusHighlighter(),
    ],
  };
}
