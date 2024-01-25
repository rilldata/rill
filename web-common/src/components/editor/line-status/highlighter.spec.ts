import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { beforeEach, describe, expect, it } from "vitest";
import { lineStatus, setLineStatuses } from ".";

const doc = `name: test dashboard
measures:
  - label: test measure
    description: test measure
    expression: count(*)
`;

/** convenience function for creating a new CodeMirror editor  */
function newEditorView(content: string, parent: HTMLElement) {
  return new EditorView({
    state: EditorState.create({
      doc: content,
      extensions: [lineStatus()],
    }),
    parent,
  });
}

/** convenience function for selecting the gutter elements */
function getLines(container: HTMLElement) {
  return container.querySelectorAll(
    `.cm-content > .cm-line`,
  ) as NodeListOf<HTMLElement>;
}

describe("Line Status BG Highlighter Extension (CodeMirror)", () => {
  let container: HTMLElement;
  let view: EditorView;
  beforeEach(() => {
    container = document.createElement("div");
    /** create the code mirror instance before each test.
     * The doc length is irrelevant here.
     */
    view = newEditorView(doc, container);
  });

  it("renders lines without the cm-line-error class", () => {
    // filter out the initialSpacer element, which maintains visibility: hidden.

    const lines = getLines(container);
    const classes = Array.from(lines).map((line) => line.className);
    expect(classes.every((cls) => !cls.includes("cm-line-error"))).toBe(true);
  });

  it("renders a set of line statuses & correctly updates them", () => {
    setLineStatuses([{ line: 2, level: "error" }], view);
    let lines = getLines(container);
    expect(lines[1].className.includes("cm-line-error")).toBe(true);

    // set new line statuses.
    setLineStatuses(
      [
        { line: 1, level: "error" },
        { line: 3, level: "error" },
      ],
      view,
    );
    lines = getLines(container);
    Array.from(lines).forEach((line, i) => {
      expect(line.className.includes("cm-line-error")).toBe(
        i === 0 || i === 2 ? true : false,
      );
    });

    // remove all line statuses.
    setLineStatuses([], view);
    lines = getLines(container);
    expect(
      Array.from(lines).every(
        (line) => !line.className.includes("cm-line-error"),
      ),
    ).toBe(true);
  });

  it("does not render line statuses whose line number is greater than the number of lines in the doc", () => {
    setLineStatuses([{ line: 100, level: "error" }], view);
    const lines = getLines(container);
    expect(
      Array.from(lines).every(
        (line) => !line.className.includes("cm-line-error"),
      ),
    ).toBe(true);
  });
});
