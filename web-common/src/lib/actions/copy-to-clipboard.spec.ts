// @vitest-environment jsdom
import { afterEach, beforeEach, expect, it, vi } from "vitest";
import { copyToClipboard, tryCopyToClipboard } from "./copy-to-clipboard";

const mocks = vi.hoisted(() => ({ emit: vi.fn(), write: vi.fn() }));
vi.mock("@rilldata/web-common/lib/event-bus/event-bus", () => ({
  eventBus: { emit: mocks.emit },
}));

beforeEach(() => {
  vi.clearAllMocks();
  Object.defineProperty(navigator, "clipboard", {
    configurable: true,
    value: { writeText: mocks.write },
  });
});
afterEach(() => vi.restoreAllMocks());

it("confirms and notifies only after a clipboard write succeeds", async () => {
  let resolve!: () => void;
  mocks.write.mockReturnValueOnce(
    new Promise<void>((done) => (resolve = done)),
  );
  copyToClipboard("value", "Copied");
  expect(mocks.emit).not.toHaveBeenCalled();
  resolve();
  await vi.waitFor(() =>
    expect(mocks.emit).toHaveBeenCalledWith("notification", {
      message: "Copied",
    }),
  );
});

it("reports a rejected write without a success notification", async () => {
  vi.spyOn(console, "error").mockImplementation(() => {});
  mocks.write.mockRejectedValueOnce(new Error("Clipboard permission denied"));
  expect(await tryCopyToClipboard("value")).toBe(false);
  expect(mocks.emit).not.toHaveBeenCalled();
});

it("reports unsupported clipboard access", async () => {
  vi.spyOn(console, "warn").mockImplementation(() => {});
  Object.defineProperty(navigator, "clipboard", { value: undefined });
  expect(await tryCopyToClipboard("value")).toBe(false);
  expect(mocks.write).not.toHaveBeenCalled();
});

it("keeps password copying silent", async () => {
  mocks.write.mockResolvedValueOnce(undefined);
  expect(await tryCopyToClipboard("password")).toBe(true);
  expect(mocks.emit).not.toHaveBeenCalled();
});

it("preserves silent copying for existing callers", async () => {
  mocks.write.mockResolvedValueOnce(undefined);
  expect(copyToClipboard("password", undefined, false)).toBeUndefined();
  await vi.waitFor(() => expect(mocks.write).toHaveBeenCalledWith("password"));
  expect(mocks.emit).not.toHaveBeenCalled();
});
