import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { writable } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import ControlInput from "./ControlInput.svelte";
import ControlBar from "./ControlBar.svelte";
import type { CanvasEntity } from "../stores/canvas-entity";
import {
  parseControlValue,
  serializeControlValue,
  type AppliedControl,
  type ControlValue,
} from "./controls";

const search = writable<{
  data?: ControlValue[];
  isLoading: boolean;
  error?: Error;
}>({ data: [], isLoading: false });
const dimensionSearch = vi.fn((..._args: unknown[]) => search);
vi.mock(
  "@rilldata/web-common/features/dashboards/filters/dimension-filters/dimension-filter-values",
  () => ({
    useDimensionSearch: (...args: unknown[]) => dimensionSearch(...args),
  }),
);
vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => ({ instanceId: "test" }),
}));

afterEach(cleanup);
beforeEach(() => {
  search.set({ data: [], isLoading: false });
  vi.stubGlobal(
    "ResizeObserver",
    class {
      observe() {}
      unobserve() {}
      disconnect() {}
    },
  );
});

function control(
  values: ControlValue[] | null,
  extra: AppliedControl["control"] = {},
): AppliedControl {
  return {
    control: { name: "country", kind: "filter", target: "country", ...extra },
    values,
    level: "canvas",
    scope: [],
    urlKey: "c.country",
    targetKey: "filter:country",
  };
}
function show(value: AppliedControl) {
  const onChange = vi.fn();
  const view = render(ControlInput, {
    control: value,
    fieldLabel: (name: string) => name,
    metricsViewNames: ["sales"],
    onChange,
  });
  return { onChange, view };
}
async function open() {
  await fireEvent.keyDown(screen.getByRole("button", { name: "country" }), {
    key: "Enter",
  });
}

describe("slot menus", () => {
  // A KPI grid's measures, which viewers pick from three, on their default.
  const kpis = (values: ControlValue[] | null): AppliedControl => ({
    control: {
      name: "measures",
      label: "country",
      kind: "slot",
      target: "measures",
      multiple: true,
      options: ["spend", "clicks", "impressions"],
    },
    values,
    shownValues: ["spend", "clicks"],
    level: "item",
    scope: ["kpis"],
    urlKey: "c.kpis.measures",
    targetKey: "slot:measures",
  });

  it("starts from the fields the widget shows", async () => {
    const { onChange } = show(kpis(null));
    expect(screen.getByRole("button", { name: "country" })).toHaveTextContent(
      "spend +1 other",
    );
    await open();
    const checked = screen
      .getAllByRole("menuitemcheckbox")
      .map((item) => item.getAttribute("aria-checked"));
    expect(checked).toEqual(["true", "true", "false"]);
    await fireEvent.click(
      screen.getByRole("menuitemcheckbox", { name: "impressions" }),
    );
    expect(onChange).toHaveBeenLastCalledWith([
      "spend",
      "clicks",
      "impressions",
    ]);
  });

  it("keeps one field, and goes back to the default with the widget's own fields", async () => {
    const one = show(kpis(["spend"]));
    await open();
    const spend = screen.getByRole("menuitemcheckbox", { name: "spend" });
    expect(spend).toHaveAttribute("aria-disabled", "true");
    await fireEvent.click(spend);
    expect(one.onChange).not.toHaveBeenCalled();
    one.view.unmount();

    const { onChange } = show(kpis(["spend"]));
    await open();
    await fireEvent.click(
      screen.getByRole("menuitemcheckbox", { name: "clicks" }),
    );
    expect(onChange).toHaveBeenLastCalledWith(null);
  });
});

describe("filter value menus", () => {
  it("checks and removes a numeric option whose URL value is text", async () => {
    search.set({ data: [1, 2, "1"], isLoading: false });
    const { onChange } = show(control(["1"], { multiple: true }));
    await open();
    const one = await screen.findByRole("menuitemcheckbox", { name: "1" });
    expect(one).toHaveAttribute("aria-checked", "true");
    expect(screen.getAllByRole("menuitemcheckbox", { name: "1" })).toHaveLength(
      1,
    );
    await fireEvent.click(one);
    expect(onChange).toHaveBeenLastCalledWith([]);
  });

  it("preserves a selected number through the URL", async () => {
    search.set({ data: [1, 2], isLoading: false });
    const { onChange } = show(control([], { multiple: true }));
    await open();
    await fireEvent.click(
      await screen.findByRole("menuitemcheckbox", { name: "2" }),
    );
    expect(onChange).toHaveBeenLastCalledWith([2]);
    const spec = control([]).control;
    expect(
      parseControlValue(
        spec,
        serializeControlValue(spec, onChange.mock.lastCall![0])!,
      ).values,
    ).toEqual([2]);
  });

  it("distinguishes null from the string null", async () => {
    search.set({ data: [null, "null"], isLoading: false });
    const { onChange } = show(control([null], { multiple: true }));
    await open();
    expect(
      await screen.findByRole("menuitemcheckbox", { name: "(null)" }),
    ).toHaveAttribute("aria-checked", "true");
    const literal = screen.getByRole("menuitemcheckbox", { name: "null" });
    expect(literal).toHaveAttribute("aria-checked", "false");
    await fireEvent.click(literal);
    expect(onChange).toHaveBeenLastCalledWith([null, "null"]);
  });

  it("clears a selected single filter even when it has a default", async () => {
    const { onChange } = show(
      control(["Acme"], { options: ["Acme", "Globex"], default: ["Acme"] }),
    );
    await open();
    await fireEvent.click(
      await screen.findByRole("menuitemcheckbox", { name: "Acme" }),
    );
    expect(onChange).toHaveBeenLastCalledWith([]);
  });

  it("keeps a selected field instead of clearing it", async () => {
    const { onChange } = show(
      control(["revenue"], {
        kind: "measure",
        options: ["revenue", "profit"],
        display: "dropdown",
      }),
    );
    await open();
    await fireEvent.click(
      await screen.findByRole("menuitemcheckbox", { name: "revenue" }),
    );
    expect(onChange).toHaveBeenLastCalledWith(["revenue"]);
  });

  it("shows loading instead of an empty result", async () => {
    search.set({ isLoading: true });
    show(control([]));
    await open();
    expect(await screen.findByRole("status")).toHaveTextContent("Loading...");
    expect(screen.queryByText("No values")).not.toBeInTheDocument();
  });

  it("shows failures and retries the affected values queries", async () => {
    search.set({ data: [], isLoading: false, error: new Error("offline") });
    const retry = vi.spyOn(queryClient, "refetchQueries").mockResolvedValue();
    show(control([]));
    await open();
    expect(await screen.findByRole("alert")).toHaveTextContent(
      "Could not load values.",
    );
    expect(screen.queryByText("No values")).not.toBeInTheDocument();
    await fireEvent.click(screen.getByRole("menuitem", { name: "Retry" }));
    expect(retry).toHaveBeenCalledWith({
      queryKey: [
        "QueryService",
        "metricsViewAggregation",
        "test",
        { metricsView: "sales", dimensions: [{ name: "country" }] },
      ],
    });
    retry.mockRestore();
  });
});

