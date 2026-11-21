import type { V1Message } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { runtimeServiceCompleteUIAction } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { afterEach, describe, expect, it, vi } from "vitest";
import { ToolName } from "../../types";
import { getToolConfig } from "./tool-registry";

vi.mock("@rilldata/web-common/runtime-client/v2/gen/runtime-service", () => ({
  runtimeServiceCompleteUIAction: vi.fn().mockResolvedValue({}),
}));

afterEach(() => {
  document.body.replaceChildren();
  vi.clearAllMocks();
});

const client = {} as RuntimeClient;

function callMessage(actionId: string): V1Message {
  return {
    id: "call-1",
    contentData: JSON.stringify({ action_id: actionId }),
  };
}

describe("click_ui tool", () => {
  it("runs the action on the call and reports success", async () => {
    const button = document.createElement("button");
    button.dataset.aiAction = "dashboard.start-pivot";
    vi.spyOn(button, "getBoundingClientRect").mockReturnValue({
      width: 100,
      height: 30,
    } as DOMRect);
    const onClick = vi.fn();
    button.addEventListener("click", onClick);
    document.body.appendChild(button);

    const config = getToolConfig(ToolName.CLICK_UI);
    expect(config.onResult).toBeUndefined();
    config.onCall?.(callMessage("dashboard.start-pivot"), client, "conv-1");

    await vi.waitFor(() =>
      expect(runtimeServiceCompleteUIAction).toHaveBeenCalledWith(client, {
        conversationId: "conv-1",
        callMessageId: "call-1",
        error: "",
      }),
    );
    expect(onClick).toHaveBeenCalledOnce();
  });

  it("reports an error when the action is unavailable", async () => {
    getToolConfig(ToolName.CLICK_UI).onCall?.(
      callMessage("dashboard.missing"),
      client,
      "conv-1",
    );

    await vi.waitFor(() =>
      expect(runtimeServiceCompleteUIAction).toHaveBeenCalledWith(
        client,
        expect.objectContaining({
          callMessageId: "call-1",
          error: expect.stringContaining("no longer uniquely available"),
        }),
      ),
    );
  });
});
