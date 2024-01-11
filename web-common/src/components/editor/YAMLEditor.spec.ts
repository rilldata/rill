import { render } from "@testing-library/svelte";
import { tick } from "svelte";
import { beforeAll, describe, expect, it, vi } from "vitest";
import YAMLEditor from "./YAMLEditor.svelte";

function getLines(container: HTMLElement) {
  return Array.from(container.querySelectorAll(".cm-line")).map(
    (l) => l.textContent,
  );
}

describe("YAMLEditor.svelte", () => {
  beforeAll(() => {
    document.createRange = () => {
      const range = new Range();

      range.getBoundingClientRect = vi.fn();

      range.getClientRects = () => {
        return {
          item: () => null,
          length: 0,
          [Symbol.iterator]: vi.fn(),
        };
      };

      return range;
    };
  });

  it("starts with an editor with one line & updates to content prop creates more lines", async () => {
    const { component, unmount, container } = render(YAMLEditor, {
      content: "",
    });
    const propMap = component?.$$?.props;
    const viewInstance = component.$$.ctx[propMap["view"]];

    expect(getLines(container)).toHaveLength(1);

    const onUpdate = vi.fn();
    component?.$on("update", onUpdate);

    const content = "foo: 10\nbar: 20\nfoo: 10\nbar: 20";

    viewInstance.dispatch({
      changes: {
        from: 0,
        to: 0,
        insert: content,
      },
    });
    await tick();

    expect(getLines(container)).toHaveLength(4);
    expect(onUpdate).toHaveBeenCalledOnce();
    unmount();
  });
});
