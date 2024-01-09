import { RangeSetBuilder } from "@codemirror/state";
import { gutter, GutterMarker } from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import { lineStatusesStateField, updateLineStatuses } from "./state";
import StatusGutterMarkerComponent from "./StatusGutterMarker.svelte";

export const LINE_STATUS_GUTTER_CLASS = "cm-line-status-gutter";

class StatusGutterMarker extends GutterMarker {
  element: HTMLElement;
  component: SvelteComponent;
  line: number;
  active: boolean;

  constructor(line, level, message, active = false) {
    super();

    this.line = line;
    this.element = document.createElement("div");
    this.active = active;
    this.component = new StatusGutterMarkerComponent({
      target: this.element,
      props: { level, message, active },
    });
  }
  eq(mkr) {
    return mkr.line === this.line && mkr.active === this.active;
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
    class: LINE_STATUS_GUTTER_CLASS,
    markers: (view) => {
      const totalLines = view.state.doc.lines;
      // Create a RangeSetBuilder to store the GutterMarkers
      const builder = new RangeSetBuilder<GutterMarker>();

      // check for an empty document.
      const isEmpty = !view.state.doc.toString()?.length;

      // get the line statuses.
      const lineStatuses = view.state
        .field(lineStatusesStateField)
        // remove any line statuses that are greater than the total lines
        .filter((ls) => ls?.line && ls.line !== null && ls.line <= totalLines);

      if (!lineStatuses?.length || isEmpty) return builder.finish();

      const activeLine = view.state.doc.lineAt(
        view.state.selection.main.head,
      ).number;
      // Iterate through each remaining line status
      for (const { line, level, message } of lineStatuses) {
        const from = view.state.doc.line(line).from;
        // Create a GutterMarker for the line
        const marker = new StatusGutterMarker(
          line,
          level,
          message,
          line === activeLine,
        );
        builder.add(from, from, marker);
      }
      // Add the GutterMarker to the RangeSetBuilder
      return builder.finish();
    },

    // note: unlike the line number gutter, this spacer does not need to be
    // updated when the document changes
    initialSpacer: (view) =>
      new StatusGutterMarker(
        view.state.doc.lines,
        "error",
        "no message needed.",
      ),

    lineMarkerChange(update) {
      return update.transactions.some((tr) => {
        const effectPresent = tr.effects.some((effect) =>
          effect.is(updateLineStatuses),
        );

        return effectPresent;
      });
    },
  });
