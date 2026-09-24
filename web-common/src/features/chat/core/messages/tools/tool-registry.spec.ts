import type { V1Message } from "@rilldata/web-common/runtime-client";
import { afterEach, describe, expect, it, vi } from "vitest";
import { MessageContentType, ToolName } from "../../types";
import { getToolConfig } from "./tool-registry";

afterEach(() => {
  document.body.replaceChildren();
  vi.restoreAllMocks();
});

describe("click_ui tool", () => {
  it("clicks only after the server returns a successful result", () => {
    const button = document.createElement("button");
    button.dataset.aiAction = "dashboard.start-pivot";
    vi.spyOn(button, "getBoundingClientRect").mockReturnValue({
      width: 100,
      height: 30,
    } as DOMRect);
    const onClick = vi.fn();
    button.addEventListener("click", onClick);
    document.body.appendChild(button);

    const callMessage: V1Message = {
      contentData: JSON.stringify({ action_id: "dashboard.start-pivot" }),
    };
    const config = getToolConfig(ToolName.CLICK_UI);

    expect(config.onCall).toBeUndefined();
    config.onResult?.(callMessage, {
      contentType: MessageContentType.ERROR,
    });
    expect(onClick).not.toHaveBeenCalled();

    config.onResult?.(callMessage, {
      contentType: MessageContentType.JSON,
    });
    expect(onClick).toHaveBeenCalledOnce();
  });
});
