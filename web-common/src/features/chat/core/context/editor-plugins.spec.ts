import { beforeEach, describe, expect, it, type Mock, vi } from "vitest";
import { readable } from "svelte/store";
import { Editor } from "@tiptap/core";
import { TextSelection, type Transaction } from "@tiptap/pm/state";
import { getEditorPlugins } from "@rilldata/web-common/features/chat/core/context/editor-plugins.svelte.ts";
import InlineContextPicker from "@rilldata/web-common/features/chat/core/context/picker/InlineContextPicker.svelte";
import {
  type InlineContext,
  InlineContextType,
} from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

const { mountedPickers } = vi.hoisted(() => ({
  mountedPickers: new Set<Record<string, unknown>>(),
}));

// Pickers need a runtime client, so only track which pickers are open.
vi.mock("svelte", async (importOriginal) => {
  const svelte = await importOriginal<typeof import("svelte")>();
  return {
    ...svelte,
    getAllContexts: () => new Map(),
    mount: (
      component: unknown,
      { props }: { props: Record<string, unknown> },
    ) => {
      const comp = { props, closeDropdown: () => {} };
      if (component === InlineContextPicker) mountedPickers.add(comp);
      return comp;
    },
    unmount: (comp: Record<string, unknown>) => mountedPickers.delete(comp),
  };
});

describe("editor plugins", () => {
  let editor: Editor;
  let onSubmit: Mock<() => void>;

  beforeEach(() => {
    mountedPickers.clear();
    onSubmit = vi.fn<() => void>();
    const element = document.body.appendChild(document.createElement("div"));
    editor = new Editor({
      element,
      extensions: getEditorPlugins({
        placeholder: "",
        onSubmit,
        // Only which picker is open is asserted here, so its options are never read.
        skillOptions: () => readable([]),
      }),
    });
    return () => {
      editor.destroy();
      element.remove();
    };
  });

  // Suggestions open and close asynchronously after a transaction.
  async function dispatch(tr: Transaction) {
    editor.view.dispatch(tr);
    await new Promise((resolve) => setTimeout(resolve));
  }

  // Inserts text at the cursor the way typing does, without scrolling into view (not available in jsdom).
  function type(text: string) {
    return dispatch(editor.state.tr.insertText(text));
  }

  function pressEnter() {
    const event = new KeyboardEvent("keydown", { key: "Enter" });
    editor.view.someProp("handleKeyDown", (f) => f(editor.view, event));
  }

  function openPickers() {
    return [...mountedPickers].map((c) => c.props as Record<string, unknown>);
  }

  it("opens the skills picker with /", async () => {
    await type("/mon");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(false);
  });

  it("does not open the skills picker while the @ picker is open", async () => {
    await type("@orders");
    expect(openPickers()).toHaveLength(1);

    await type(" /mon");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(true);
    expect(openPickers()[0].searchText).toBe("orders /mon");

    pressEnter();
    expect(onSubmit).not.toHaveBeenCalled();
  });

  it("opens a single picker when both triggers match at once", async () => {
    await type("@orders /mon");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(true);
  });

  it("opens the skills picker after a space, but not inside a word", async () => {
    await type("close the month /mon");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(false);

    await type(" 12/08");
    expect(openPickers()).toHaveLength(0);

    pressEnter();
    expect(onSubmit).toHaveBeenCalledOnce();
  });

  it("starts a skill after the word the cursor is on", async () => {
    await type("hola");
    editor.commands.startSkill();
    await dispatch(editor.state.tr);
    expect(editor.getText()).toBe("hola /");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(false);

    // Mention collapses the DOM selection after inserting, which jsdom only has once something is selected.
    document.getSelection()?.selectAllChildren(editor.view.dom);
    (openPickers()[0].onSelect as (ctx: InlineContext) => void)({
      type: InlineContextType.Skill,
      skill: "monthly-close",
      value: "monthly-close",
    });
    await dispatch(editor.state.tr);
    expect(editor.getText()).toBe(
      `hola <chat-reference>type="skill" skill="monthly-close"</chat-reference> `,
    );
  });

  it("does not start a skill while a picker is open", async () => {
    await type("@orders");
    expect(editor.commands.startSkill()).toBe(false);
    await dispatch(editor.state.tr);
    expect(editor.getText()).toBe("@orders");
    expect(openPickers()).toHaveLength(1);
    expect(openPickers()[0].multiple).toBe(true);
  });

  it("submits with Enter once the picker is closed", async () => {
    await type("@orders /mon");
    await dispatch(
      editor.state.tr.setSelection(TextSelection.create(editor.state.doc, 1)),
    );
    expect(openPickers()).toHaveLength(0);

    pressEnter();
    expect(onSubmit).toHaveBeenCalledOnce();
  });
});
