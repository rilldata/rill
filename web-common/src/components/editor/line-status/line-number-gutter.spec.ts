import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { beforeEach, describe, expect, it } from "vitest";
import { lineStatus } from ".";
import { LINE_NUMBER_GUTTER_CLASS } from "./line-number-gutter";

const doc = `name: test dashboard
measures:
  - label: test measure
    description: test measure
    expression: count(*)
`;

function newEditorView(content: string, parent: HTMLElement) {
  return new EditorView({
    state: EditorState.create({
      doc: content,
      extensions: [lineStatus()],
    }),
    parent,
  });
}

function getLineNumberElements(container: HTMLElement) {
  return container.querySelectorAll(
    `.${LINE_NUMBER_GUTTER_CLASS} > .cm-gutterElement`,
  ) as NodeListOf<HTMLElement>;
}

describe("Line Number Gutter Extension (CodeMirror)", () => {
  let container: HTMLElement;
  beforeEach(() => {
    container = document.createElement("div");
  });

  it("renders a line gutter with the supplied lines & a spacer", () => {
    // filter out the initialSpacer element, which maintains visibility: hidden.
    newEditorView(doc, container);

    const lineNumbers = getLineNumberElements(container);

    const initialSpacer = lineNumbers[0];
    // the spacer should be hidden
    expect(initialSpacer.style.visibility).toBe("hidden");
    // the spacer text should be 6
    expect(initialSpacer.textContent).toBe("6");
    // the number of lines in the document + 1 for the spacer
    expect(lineNumbers.length - 1).toBe(doc.split("\n").length);
  });

  it("handles an empty buffer on instantiation", () => {
    newEditorView("", container);

    const lineNumbers = getLineNumberElements(container);

    expect(lineNumbers.length).toBe(2);
    // expose only the spacer, since this is a one-line document.
    expect(lineNumbers[0].style.visibility).toBe("hidden");
    // the spacer's largest value should be the largest one it has ever seen,
    // which in this scope is still 6 liens.
    expect(lineNumbers[0].textContent).toBe("1");
    // there should only be a 1 on the next line.
    expect(lineNumbers[1].textContent).toBe("1");
  });

  it("handles update to more lines by adding more markers and updating the spacer", () => {
    const view = newEditorView(doc, container);
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: doc + "\n",
      },
    });
    const lineNumbers = getLineNumberElements(container);
    expect(lineNumbers.length).toBe(doc.split("\n").length + 2);
    expect(lineNumbers[0].textContent).toBe("7");
  });

  it("handles update to fewer lines by removing markers and updating the spacer", () => {
    const view = newEditorView(doc, container);
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: doc.split("\n").slice(0, -1).join("\n"),
      },
    });

    const lineNumbers = getLineNumberElements(container);
    expect(lineNumbers.length).toBe(doc.split("\n").length);
    expect(lineNumbers[0].textContent).toBe("5");
  });
});
