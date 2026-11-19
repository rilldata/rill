import { afterEach, describe, expect, it, vi } from "vitest";
import { clickUIAction, getAvailableUIActions } from "./ui-actions";

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

  it("limits labels to 256 code points without splitting an emoji", () => {
    // Matches the server, which truncates labels by character count (runtime/server/chat.go).
    const nonASCII = "é".repeat(256);
    addAction("canvas.share", nonASCII);
    addAction("canvas.edit", "x".repeat(300));
    // 256 code points but 257 UTF-16 code units: the emoji at the end must survive whole.
    addAction("canvas.emoji", "x".repeat(255) + "😀");
    // The 257th code point is an emoji: it must be dropped whole, never cut into a lone surrogate.
    addAction("canvas.emoji-cut", "x".repeat(255) + "😀😀");

    const actions = getAvailableUIActions();
    expect(actions).toEqual([
      { id: "canvas.share", label: nonASCII },
      { id: "canvas.edit", label: "x".repeat(256) },
      { id: "canvas.emoji", label: "x".repeat(255) + "😀" },
      { id: "canvas.emoji-cut", label: "x".repeat(255) + "😀" },
    ]);
    for (const action of actions) {
      expect(action.label.isWellFormed()).toBe(true);
    }
  });

  it("ignores actions inside a data-ai-ignore subtree", () => {
    const ignored = document.createElement("div");
    ignored.dataset.aiIgnore = "";
    const button = addAction("chat.close", "Close chat");
    ignored.appendChild(button);
    document.body.appendChild(ignored);

    expect(getAvailableUIActions()).toEqual([]);
  });

  it("clicks exactly one currently available action", () => {
    const button = addAction("dashboard.search", "Search");
    const onClick = vi.fn();
    button.addEventListener("click", onClick);

    expect(clickUIAction("dashboard.search")).toBe(true);
    expect(onClick).toHaveBeenCalledOnce();
    expect(clickUIAction("missing")).toBe(false);
  });
});
