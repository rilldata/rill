import { gutter, GutterMarker } from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import { lineStatusesStateField, updateLineStatuses } from "./state";
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
      const visibleRanges = view.visibleRanges;
      const lineStart = line.from;
      const lineEnd = line.to;
      if (
        !visibleRanges.some(
          (range) => range.from <= lineStart && range.to >= lineEnd
        )
      ) {
        return null;
      }

      const hasContents = view.state.doc.toString() !== "";
      const currentLineStatuses = view.state.field(lineStatusesStateField);
      if (!currentLineStatuses) return;
      const lineStatuses = currentLineStatuses
        .filter((line) => {
          return line.line !== null && line.line !== 0;
        })
        .map((line) => {
          return {
            ...line,
            from: hasContents ? view?.state?.doc?.line(line.line)?.from : null,
            to: hasContents ? view?.state?.doc?.line(line.line)?.to : null,
          };
        });

      const matchFromAndTo = lineStatuses.find((lineStatus) => {
        return lineStatus.from === line.from && lineStatus.to === line.to;
      });

      const currentLine = view.state.doc.lineAt(
        view.state.selection.main.head
      ).number;

      const thisLine = view.state.doc.lineAt(line.from).number;

      return new StatusGutterMarker(
        thisLine,
        matchFromAndTo?.level,
        matchFromAndTo?.message,
        currentLine === thisLine
      );
    },
    initialSpacer: (view) =>
      new StatusGutterMarker(
        view.state.doc.lines,
        "error",
        "no message needed."
      ),

    lineMarkerChange(update) {
      return update.transactions.some((tr) => {
        const hasUpdate = tr.effects.some((effect) =>
          effect.is(updateLineStatuses)
        );
        console.log(hasUpdate, tr.effects);
        return hasUpdate;
      });
    },
  });
