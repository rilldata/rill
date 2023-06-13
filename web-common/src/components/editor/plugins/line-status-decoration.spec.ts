import { basicSetup } from "@codemirror/basic-setup";
import { EditorState } from "@codemirror/state";
import { EditorView } from "@codemirror/view";
import { createLineStatusFactory, lineState } from "./line-status-decoration";

describe("lineStatus extension", () => {
  let view;

  beforeEach(() => {
    const lineStatusFactory = createLineStatusFactory();
    view = new EditorView({
      state: EditorState.create({
        doc: "line 1\nline 2\nline 3",
        extensions: [basicSetup, lineStatusFactory.extension],
      }),
    });
  });

  afterEach(() => {
    view.destroy();
  });

  test("creates line status state properly", () => {
    const lineStatus = [
      { line: 1, message: "Error message 1", level: "error" },
      { line: 2, message: "Error message 2", level: "warning" },
    ];

    const lineStatusFactory = createLineStatusFactory();
    lineStatusFactory.update(lineStatus)(view);

    const currentLineStatus = view.state.field(lineState);
    expect(currentLineStatus).toEqual(lineStatus);
  });
});
