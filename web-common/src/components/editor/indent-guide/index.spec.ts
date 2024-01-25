import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { beforeEach, describe, expect, it } from "vitest";
import { indentGuide } from ".";

/** convenience function for creating a new CodeMirror editor  */
function newEditorView(content: string, parent: HTMLElement) {
  return new EditorView({
    state: EditorState.create({
      doc: content,
      extensions: [indentGuide()],
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

describe("Indent Guide Extension (CodeMirror)", () => {
  let container: HTMLElement;
  beforeEach(() => {
    container = document.createElement("div");
  });

  it("renders indent guides for different levels of indentation", () => {
    const view = newEditorView("test", container);
    let guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    // one line
    expect(guides.length).toBe(1);
    // zero indent guides
    expect(guides[0].length).toBe(0);

    // add a single space – creates one indent guide.
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: " test",
      },
    });

    guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    // one line starting with space
    // one indent guide
    expect(guides.length).toBe(1);
    expect(guides.flat(Infinity).length).toBe(1);

    // add a second space – should still be one indent guide.
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: "  test",
      },
    });
    guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    // one line starting with space
    // one indent guide
    expect(guides.length).toBe(1);
    expect(guides.flat(Infinity).length).toBe(1);

    // add a third space – should now be two indent guides for this line.
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: "   test",
      },
    });
    guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    // one line starting with space
    // one indent guide
    expect(guides.length).toBe(1);
    expect(guides.flat(Infinity).length).toBe(2);

    // add a fourth space – should still be two indent guides for this line.
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        insert: "    test",
      },
    });
    guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    // one line starting with space
    // one indent guide
    expect(guides.length).toBe(1);
    expect(guides.flat(Infinity).length).toBe(2);

    // multiple lines should have multiple indent guides.
    view.dispatch({
      changes: {
        from: 0,
        to: view.state.doc.length,
        // one for first line, 3 for second
        insert: `  test
      another test`,
      },
    });
    guides = Array.from(getLines(container)).map((line) =>
      Array.from(line.querySelectorAll(".cm-indent-guide")),
    );
    expect(guides[0]?.length).toBe(1);
    expect(guides[1]?.length).toBe(3);
  });
});
