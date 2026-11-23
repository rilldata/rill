import type { V1Message } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import { MessageContentType, MessageType, ToolName } from "../types";
import { transformToBlocks } from "./block-transform";

function routerMessage(
  id: string,
  role: "user" | "assistant",
  text: string,
): V1Message {
  return {
    id,
    role,
    type: role === "user" ? MessageType.CALL : MessageType.RESULT,
    tool: ToolName.ROUTER_AGENT,
    contentType: MessageContentType.TEXT,
    contentData: text,
  };
}

describe("transformToBlocks", () => {
  it("renders a message that is in the list twice only once", () => {
    const messages = [
      routerMessage("m1", "user", "hi"),
      routerMessage("m2", "assistant", "hello"),
      routerMessage("m2", "assistant", "hello"),
      routerMessage("m3", "user", "bye"),
    ];

    const blocks = transformToBlocks(messages, false, false);

    const ids = blocks.map((block) => block.id);
    expect(ids).toEqual(["m1", "m2", "m3"]);
  });
});
