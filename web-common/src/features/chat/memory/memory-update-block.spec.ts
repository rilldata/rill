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

const appliedOps = JSON.stringify({
  ops: [
    {
      op: "add",
      memory_id: "m1",
      category: "preference",
      content: "Prefers tables over charts.",
    },
    { op: "noop", reason: "already remembered" },
  ],
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

  it("returns null when nothing changed", () => {
    const call = msg({ id: "c1", tool: ToolName.EXTRACT_MEMORIES });
    const result = msg({
      parentId: "c1",
      type: MessageType.RESULT,
      tool: ToolName.EXTRACT_MEMORIES,
      contentData: JSON.stringify({ ops: [{ op: "noop", reason: "x" }] }),
    });
    expect(createMemoryUpdateBlock(call, result)).toBeNull();
  });

  it("keeps only applied ops", () => {
    const call = msg({ id: "c1", tool: ToolName.UPDATE_MEMORY });
    const result = msg({
      parentId: "c1",
      type: MessageType.RESULT,
      tool: ToolName.UPDATE_MEMORY,
      contentData: appliedOps,
    });
    const block = createMemoryUpdateBlock(call, result);
    expect(block).not.toBeNull();
    expect(block!.type).toBe("memory-update");
    expect(block!.ops).toHaveLength(1);
    expect(block!.ops[0].memory_id).toBe("m1");
  });
});

describe("transformToBlocks with memory tools", () => {
  it("renders a memory update block after the assistant answer and hides extraction progress", () => {
    const messages: V1Message[] = [
      msg({
        id: "root",
        role: "user",
        tool: ToolName.ROUTER_AGENT,
        contentData: JSON.stringify({ prompt: "Always show tables" }),
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
      msg({ id: "extract", parentId: "root", tool: ToolName.EXTRACT_MEMORIES }),
      msg({
        id: "extract-progress",
        parentId: "extract",
        type: MessageType.PROGRESS,
        contentType: MessageContentType.TEXT,
        contentData: '{"ops":[]}',
      }),
      msg({
        id: "extract-result",
        parentId: "extract",
        type: MessageType.RESULT,
        tool: ToolName.EXTRACT_MEMORIES,
        contentData: appliedOps,
      }),
    ];

    const blocks = transformToBlocks(messages, false, false);
    const types = blocks.map((b) => b.type);
    expect(types).toEqual(["text", "text", "memory-update"]);
    expect(types).not.toContain("thinking");
  });

  it("renders no block for an extraction that learned nothing", () => {
    const messages: V1Message[] = [
      msg({ id: "extract", parentId: "root", tool: ToolName.EXTRACT_MEMORIES }),
      msg({
        id: "extract-result",
        parentId: "extract",
        type: MessageType.RESULT,
        tool: ToolName.EXTRACT_MEMORIES,
        contentData: JSON.stringify({ ops: [] }),
      }),
    ];
    expect(transformToBlocks(messages, false, false)).toEqual([]);
  });
});
