import { beforeEach, describe, expect, it } from "vitest";
import {
  loadEphemeralMeasureLibrary,
  mergeEphemeralMeasureDefs,
  saveEphemeralMeasureLibrary,
  syncEphemeralMeasureLibrary,
  upsertIntoEphemeralMeasureLibrary,
} from "./library";
import type { EphemeralMeasureDef } from "./types";

const profit: EphemeralMeasureDef = {
  name: "profit",
  displayName: "Profit",
  expression: "revenue - cost",
};
const arpu: EphemeralMeasureDef = {
  name: "arpu",
  displayName: "ARPU",
  expression: "revenue / users",
  formatPreset: "currency_usd",
};

describe("ephemeral measure library", () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it("round-trips definitions per metrics view and namespace", () => {
    saveEphemeralMeasureLibrary("mv", "org__proj__", [profit, arpu]);
    expect(loadEphemeralMeasureLibrary("mv", "org__proj__")).toEqual([
      profit,
      arpu,
    ]);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([]);
    expect(loadEphemeralMeasureLibrary("other", "org__proj__")).toEqual([]);
  });

  it("ignores malformed stored values", () => {
    localStorage.setItem("rill:app:adhoc-measures:mv", "not json");
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([]);

    localStorage.setItem(
      "rill:app:adhoc-measures:mv",
      JSON.stringify([profit, { name: "broken" }, "x"]),
    );
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([profit]);
  });

  it("upsert adds and replaces without removing", () => {
    saveEphemeralMeasureLibrary("mv", undefined, [profit]);
    const editedProfit = { ...profit, displayName: "Net profit" };
    upsertIntoEphemeralMeasureLibrary("mv", undefined, [editedProfit, arpu]);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([
      editedProfit,
      arpu,
    ]);

    // A load that lacks a definition must not delete it.
    upsertIntoEphemeralMeasureLibrary("mv", undefined, [arpu]);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([
      editedProfit,
      arpu,
    ]);
    upsertIntoEphemeralMeasureLibrary("mv", undefined, undefined);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([
      editedProfit,
      arpu,
    ]);
  });

  it("sync removes definitions dropped since the previous state", () => {
    saveEphemeralMeasureLibrary("mv", undefined, [profit, arpu]);
    // profit deleted by the user, arpu edited.
    const editedArpu = { ...arpu, expression: "revenue / active_users" };
    syncEphemeralMeasureLibrary("mv", undefined, [profit, arpu], [editedArpu]);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([editedArpu]);

    // Deleting the last definition clears the entry.
    syncEphemeralMeasureLibrary("mv", undefined, [editedArpu], undefined);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([]);
    expect(localStorage.getItem("rill:app:adhoc-measures:mv")).toBeNull();
  });

  it("sync keeps definitions from other explores on the same metrics view", () => {
    // arpu was created in another explore and is not in this state at all.
    saveEphemeralMeasureLibrary("mv", undefined, [arpu]);
    syncEphemeralMeasureLibrary("mv", undefined, undefined, [profit]);
    expect(loadEphemeralMeasureLibrary("mv", undefined)).toEqual([
      arpu,
      profit,
    ]);
  });

  it("merge unions by name with base order and overrides", () => {
    const editedProfit = { ...profit, displayName: "Net profit" };
    expect(mergeEphemeralMeasureDefs([profit], [arpu, profit])).toEqual([
      profit,
      arpu,
    ]);
    expect(mergeEphemeralMeasureDefs([profit], [arpu], [editedProfit])).toEqual(
      [editedProfit, arpu],
    );
  });
});
