import {
  clearExternalHover,
  setExternalHover,
} from "@rilldata/web-common/features/components/charts/highlight-controller";
import { describe, expect, it, vi } from "vitest";

// Mock Vega View that tracks signal calls
function createMockView(signals: Record<string, unknown> = {}) {
  const state = { ...signals };
  return {
    signal(name: string, value?: unknown) {
      if (arguments.length === 1) {
        if (!(name in state)) {
          throw new Error(`Unrecognized signal name: "${name}"`);
        }
        return state[name];
      }
      state[name] = value;
      return this;
    },
    runAsync: vi.fn().mockResolvedValue(undefined),
    _state: state,
  };
}

describe("setExternalHover", () => {
  it("sets hover_tuple with matching values/fields length (x-only selection)", async () => {
    const time = new Date("2024-06-15T00:00:00Z");

    const view = createMockView({
      hover_tuple_fields: [{ type: "E", field: "yearmonthdate_timestamp" }],
      hover_tuple: null,
    });

    // Even though dimensionValue is provided, values should only contain
    // epochTime because hover_tuple_fields has only 1 entry (x-encoding only)
    setExternalHover(view as any, time, "US");

    const tuple = view._state.hover_tuple as any;
    expect(tuple.values).toHaveLength(1);
    expect(tuple.fields).toHaveLength(1);
    expect(tuple.values[0]).toBe(time.getTime());
  });

  it("includes dimension value when hover selection has 2 fields", async () => {
    const time = new Date("2024-06-15T00:00:00Z");

    const view = createMockView({
      hover_tuple_fields: [
        { type: "E", field: "yearmonthdate_ts" },
        { type: "E", field: "dimension" },
      ],
      hover_tuple: null,
    });

    setExternalHover(view as any, time, "US");

    const tuple = view._state.hover_tuple as any;
    expect(tuple.values).toHaveLength(2);
    expect(tuple.values[0]).toBe(time.getTime());
    expect(tuple.values[1]).toBe("US");
  });

  it("sets time-only values when dimensionValue is undefined", async () => {
    const time = new Date("2024-06-15T00:00:00Z");

    const view = createMockView({
      hover_tuple_fields: [{ type: "E", field: "yearmonthdate_timestamp" }],
      hover_tuple: null,
    });

    setExternalHover(view as any, time, undefined);

    const tuple = view._state.hover_tuple as any;
    expect(tuple.values).toEqual([time.getTime()]);
    expect(tuple.fields).toHaveLength(1);
  });

  it("does not throw when hover_tuple_fields signal is missing", async () => {
    const time = new Date("2024-06-15T00:00:00Z");

    // View without hover signals (e.g. chart type without hover selection)
    const view = createMockView({});

    expect(() => setExternalHover(view as any, time, undefined)).not.toThrow();
  });

  it("skips update when values haven't changed", async () => {
    const time = new Date("2024-06-15T00:00:00Z");

    const view = createMockView({
      hover_tuple_fields: [{ type: "E", field: "yearmonthdate_timestamp" }],
      hover_tuple: {
        unit: "",
        fields: [{ type: "E", field: "yearmonthdate_timestamp" }],
        values: [time.getTime()],
      },
    });

    setExternalHover(view as any, time, undefined);

    // runAsync should not be called since values didn't change
    expect(view.runAsync).not.toHaveBeenCalled();
  });
});

describe("clearExternalHover", () => {
  it("sets hover_tuple to null", async () => {
    const view = createMockView({
      hover_tuple: { unit: "", fields: [], values: [123] },
    });

    clearExternalHover(view as any);

    expect(view._state.hover_tuple).toBeNull();
    expect(view.runAsync).toHaveBeenCalled();
  });
});
