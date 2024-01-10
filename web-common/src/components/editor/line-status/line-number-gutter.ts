import { GutterMarker, gutter } from "@codemirror/view";
import LineNumberGutterMarkerComponent from "./LineNumberGutterMarker.svelte";
import { lineStatusesStateField, updateLineStatuses } from "./state";
import type { SvelteComponent } from "svelte";

export const LINE_NUMBER_GUTTER_CLASS = "cm-line-number-gutter";

class NumberMarker extends GutterMarker {
  element: HTMLElement;
  component: SvelteComponent;
  line: number;
  level: "error" | "warning" | "info" | "success" | undefined;
  active: boolean;
  constructor(
    line: number,
    level: "error" | "warning" | "info" | "success" | undefined,
    active: boolean,
  ) {
    super();

    this.line = line;
    this.level = level;
    this.active = active;
    this.element = document.createElement("div");
    this.component = new LineNumberGutterMarkerComponent({
      target: this.element,
      props: { line, level, active },
    });
  }
  eq(mkr) {
    return (
      mkr.line === this.line &&
      mkr.level === this.level &&
      mkr.active === this.active
    );
  }
  toDOM() {
    return this.element;
  }
  destroy() {
    this.component.$destroy();
  }
}

export const createLineNumberGutter = () =>
  gutter({
    class: LINE_NUMBER_GUTTER_CLASS,
    initialSpacer: (view) =>
      new NumberMarker(view.state.doc.lines, undefined, false),
    updateSpacer: (spacer: NumberMarker, update) => {
      spacer.component.$set({
        line: update.state.doc.lines,
        level: undefined,
        active: false,
      });
      return spacer;
    },
    lineMarker(view, line) {
      const visibleRanges = view.visibleRanges;
      const lineStart = line.from;
      const lineEnd = line.to;
      // render only line numbers in the viewport.
      // FIXME: get the semantics right when there is an empty string.
      if (
        !visibleRanges.some(
          (range) => range.from <= lineStart && range.to >= lineEnd,
        ) &&
        !(view.state.doc.lines === 1)
      ) {
        return;
      }

      const activeLine = view.state.doc.lineAt(
        view.state.selection.main.head,
      ).number;
      // Retrieve the line status for this line
      const lineStatuses = view.state.field(lineStatusesStateField);
      const lineNumber = view.state.doc.lineAt(line.from).number;
      const thisStatus = lineStatuses.find((ls) => ls.line === lineNumber);

      // Create a new NumberMarker with the line number and the background color for this line's status
      return new NumberMarker(
        lineNumber,
        thisStatus?.level,
        activeLine === lineNumber,
      );
    },
    lineMarkerChange(update) {
      return update.transactions.some((tr) => {
        const effectPresent = tr.effects.some((effect) =>
          effect.is(updateLineStatuses),
        );
        return effectPresent || update;
      });
    },
  });
