import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { tick } from "svelte";
import { writable } from "svelte/store";
import { afterEach, describe, expect, it, vi } from "vitest";
import { parseDocument } from "yaml";
import type { FileArtifact } from "../../entity-management/file-artifact";
import type { CanvasEntity } from "../stores/canvas-entity";
import type { RowsBlock } from "../stores/layout-blocks";
import ContainerEditor from "./ContainerEditor.svelte";

const YAML = `type: canvas
rows:
  - name: story
    rows: []
`;

afterEach(cleanup);

function setup() {
  const fileArtifact = {
    editorContent: writable(YAML),
    updateEditorContent: vi.fn(),
    saveLocalContent: vi.fn(() => Promise.resolve()),
  } as unknown as FileArtifact;
  const canvas = {
    parsedContent: writable(parseDocument(YAML)),
    hiddenPaths: writable(new Set<string>()),
    fileArtifact,
    setSelectedBlock: vi.fn(),
    renameBlockKey: vi.fn(),
  } as unknown as CanvasEntity;
  const props = (name: string) => ({
    container: { name } as RowsBlock,
    blockIndex: 0,
    canvas,
    fileArtifact,
    autoSave: true,
  });
  const view = render(ContainerEditor, props("story"));
  const input = screen.getByLabelText("Container name") as HTMLInputElement;
  return { view, input, props };
}

describe("ContainerEditor", () => {
  it("keeps a name being typed when the layout hands it a new block with the same name", async () => {
    const { view, input, props } = setup();
    await fireEvent.input(input, { target: { value: "revenue_story" } });

    // A reconcile of an unrelated edit rebuilds the layout's block objects.
    await view.rerender(props("story"));
    await tick();
    expect(input.value).toBe("revenue_story");

    // A rename of the block still resets the draft.
    await view.rerender(props("story_2"));
    await tick();
    expect(input.value).toBe("story_2");
  });
});
