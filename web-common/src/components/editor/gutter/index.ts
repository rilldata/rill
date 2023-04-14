import { gutter, GutterMarker } from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import { lineStatusStateField, updateLineStatus } from "../line-status";
import StatusGutterMarkerComponent from "./StatusGutterMarker.svelte";

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

export const createStatusLineGutter = () =>
  gutter({
    lineMarker(view, line) {
      const hasContents = view.state.doc.toString() !== "";
      const lineStates = view.state
        .field(lineStatusStateField)
        .filter((line) => {
          return line.line !== null;
        })
        .map((line) => {
          return {
            ...line,
            from: hasContents ? view?.state?.doc?.line(line.line)?.from : null,
            to: hasContents ? view?.state?.doc?.line(line.line)?.to : null,
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
  });
