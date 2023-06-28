import { RangeSetBuilder } from "@codemirror/rangeset";
import { gutter, GutterMarker } from "@codemirror/view";
import type { SvelteComponent } from "svelte";
import LineNumberGutterMarkerComponent from "./LineNumberGutterMarker.svelte";
import { lineStatusesStateField, updateLineStatuses } from "./state";
import StatusGutterMarkerComponent from "./StatusGutterMarker.svelte";

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
      props: { line, level, message, active },
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

class NumberMarker extends GutterMarker {
  element: HTMLElement;
  component: SvelteComponent;
  line: number;
  level: "error" | "warning" | "info" | "success" | undefined;
  active: boolean;
  constructor(
    line: number,
    level: "error" | "warning" | "info" | "success" | undefined,
    active: boolean
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
    initialSpacer: (view) =>
      new NumberMarker(view.state.doc.lines, undefined, false),
    lineMarker(view, line) {
      const visibleRanges = view.visibleRanges;
      const lineStart = line.from;
      const lineEnd = line.to;

      // render only line numbers in the viewport.
      if (
        (lineStart === 0 && lineEnd === 0) ||
        !visibleRanges.some(
          (range) => range.from <= lineStart && range.to >= lineEnd
        )
      ) {
        return;
      }

      const activeLine = view.state.doc.lineAt(
        view.state.selection.main.head
      ).number;
      // Retrieve the line status for this line
      const lineStatuses = view.state.field(lineStatusesStateField);
      const lineNumber = view.state.doc.lineAt(line.from).number;
      const thisStatus = lineStatuses.find((ls) => ls.line === lineNumber);

      // Create a new NumberMarker with the line number and the background color for this line's status
      return new NumberMarker(
        lineNumber,
        thisStatus?.level,
        activeLine === lineNumber
      );
    },
    lineMarkerChange(update) {
      return update.transactions.some((tr) => {
        const effectPresent = tr.effects.some((effect) =>
          effect.is(updateLineStatuses)
        );
        return effectPresent || update;
      });
    },
  });

export const createStatusLineGutter = () =>
  gutter({
    markers: (view) => {
      // Create a RangeSetBuilder to store the GutterMarkers
      const builder = new RangeSetBuilder<GutterMarker>();

      // check for an empty document.
      const isEmpty = !view.state.doc.toString()?.length;

      // get the line statuses.
      const lineStatuses = view.state
        .field(lineStatusesStateField)
        .filter((ls) => ls.line !== null && ls.line !== 0);

      if (!lineStatuses?.length || isEmpty) return builder.finish();

      const activeLine = view.state.doc.lineAt(
        view.state.selection.main.head
      ).number;
      // Iterate through each line status
      for (const { line, level, message } of lineStatuses) {
        // Create a GutterMarker for the line
        const from = view.state.doc.line(line).from;

        const marker = new StatusGutterMarker(
          line,
          level,
          message,
          line === activeLine
        );
        builder.add(from, from, marker);
      }
      // Add the GutterMarker to the RangeSetBuilder
      return builder.finish();
    },

    initialSpacer: (view) =>
      new StatusGutterMarker(
        view.state.doc.lines,
        "error",
        "no message needed."
      ),

    lineMarkerChange(update) {
      return update.transactions.some((tr) => {
        const effectPresent = tr.effects.some((effect) =>
          effect.is(updateLineStatuses)
        );

        return effectPresent;
      });
    },
  });
