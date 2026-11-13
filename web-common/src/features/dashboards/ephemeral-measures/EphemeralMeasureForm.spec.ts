import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import EphemeralMeasureForm from "./EphemeralMeasureForm.svelte";
import type { EphemeralMeasureDef } from "./types";

const MEASURES: MetricsViewSpecMeasure[] = [
  { name: "revenue", displayName: "Revenue" },
  { name: "cost", displayName: "Cost" },
];
const PROFIT: EphemeralMeasureDef = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};

afterEach(cleanup);

function renderForm(props: Record<string, unknown>) {
  return render(EphemeralMeasureForm, {
    props: {
      metricsView: { measures: MEASURES, dimensions: [] },
      referenceableMeasures: MEASURES,
      existingDefs: [],
      onSave: vi.fn(),
      onClose: vi.fn(),
      ...props,
    },
  });
}

const saveButton = () => screen.getByRole("button", { name: "Save" });

describe("EphemeralMeasureForm persistence", () => {
  it("keeps the dialog open with an error when the save fails", async () => {
    const onSave = vi.fn().mockResolvedValue(false);
    const onClose = vi.fn();
    renderForm({ editingDef: PROFIT, onSave, onClose });

    await fireEvent.click(saveButton());

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_save_failed()),
    ).toBeInTheDocument();
    expect(onSave).toHaveBeenCalledWith(PROFIT);
    expect(onClose).not.toHaveBeenCalled();
    expect(saveButton()).toBeEnabled();
  });

  it("retries a failed create under the same name", async () => {
    const onSave = vi.fn().mockResolvedValue(false);
    const view = renderForm({ onSave });
    await fireEvent.input(screen.getByLabelText("Name"), {
      target: { value: "Profit" },
    });
    await fireEvent.click(screen.getByRole("button", { name: "Revenue" }));
    await fireEvent.click(saveButton());
    await screen.findByText(m.dashboard_pivot_ephemeral_save_failed());
    const [first] = onSave.mock.calls[0] as [EphemeralMeasureDef];

    // The failed canvas write may still have added the definition to the
    // component's spec; the retry must not add a second one.
    await view.rerender({ existingDefs: [first] });
    await fireEvent.click(saveButton());

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(2));
    expect((onSave.mock.calls[1] as [EphemeralMeasureDef])[0].name).toBe(
      first.name,
    );
  });

  it("reports a failed delete", async () => {
    const onRemove = vi.fn().mockResolvedValue(false);
    renderForm({ editingDef: PROFIT, onRemove });

    await fireEvent.click(screen.getByRole("button", { name: "Delete" }));

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_delete_failed()),
    ).toBeInTheDocument();
  });
});
