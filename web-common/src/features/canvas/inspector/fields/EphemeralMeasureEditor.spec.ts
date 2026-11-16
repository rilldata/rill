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

// A canvas component whose file writes fail after changing its spec in
// memory, as a rejected PutFile does.
function failingComponent() {
  const specStore = writable<Record<string, unknown>>(SPEC);
  const revertChanges = vi.fn();
  const component = {
    specStore,
    parent: { fileArtifact: { revertChanges } },
    updateProperty: vi.fn((key: string, value: unknown) => {
      specStore.update((spec) => ({ ...spec, [key]: value }));
      return Promise.resolve(false);
    }),
    updateProperties: vi.fn((patch: Record<string, unknown>) => {
      specStore.update((spec) => ({ ...spec, ...patch }));
      return Promise.resolve(false);
    }),
  };
  return { component, specStore, revertChanges };
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
    const { component, specStore, revertChanges } = failingComponent();
    const onClose = renderEditor(component);

    await fireEvent.click(screen.getByRole("button", { name: "Save" }));

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_save_failed()),
    ).toBeInTheDocument();
    expect(component.updateProperty).toHaveBeenCalled();
    expect(revertChanges).toHaveBeenCalledTimes(1);
    expect(get(specStore)).toEqual(SPEC);
    expect(onClose).not.toHaveBeenCalled();
  });

  it("undoes a delete that failed", async () => {
    const { component, specStore, revertChanges } = failingComponent();
    renderEditor(component);

    await fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_delete_failed()),
    ).toBeInTheDocument();
    expect(component.updateProperties).toHaveBeenCalled();
    expect(revertChanges).toHaveBeenCalledTimes(1);
    expect(get(specStore)).toEqual(SPEC);
  });
});
