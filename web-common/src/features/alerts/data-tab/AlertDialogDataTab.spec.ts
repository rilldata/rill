import type { AlertFormValues } from "@rilldata/web-common/features/alerts/form-utils";
import type { ExpressionFilterManager } from "@rilldata/web-common/features/dashboards/filters/ExpressionFilterManager.svelte.ts";
import {
  MeasureFilterOperation,
  MeasureFilterType,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-options";
import type { DashboardConfigProvider } from "@rilldata/web-common/features/dashboards/providers/DashboardConfigProvider.svelte.ts";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
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

const profitMeasure = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};

function renderDataTab(
  values: Partial<AlertFormValues> = {},
  aiConditionAllowed = false,
) {
  const form = writable({
    metricsViewName: "orders",
    exploreName: "orders_explore",
    measure: "revenue",
    splitByDimension: "country",
    notifyRowsLimit: "5",
    criteria: [percentOfTotalCriteria, valueCriteria],
    ...values,
  } as AlertFormValues);
  render(AlertDialogDataTab, {
    props: {
      superFormInstance: {
        form,
        errors: writable({}),
      } as unknown as SuperForm<AlertFormValues>,
      expressionFilterManager: {} as ExpressionFilterManager,
      timeFilterManager: {} as TimeFilterManager,
      dashboardConfigProvider: {} as DashboardConfigProvider,
      aiConditionAllowed,
    },
  });
  return form;
}

async function openSelect(label: string) {
  await fireEvent.keyDown(screen.getByLabelText(label), { key: "Enter" });
}

async function pickSplitBy(label: string) {
  await openSelect(m.alert_form_data_split_by());
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

  // Percent-of-total criteria stay; validation blocks saving them without a split.
  it("clears the split and keeps the criteria when None is picked", async () => {
    const form = renderDataTab();

    await pickSplitBy(m.alert_form_data_none());

    expect(get(form).splitByDimension).toBe("");
    expect(get(form).criteria).toEqual([percentOfTotalCriteria, valueCriteria]);
  });

  // The AI resolver only gets measure names, so the analyst can't resolve an ephemeral measure's formula.
  it("drops an ephemeral measure when switching to an AI condition", async () => {
    const form = renderDataTab(
      { measure: profitMeasure.name, ephemeralMeasures: [profitMeasure] },
      true,
    );

    await openSelect(m.alert_form_data_measure());
    expect(
      await screen.findByRole("option", { name: profitMeasure.displayName }),
    ).toBeInTheDocument();
    await fireEvent.keyDown(
      screen.getByLabelText(m.alert_form_data_measure()),
      {
        key: "Escape",
      },
    );

    await fireEvent.click(
      screen.getByRole("button", { name: m.alert_form_condition_type_ai() }),
    );

    expect(get(form).measure).toBe("");
    await openSelect(m.alert_form_data_measure());
    expect(
      await screen.findByRole("option", { name: "Revenue" }),
    ).toBeInTheDocument();
    expect(
      screen.queryByRole("option", { name: profitMeasure.displayName }),
    ).toBeNull();
  });
});
