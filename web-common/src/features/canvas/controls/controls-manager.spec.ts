import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import { get, writable } from "svelte/store";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { ControlsManager } from "./controls-manager";

const goto = vi.hoisted(() => vi.fn(() => Promise.resolve()));
vi.mock("$app/navigation", () => ({ goto }));

const SPEC: V1CanvasSpec = {
  controls: [
    {
      name: "total_bids",
      kind: "measure",
      target: "total_bids",
      options: ["total_bids", "spend"],
    },
  ],
  rows: [
    {
      container: {
        name: "deep_dive",
        controls: [
          {
            name: "publisher",
            kind: "filter",
            target: "publisher",
            default: ["Acme"],
            multiple: true,
          },
        ],
        rows: [{ items: [{ component: "chart" }] }],
      },
    },
  ],
};

const ITEM_PATH = ["rows", 0, "rows", 0, "items", 0, "bar_chart"];

function setup(search = "", enabled = true) {
  const params = writable(new URLSearchParams(search));
  const flag = writable(enabled);
  const manager = new ControlsManager(writable(SPEC), params, flag);
  return { manager, params, flag };
}

function values(manager: ControlsManager) {
  return Object.fromEntries(
    get(manager.forPath(() => ITEM_PATH)).map((c) => [c.urlKey, c.values]),
  );
}

describe("ControlsManager", () => {
  beforeEach(() => {
    goto.mockClear();
    window.history.replaceState({}, "", "/canvas?tr=P7D");
  });

  it("gives controls their defaults when the URL doesn't set them", () => {
    const { manager } = setup();
    expect(values(manager)).toEqual({
      "c.deep_dive.publisher": ["Acme"],
      "c.total_bids": ["total_bids"],
    });
    expect(get(manager.changed)).toEqual([]);
    expect(get(manager.invalid)).toEqual([]);
  });

  it("reads values from the URL and follows it as it changes", () => {
    const { manager, params } = setup("c.total_bids=spend");
    expect(values(manager)["c.total_bids"]).toEqual(["spend"]);
    expect(get(manager.changed).map((c) => c.urlKey)).toEqual(["c.total_bids"]);

    params.set(new URLSearchParams("c.deep_dive.publisher=Acme,Globex%20Inc"));
    expect(values(manager)).toEqual({
      "c.deep_dive.publisher": ["Acme", "Globex Inc"],
      "c.total_bids": ["total_bids"],
    });
  });

  it("treats an empty param as a cleared filter", () => {
    const { manager } = setup("c.deep_dive.publisher=");
    expect(values(manager)["c.deep_dive.publisher"]).toEqual([]);
    expect(get(manager.changed)).toHaveLength(1);
  });

  it("falls back to the default for a value that isn't an option, and reports it", () => {
    const { manager } = setup("c.total_bids=revenue");
    expect(values(manager)["c.total_bids"]).toEqual(["total_bids"]);
    expect(get(manager.invalid)).toEqual(["c.total_bids"]);
    expect(get(manager.changed)).toEqual([]);
  });

  it("writes a changed value to the URL and keeps other params", () => {
    const { manager } = setup();
    const [metric] = get(manager.all);
    manager.set(metric, ["spend"]);
    expect(goto).toHaveBeenCalledWith("?tr=P7D&c.total_bids=spend");
  });

  it("drops the param when a control goes back to its default", () => {
    window.history.replaceState({}, "", "/canvas?tr=P7D&c.total_bids=spend");
    const { manager } = setup("tr=P7D&c.total_bids=spend");
    const [metric] = get(manager.all);
    manager.set(metric, ["total_bids"]);
    expect(goto).toHaveBeenLastCalledWith("?tr=P7D");
    manager.set(metric, null);
    expect(goto).toHaveBeenLastCalledWith("?tr=P7D");
  });

  it("writes an empty param when a filter with a default is cleared", () => {
    const { manager } = setup();
    const filter = get(manager.all)[1];
    manager.set(filter, []);
    expect(goto).toHaveBeenCalledWith("?tr=P7D&c.deep_dive.publisher=");
  });

  it("resets every control and leaves other URL state", () => {
    window.history.replaceState(
      {},
      "",
      "/canvas?c.total_bids=spend&tr=P7D&c.deep_dive.publisher=",
    );
    const { manager } = setup("c.total_bids=spend&tr=P7D");
    manager.resetAll();
    expect(goto).toHaveBeenCalledWith("?tr=P7D");
  });

  it("reports the new search before navigating, also when nothing is left", () => {
    window.history.replaceState({}, "", "/canvas?c.total_bids=spend");
    const onNavigate = vi.fn();
    const manager = new ControlsManager(
      writable(SPEC),
      writable(new URLSearchParams("c.total_bids=spend")),
      writable(true),
      onNavigate,
    );
    manager.resetAll();
    expect(onNavigate).toHaveBeenCalledWith("");
    expect(goto).toHaveBeenCalledWith("?");
  });

  it("ignores URL values while controls are disabled", () => {
    const { manager, flag } = setup("c.total_bids=spend", false);
    expect(values(manager)["c.total_bids"]).toEqual(["total_bids"]);
    expect(get(manager.changed)).toEqual([]);

    flag.set(true);
    expect(values(manager)["c.total_bids"]).toEqual(["spend"]);
  });

  describe("locked controls", () => {
    function locked(search: string, values: Record<string, string>) {
      return new ControlsManager(
        writable(SPEC),
        writable(new URLSearchParams(search)),
        writable(true),
        () => {},
        values,
      );
    }

    it("uses the locked value over the URL, keyed with or without the prefix", () => {
      const manager = locked("c.total_bids=total_bids", {
        total_bids: "spend",
        "c.deep_dive.publisher": "Globex",
      });
      expect(values(manager)).toEqual({
        "c.deep_dive.publisher": ["Globex"],
        "c.total_bids": ["spend"],
      });
      expect(get(manager.all).every((c) => c.locked)).toBe(true);
    });

    it("doesn't count a locked control as changed or let viewers set it", () => {
      const manager = locked("", { total_bids: "spend" });
      expect(get(manager.changed)).toEqual([]);
      const [metric, filter] = get(manager.all);
      expect(metric.locked).toBe(true);
      expect(filter.locked).toBe(false);

      manager.set(metric, ["total_bids"]);
      expect(goto).not.toHaveBeenCalled();
      manager.set(filter, []);
      expect(goto).toHaveBeenCalledTimes(1);
    });

    it("keeps the default for a locked value the control doesn't allow", () => {
      const manager = locked("c.total_bids=spend", { total_bids: "revenue" });
      expect(values(manager)["c.total_bids"]).toEqual(["total_bids"]);
      expect(get(manager.all)[0].locked).toBe(true);
    });
  });
});
