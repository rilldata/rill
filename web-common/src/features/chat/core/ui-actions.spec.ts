import { afterEach, describe, expect, it, vi } from "vitest";
import {
  getAvailableUIActions,
  registerUIActionHandler,
  runUIAction,
} from "./ui-actions";

function addAction(
  id: string,
  label: string,
  options: { disabled?: boolean; hidden?: boolean } = {},
): HTMLButtonElement {
  const button = document.createElement("button");
  button.dataset.aiAction = id;
  button.setAttribute("aria-label", label);
  button.disabled = options.disabled ?? false;
  button.hidden = options.hidden ?? false;
  vi.spyOn(button, "getBoundingClientRect").mockReturnValue({
    width: 100,
    height: 30,
  } as DOMRect);
  document.body.appendChild(button);
  return button;
}

afterEach(() => {
  document.body.replaceChildren();
  vi.restoreAllMocks();
});

describe("UI actions", () => {
  it("discovers only visible, enabled, uniquely-addressable actions", () => {
    addAction("dashboard.search", "Search");
    addAction("dashboard.disabled", "Disabled", { disabled: true });
    addAction("dashboard.hidden", "Hidden", { hidden: true });
    addAction("dashboard.duplicate", "First duplicate");
    addAction("dashboard.duplicate", "Second duplicate");

    expect(getAvailableUIActions()).toEqual([
      { id: "dashboard.search", label: "Search" },
    ]);
  });

  it("ignores actions inside a data-ai-ignore subtree", () => {
    const ignored = document.createElement("div");
    ignored.dataset.aiIgnore = "";
    const button = addAction("chat.close", "Close chat");
    ignored.appendChild(button);
    document.body.appendChild(ignored);

    expect(getAvailableUIActions()).toEqual([]);
  });

  it("clicks exactly one currently available action", async () => {
    const button = addAction("dashboard.search", "Search");
    const onClick = vi.fn();
    button.addEventListener("click", onClick);

    await runUIAction("dashboard.search");
    expect(onClick).toHaveBeenCalledOnce();
    await expect(runUIAction("missing")).rejects.toThrow(
      "no longer uniquely available",
    );
  });

  it("runs a registered handler instead of clicking", async () => {
    const button = addAction("component.eject", "Eject");
    const onClick = vi.fn();
    button.addEventListener("click", onClick);
    const handler = vi.fn().mockResolvedValue(undefined);
    const unregister = registerUIActionHandler("component.eject", handler);

    await runUIAction("component.eject");
    expect(handler).toHaveBeenCalledOnce();
    expect(onClick).not.toHaveBeenCalled();

    unregister();
    await runUIAction("component.eject");
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("propagates a registered handler's failure", async () => {
    addAction("component.eject", "Eject");
    const unregister = registerUIActionHandler("component.eject", () =>
      Promise.reject(new Error("compile failed")),
    );

    await expect(runUIAction("component.eject")).rejects.toThrow(
      "compile failed",
    );
    unregister();
  });

  it("does not run a handler whose element is unavailable", async () => {
    addAction("component.eject", "Eject", { disabled: true });
    const handler = vi.fn().mockResolvedValue(undefined);
    const unregister = registerUIActionHandler("component.eject", handler);

    await expect(runUIAction("component.eject")).rejects.toThrow();
    expect(handler).not.toHaveBeenCalled();
    unregister();
  });
});
