import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import type { TimeControls } from "@rilldata/web-common/features/dashboards/stores/TimeControls.ts";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { fireEvent, render, screen } from "@testing-library/svelte";
import { get, readable, writable } from "svelte/store";
import type { SuperForm } from "sveltekit-superforms/client";
import { beforeAll, describe, expect, it, vi } from "vitest";

// The filters and the preview query the runtime; this test only covers the measure and split fields.
vi.mock(
  "@rilldata/web-common/features/scheduled-reports/FiltersForm.svelte",
  async () => ({
    default: (
      await import(
        "@rilldata/web-common/components/i18n/__fixtures__/SlotPassthrough.svelte"
      )
    ).default,
  }),
);
vi.mock(
  "@rilldata/web-common/features/alerts/data-tab/DataPreview.svelte",
  async () => ({
    default: (
      await import(
        "@rilldata/web-common/components/i18n/__fixtures__/SlotPassthrough.svelte"
      )
    ).default,
  }),
);
vi.mock("@rilldata/web-common/runtime-client/v2", async (importOriginal) => ({
  ...(await importOriginal<object>()),
  useRuntimeClient: () => ({}),
}));
vi.mock(
  "@rilldata/web-common/features/dashboards/selectors",
  async (importOriginal) => ({
    ...(await importOriginal<object>()),
    useMetricsViewValidSpec: () =>
      readable({
        data: {
          measures: [{ name: "revenue", displayName: "Revenue" }],
          dimensions: [
            { name: "country", displayName: "Country" },
            { name: "device", displayName: "Device" },
          ],
        },
      }),
  }),
);

import AlertDialogDataTab from "./AlertDialogDataTab.svelte";

const percentOfTotalCriteria = {
  measure: "revenue",
  operation: MeasureFilterOperation.GreaterThan,
  type: MeasureFilterType.PercentOfTotal,
  value1: "10",
  value2: "",
};
const valueCriteria = {
  measure: "revenue",
  operation: MeasureFilterOperation.GreaterThan,
  type: MeasureFilterType.Value,
  value1: "100",
  value2: "",
};

function renderDataTab() {
  const form = writable({
    metricsViewName: "orders",
    exploreName: "orders_explore",
    measure: "revenue",
    splitByDimension: "country",
    notifyRowsLimit: "5",
    criteria: [percentOfTotalCriteria, valueCriteria],
  } as AlertFormValues);
  render(AlertDialogDataTab, {
    props: {
      superFormInstance: {
        form,
        errors: writable({}),
      } as unknown as SuperForm<AlertFormValues>,
      filters: {} as ExpressionFilterManager,
      timeControls: {} as TimeControls,
    },
  });
  return form;
}

async function pickSplitBy(label: string) {
  await fireEvent.keyDown(screen.getByLabelText(m.alert_form_data_split_by()), {
    key: "Enter",
  });
  await fireEvent.pointerUp(
    await screen.findByRole("option", { name: label }),
    {
      pointerType: "mouse",
    },
  );
}

describe("AlertDialogDataTab", () => {
  mockAnimationsForComponentTesting();
  beforeAll(() => {
    // jsdom lacks scrollIntoView, which bits-ui calls to highlight the open list.
    HTMLElement.prototype.scrollIntoView ??= () => {};
  });

  it("removes percent-of-total criteria when the split is removed", async () => {
    const form = renderDataTab();

    await pickSplitBy(m.alert_form_data_none());

    expect(get(form).splitByDimension).toBe("");
    expect(get(form).criteria).toEqual([valueCriteria]);
  });

  it("keeps percent-of-total criteria when switching to another dimension", async () => {
    const form = renderDataTab();

    await pickSplitBy("Device");

    expect(get(form).splitByDimension).toBe("device");
    expect(get(form).criteria).toEqual([percentOfTotalCriteria, valueCriteria]);
  });
});
