import { describe, expect, it } from "vitest";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { transformToBlocks } from "../core/messages/block-transform";
import { MessageContentType, MessageType, ToolName } from "../core/types";
import { createMemoryUpdateBlock } from "./memory-update-block";

function msg(overrides: Partial<V1Message>): V1Message {
  return {
    id: overrides.id ?? Math.random().toString(36).slice(2),
    parentId: "",
    role: "assistant",
    type: MessageType.CALL,
    contentType: MessageContentType.JSON,
    contentData: "{}",
    ...overrides,
  };
}

const appliedResult = JSON.stringify({
  op: "add",
  memory_id: "m1",
  content: "Prefers tables over charts.",
});

const rejectedResult = JSON.stringify({
  op: "noop",
  reason: "already remembered",
});

describe("createMemoryUpdateBlock", () => {
  it("returns null while the result is pending", () => {
    const call = msg({ id: "c1", tool: ToolName.UPDATE_MEMORY });
    expect(createMemoryUpdateBlock(call, undefined)).toBeNull();
  });

  it("returns null when the call failed", () => {
    const call = msg({ id: "c1", tool: ToolName.UPDATE_MEMORY });
    const result = msg({
      parentId: "c1",
      type: MessageType.RESULT,
      tool: ToolName.UPDATE_MEMORY,
      contentType: MessageContentType.ERROR,
      contentData: "boom",
    });
    expect(createMemoryUpdateBlock(call, result)).toBeNull();
  });

  it("returns null when the operation was rejected", () => {
    const call = msg({ id: "c1", tool: ToolName.UPDATE_MEMORY });
    const result = msg({
      parentId: "c1",
      type: MessageType.RESULT,
      tool: ToolName.UPDATE_MEMORY,
      contentData: rejectedResult,
    });
    expect(createMemoryUpdateBlock(call, result)).toBeNull();
  });

  it("returns the applied operation", () => {
    const call = msg({ id: "c1", tool: ToolName.UPDATE_MEMORY });
    const result = msg({
      parentId: "c1",
      type: MessageType.RESULT,
      tool: ToolName.UPDATE_MEMORY,
      contentData: appliedResult,
    });
    const block = createMemoryUpdateBlock(call, result);
    expect(block).not.toBeNull();
    expect(block!.type).toBe("memory-update");
    expect(block!.result.op).toBe("add");
    expect(block!.result.memory_id).toBe("m1");
    expect(block!.result.content).toBe("Prefers tables over charts.");
  });
});

describe("transformToBlocks with update_memory", () => {
  it("renders a memory update block for an applied operation", () => {
    const messages: V1Message[] = [
      msg({
        id: "root",
        role: "user",
        tool: ToolName.ROUTER_AGENT,
        contentData: JSON.stringify({ prompt: "Always show tables" }),
      }),
      msg({ id: "update", parentId: "root", tool: ToolName.UPDATE_MEMORY }),
      msg({
        id: "update-result",
        parentId: "update",
        type: MessageType.RESULT,
        tool: ToolName.UPDATE_MEMORY,
        contentData: appliedResult,
      }),
      msg({
        id: "root-result",
        parentId: "root",
        role: "assistant",
        type: MessageType.RESULT,
        tool: ToolName.ROUTER_AGENT,
        contentData: JSON.stringify({
          response: "Sure",
          agent: ToolName.ANALYST_AGENT,
        }),
      }),
    ];

    const blocks = transformToBlocks(messages, false, false);
    const types = blocks.map((b) => b.type);
    expect(types).toEqual(["text", "memory-update", "text"]);
    expect(types).not.toContain("thinking");
  });

  it("renders no block for a rejected operation", () => {
    const messages: V1Message[] = [
      msg({ id: "update", parentId: "root", tool: ToolName.UPDATE_MEMORY }),
      msg({
        id: "update-result",
        parentId: "update",
        type: MessageType.RESULT,
        tool: ToolName.UPDATE_MEMORY,
        contentData: rejectedResult,
      }),
    ];
    expect(transformToBlocks(messages, false, false)).toEqual([]);
  });
});