describe("going back to the default", () => {
  const range = (values: ControlValue[] | null, extra = {}) =>
    control(values, {
      kind: "time_range",
      target: "",
      label: "country",
      options: ["PT24H", "P7D", "P14D", "P4W", "P3M"],
      ...extra,
    });

  it("offers a time control's inherited range as an option, not a reset", async () => {
    const { onChange } = show(range(["P7D"]));
    expect(screen.getByRole("button", { name: "country" })).toHaveTextContent(
      "Last 7 Days",
    );
    await open();
    const inherit = await screen.findByRole("menuitemcheckbox", {
      name: "Inherit from canvas",
    });
    expect(inherit).toHaveAttribute("aria-checked", "false");
    expect(screen.queryByRole("menuitem", { name: "Reset" })).toBeNull();
    await fireEvent.click(inherit);
    expect(onChange).toHaveBeenLastCalledWith(null);
  });

  it("checks the inherited range on the default, and has none with a default of its own", async () => {
    const inheriting = show(range(null));
    expect(screen.getByRole("button", { name: "country" })).toHaveTextContent(
      "Inherit from canvas",
    );
    await open();
    expect(
      await screen.findByRole("menuitemcheckbox", {
        name: "Inherit from canvas",
      }),
    ).toHaveAttribute("aria-checked", "true");
    inheriting.view.unmount();

    show(range(["P14D"], { default: ["P7D"] }));
    await open();
    await screen.findByRole("menuitemcheckbox", { name: "Last 7 Days" });
    expect(
      screen.queryByRole("menuitemcheckbox", { name: "Inherit from canvas" }),
    ).toBeNull();
  });

  it("goes back to the inherited range when the picked button is pressed again", async () => {
    const { onChange } = show(
      range(["P7D"], { options: ["P7D", "P14D"], display: "buttons" }),
    );
    await fireEvent.click(screen.getByRole("button", { name: "Last 7 Days" }));
    expect(onChange).toHaveBeenLastCalledWith(null);
    await fireEvent.click(screen.getByRole("button", { name: "Last 14 Days" }));
    expect(onChange).toHaveBeenLastCalledWith(["P14D"]);
  });

  it("clears a multi-select filter from the footer", async () => {
    search.set({ data: ["Acme", "Globex"], isLoading: false });
    const { onChange } = show(control(["Acme"], { multiple: true }));
    await open();
    await fireEvent.click(
      await screen.findByRole("menuitem", { name: "Clear selection" }),
    );
    expect(onChange).toHaveBeenLastCalledWith(null);
  });
});

it("queries values only from widgets affected by this level control, within the filter bar", async () => {
  const publisher = control([]);
  const inherited = writable([publisher]);
  // The filter bar filters the other dimensions of "inside"; a filter on "country" itself is left out.
  const deviceFilter = { cond: { op: "OPERATION_IN" } };
  const getOtherDimensionsFilter = vi.fn((name: string, view: string) =>
    name === "country" && view === "inside" ? deviceFilter : undefined,
  );
  const entity = {
    controls: {
      enabled: writable(true),
      changed: writable([]),
      hovered: writable(null),
    },
    exportMode: writable(false),
    _metricsViews: writable({
      inside: { state: { validSpec: { dimensions: [{ name: "country" }] } } },
      outside: { state: { validSpec: { dimensions: [{ name: "country" }] } } },
    }),
    componentsStore: writable(
      new Map([
        ["affected", { metricsViewName: "inside", controls: inherited }],
        ["unaffected", { metricsViewName: "outside", controls: writable([]) }],
        ["ignored", { metricsViewName: "outside", controls: writable([]) }],
      ]),
    ),
    timeFilterManager: { timeStart: undefined, timeEnd: undefined },
    expressionFilterManager: { getOtherDimensionsFilter },
  } as unknown as CanvasEntity;
  render(ControlBar, { controls: [publisher], canvasEntity: entity });
  await open();
  await waitFor(() =>
    expect(dimensionSearch).toHaveBeenLastCalledWith(
      expect.anything(),
      ["inside"],
      "country",
      expect.objectContaining({
        metricsViewWheres: new Map([["inside", deviceFilter]]),
      }),
    ),
  );
  inherited.set([]);
  await waitFor(() =>
    expect(dimensionSearch).toHaveBeenLastCalledWith(
      expect.anything(),
      [],
      "country",
      expect.anything(),
    ),
  );
});
