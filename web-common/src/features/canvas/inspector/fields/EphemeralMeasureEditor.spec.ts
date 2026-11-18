import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { get, readable, writable } from "svelte/store";
import { afterEach, describe, expect, it, vi } from "vitest";
import EphemeralMeasureEditor from "./EphemeralMeasureEditor.svelte";

vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => ({ instanceId: "default" }),
}));
vi.mock(
  "@rilldata/web-common/features/canvas/state-managers/state-managers",
  () => ({
    getCanvasStore: () => ({
      canvasEntity: {
        metricsView: {
          getMetricsViewFromName: () =>
            readable({
              metricsView: {
                measures: [
                  { name: "revenue", displayName: "Revenue" },
                  { name: "cost", displayName: "Cost" },
                ],
                dimensions: [],
              },
            }),
        },
      },
    }),
  }),
);
afterEach(cleanup);

const PROFIT = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};
const SPEC = {
  metrics_view: "orders",
  measures: ["revenue", "profit"],
  adhoc_measures: [
    { name: "profit", display_name: "Profit", expression: "revenue - cost" },
  ],
};

// A canvas component whose file writes fail after changing the canvas file's
// editor content and the spec in memory, as a rejected PutFile does.
function failingComponent({ unsaved }: { unsaved: boolean }) {
  const specStore = writable<Record<string, unknown>>(SPEC);
  const editorContent = writable<string | null>("before");
  const untouch = vi.fn();
  const fileArtifact = {
    path: "canvas.yaml",
    editorContent,
    hasUnsavedChanges: readable(unsaved),
    updateEditorContent: vi.fn((content: string) => editorContent.set(content)),
    saveState: { untouch },
  };
  const write = (patch: Record<string, unknown>) => {
    editorContent.set("after");
    specStore.update((spec) => ({ ...spec, ...patch }));
    return Promise.resolve(false);
  };
  const component = {
    specStore,
    parent: { fileArtifact },
    updateProperty: vi.fn((key: string, value: unknown) =>
      write({ [key]: value }),
    ),
    updateProperties: vi.fn(write),
  };
  return { component, specStore, editorContent, untouch };
}

function renderEditor(component: unknown, onClose = vi.fn()) {
  render(EphemeralMeasureEditor, {
    props: {
      component: component as never,
      canvasName: "canvas",
      metricName: "orders",
      editingDef: PROFIT,
      onClose,
    },
  });
  return onClose;
}

describe("EphemeralMeasureEditor failed saves", () => {
  it("undoes a save that failed and keeps the dialog open", async () => {
    const { component, specStore, editorContent, untouch } = failingComponent({
      unsaved: false,
    });
    const onClose = renderEditor(component);

    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_save_failed()),
    ).toBeInTheDocument();
    expect(component.updateProperty).toHaveBeenCalled();
    expect(get(editorContent)).toBe("before");
    expect(untouch).toHaveBeenCalledTimes(1);
    expect(get(specStore)).toEqual(SPEC);
    expect(onClose).not.toHaveBeenCalled();
  });

  it("keeps edits that were unsaved before the failed change", async () => {
    const { component, editorContent, untouch } = failingComponent({
      unsaved: true,
    });
    renderEditor(component);

    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    await screen.findByText(m.dashboard_pivot_ephemeral_save_failed());
    expect(get(editorContent)).toBe("before");
    expect(untouch).not.toHaveBeenCalled();
  });

  it("undoes a delete that failed", async () => {
    const { component, specStore, editorContent } = failingComponent({
      unsaved: false,
    });
    renderEditor(component);

    await fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_delete_failed()),
    ).toBeInTheDocument();
    expect(component.updateProperties).toHaveBeenCalled();
    expect(get(editorContent)).toBe("before");
    expect(get(specStore)).toEqual(SPEC);
  });
});
