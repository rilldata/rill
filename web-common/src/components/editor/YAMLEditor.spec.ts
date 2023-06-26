import { render } from "@testing-library/svelte";
import { tick } from "svelte";
import { beforeAll, describe, expect, it, vi } from "vitest";
import YAMLEditor from "./YAMLEditor.svelte";

describe("YAMLEditor.svelte", () => {
  beforeAll(() => {
    HTMLElement.prototype.getClientRects = vi.fn<DOMRect[]>(() => {
      return [
        {
          width: 120,
          height: 120,
          top: 0,
          left: 0,
          bottom: 0,
          right: 0,
          item: null,
        },
      ];
    });
  });

  it("starts with an editor with one line & updates to content prop creates more lines", async () => {
    const { component, unmount, container } = render(YAMLEditor, {
      content: "",
    });

    // cm-content should have a single <div> element, reflecting that there's no content.
    let elems = container.querySelector(".cm-content").children;
    expect(elems?.length).toBe(1);
    component.$set({
      content: "name: test\nfield: whatever\nanother_field: another",
    });
    elems = container.querySelector(".cm-content").children;
    await tick();
    expect(elems?.length).toBe(3);
    unmount();
  });

  it("successfully fires the state update functions when the content changes", async () => {
    const update = vi.fn();
    const { component, unmount } = render(YAMLEditor, {
      content: "",
      stateFieldUpdaters: [update],
    });
    component.$set({
      content: "name: test\nfield: whatever\nanother_field: another",
    });
    await tick();
    expect(update).toHaveBeenCalled();
    unmount();
  });
});
