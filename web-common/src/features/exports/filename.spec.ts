import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { buildExportFilename } from "./filename";

describe("buildExportFilename", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 5, 19, 13, 4, 12));
  });
  afterEach(() => vi.useRealTimers());

  it("slugs the title and appends a local timestamp and the extension", () => {
    expect(buildExportFilename("Sales Overview: Q2", "png")).toBe(
      "sales-overview-q2-20260619-130412.png",
    );
  });

  it("falls back to a generic slug when the title has no usable characters", () => {
    expect(buildExportFilename("  ", "pdf")).toBe(
      "dashboard-20260619-130412.pdf",
    );
  });
});
