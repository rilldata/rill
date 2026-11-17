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

  it("names a retried create after its current name", async () => {
    const netProfit = { ...PROFIT, name: "net_profit", displayName: "Net" };
    const onSave = vi.fn().mockResolvedValue(false);
    renderForm({ onSave, existingDefs: [netProfit] });
    const name = screen.getByLabelText("Name");
    await fireEvent.input(name, { target: { value: "Profit" } });
    await fireEvent.click(screen.getByRole("button", { name: "Revenue" }));
    await fireEvent.click(saveButton());
    await screen.findByText(m.dashboard_pivot_ephemeral_save_failed());

    // Renamed after the failure, onto another definition's name.
    await fireEvent.input(name, { target: { value: "Net profit" } });
    await fireEvent.click(saveButton());

    await waitFor(() => expect(onSave).toHaveBeenCalledTimes(2));
    const retried = (onSave.mock.calls[1] as [EphemeralMeasureDef])[0];
    expect(retried.name).not.toBe("profit");
    expect(retried.name).not.toBe(netProfit.name);
    expect(retried.name).toMatch(/^net_profit/);
  });

  it("recovers when the save callback throws", async () => {
    const onSave = vi.fn().mockRejectedValue(new Error("network"));
    const consoleError = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    renderForm({ editingDef: PROFIT, onSave });

    await fireEvent.click(saveButton());

    expect(
      await screen.findByText(m.dashboard_pivot_ephemeral_save_failed()),
    ).toBeInTheDocument();
    expect(saveButton()).toBeEnabled();
    consoleError.mockRestore();
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
