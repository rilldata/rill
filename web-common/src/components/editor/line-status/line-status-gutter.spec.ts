import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { beforeEach, describe, expect, it } from "vitest";
import { lineStatus, setLineStatuses } from ".";
import { LINE_STATUS_GUTTER_CLASS } from "./line-status-gutter";

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
function getLineStatusElements(container: HTMLElement) {
  return container.querySelectorAll(
    `.${LINE_STATUS_GUTTER_CLASS} > .cm-gutterElement`,
  ) as NodeListOf<HTMLElement>;
}

describe("Line Status Gutter Extension (CodeMirror)", () => {
  let container: HTMLElement;
  let view: EditorView;
  beforeEach(() => {
    container = document.createElement("div");
    /** create the code mirror instance before each test.
     * The doc length is irrelevant here.
     */
    view = newEditorView(doc, container);
  });

  it("renders an empty div if there are no line statuses", () => {
    // filter out the initialSpacer element, which maintains visibility: hidden.

    const lineStatuses = getLineStatusElements(container);

    const initialSpacer = lineStatuses[0];
    // the spacer should be hidden
    expect(initialSpacer.style.visibility).toBe("hidden");
    // the number of lines in the document + 1 for the spacer
    expect(lineStatuses.length).toBe(1);
  });

  it("renders a set of line statuses", () => {
    setLineStatuses([{ line: 2, level: "error" }], view);

    const lineStatuses = getLineStatusElements(container);
    const initialSpacer = lineStatuses[0];
    // the spacer should be hidden
    expect(initialSpacer.style.visibility).toBe("hidden");
    // the number of lines in the document + 1 for the spacer
    expect(lineStatuses.length).toBe(2);
  });
  it("re-renders on calls to setLineStatuses", () => {
    setLineStatuses([{ line: 2, level: "error" }], view);
    setLineStatuses(
      [
        { line: 2, level: "error" },
        { line: 3, level: "error" },
      ],
      view,
    );

    let lineStatuses = getLineStatusElements(container);
    const initialSpacer = lineStatuses[0];
    // the spacer should be hidden
    expect(initialSpacer.style.visibility).toBe("hidden");
    // the number of lines in the document + 1 for the spacer
    expect(lineStatuses.length).toBe(3);

    // remove all line statuses to reset.
    setLineStatuses([], view);
    lineStatuses = getLineStatusElements(container);
    // this should only contain the spacer
    expect(lineStatuses.length).toBe(1);
  });

  it("does not render line statuses whose line number is greater than the number of lines in the doc", () => {
    setLineStatuses([{ line: 100, level: "error" }], view);

    const lineStatuses = getLineStatusElements(container);
    const initialSpacer = lineStatuses[0];
    // the spacer should be hidden
    expect(initialSpacer.style.visibility).toBe("hidden");
    // the number of lines in the document + 1 for the spacer
    expect(lineStatuses.length).toBe(1);
  });
});
