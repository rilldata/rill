import { cleanup, render, screen } from "@testing-library/svelte";
import { writable } from "svelte/store";
import { afterEach, describe, expect, it } from "vitest";
import type { CanvasEntity } from "../stores/canvas-entity";
import type { AppliedControl } from "./controls";
import LevelControls from "./LevelControls.svelte";

afterEach(cleanup);

// A canvas without controls; only the stores LevelControls reads before a bar renders.
function canvas(enabled: boolean) {
  return {
    controls: {
      enabled: writable(enabled),
      all: writable<AppliedControl[]>([]),
      changed: writable<AppliedControl[]>([]),
    },
  } as unknown as CanvasEntity;
}

const add = {
  levels: [{ kind: "canvas" as const, path: [], name: "" }],
  levelPath: [],
  whenEmpty: true,
};

describe("LevelControls", () => {
  it("offers to add a viewer control to an empty canvas in the editor", () => {
    render(LevelControls, { canvasEntity: canvas(true), scopes: [[]], add });
    expect(
      screen.getByRole("button", { name: "Add viewer control" }),
    ).toBeTruthy();
  });

  it("shows nothing on an empty level without the editor's add target", () => {
    const { container } = render(LevelControls, {
      canvasEntity: canvas(true),
      scopes: [[]],
    });
    expect(container.querySelector(".level-controls")).toBeNull();
  });

  it("leaves an empty block level to its menu", () => {
    render(LevelControls, {
      canvasEntity: canvas(true),
      scopes: [["story"]],
      add: { ...add, whenEmpty: false },
    });
    expect(screen.queryByRole("button")).toBeNull();
  });

  it("hides the button with the canvas_controls flag off", () => {
    render(LevelControls, { canvasEntity: canvas(false), scopes: [[]], add });
    expect(screen.queryByRole("button")).toBeNull();
  });
});
