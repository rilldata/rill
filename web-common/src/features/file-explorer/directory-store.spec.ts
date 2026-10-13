import { get } from "svelte/store";
import { beforeEach, describe, expect, it } from "vitest";
import { directoryState, isDirectoryExpanded } from "./directory-store";

describe("directoryState", () => {
  beforeEach(() => {
    directoryState.reset();
  });

  it("expands visible directories by default", () => {
    const state = get(directoryState);
    expect(isDirectoryExpanded(state, "/")).toBe(true);
    expect(isDirectoryExpanded(state, "/models")).toBe(true);
    expect(isDirectoryExpanded(state, "/models/staging")).toBe(true);
  });

  it("collapses hidden directories and their contents by default", () => {
    const state = get(directoryState);
    expect(isDirectoryExpanded(state, "/.claude")).toBe(false);
    expect(isDirectoryExpanded(state, "/.claude/skills")).toBe(false);
    expect(isDirectoryExpanded(state, "/models/.scratch")).toBe(false);
  });

  it("toggles a hidden directory open from its collapsed default", () => {
    directoryState.toggle("/.claude");
    expect(isDirectoryExpanded(get(directoryState), "/.claude")).toBe(true);
    // Nested directories keep their own default until toggled themselves.
    expect(isDirectoryExpanded(get(directoryState), "/.claude/skills")).toBe(
      false,
    );

    directoryState.toggle("/.claude");
    expect(isDirectoryExpanded(get(directoryState), "/.claude")).toBe(false);
  });

  it("expands hidden ancestors when expanding a nested path", () => {
    directoryState.expand("/.claude/skills/analyst");
    const state = get(directoryState);
    expect(isDirectoryExpanded(state, "/.claude")).toBe(true);
    expect(isDirectoryExpanded(state, "/.claude/skills")).toBe(true);
    expect(isDirectoryExpanded(state, "/.claude/skills/analyst")).toBe(true);
  });

  it("treats every directory as collapsed after collapseAll", () => {
    const paths = ["/models", "/models/staging", "/.claude", "/.claude/skills"];
    directoryState.collapseAll(paths);
    const state = get(directoryState);
    expect(paths.every((path) => !isDirectoryExpanded(state, path))).toBe(true);

    directoryState.expandAll(paths);
    const expanded = get(directoryState);
    expect(paths.every((path) => isDirectoryExpanded(expanded, path))).toBe(
      true,
    );
  });
});
