import { describe, expect, it } from "vitest";
import { get } from "svelte/store";
import { deletePreviewArgsStore, getPreviewArgsStore } from "./preview-state";

describe("preview state", () => {
  it("drops bindings when a file path is deleted or renamed", () => {
    const path = "/viz_library/chart.yaml";
    const original = getPreviewArgsStore(path);
    original.set({ metrics_view: "mv1" });

    deletePreviewArgsStore(path);

    const recreated = getPreviewArgsStore(path);
    expect(recreated).not.toBe(original);
    expect(get(recreated)).toEqual({});
  });
});
