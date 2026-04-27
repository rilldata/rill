import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  type V1GetConversationResponse,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get, writable } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

// =============================================================================
// MOCKS
// =============================================================================

class FakeSSEStream {
  public status = writable("closed");
  public url = "";
  public options: unknown;

  private readonly typedHandlers = new Map<
    string,
    Set<(arg?: unknown) => void>
  >();
  private readonly connectionHandlers = new Map<
    string,
    Set<(arg?: unknown) => void>
  >();

  public start = vi.fn((url: string, opts: unknown) => {
    this.url = url;
    this.options = opts;
  });
  public close = vi.fn();
  public pause = vi.fn();
  public resumeIfPaused = vi.fn(async () => {});
  public cleanup = vi.fn();

  public on = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.typedHandlers.has(event))
      this.typedHandlers.set(event, new Set());
    this.typedHandlers.get(event)!.add(listener);
    return () => this.typedHandlers.get(event)!.delete(listener);
  };
  public once = this.on;

  public onConnection = (event: string, listener: (arg?: unknown) => void) => {
    if (!this.connectionHandlers.has(event)) {
      this.connectionHandlers.set(event, new Set());
    }
    this.connectionHandlers.get(event)!.add(listener);
    return () => this.connectionHandlers.get(event)!.delete(listener);
  };
  public onceConnection = this.onConnection;

  /** Simulate a typed payload event arriving from the transport. */
  public deliver(event: string, payload: unknown) {
    this.typedHandlers.get(event)?.forEach((h) => h(payload));
  }

  public fireConnection(event: string, payload?: unknown) {
    this.connectionHandlers.get(event)?.forEach((h) => h(payload));
  }
}

const fakeStreams: FakeSSEStream[] = [];
let latestCreateOptions:
  | {
      subscriber?: {
        onParseError?: (
          err: unknown,
          message: { type?: string; data: string },
        ) => void;
      };
    }
  | undefined;

vi.mock("@rilldata/web-common/runtime-client/sse", () => {
  // Defined inside the factory so vi.mock's hoisting doesn't hit a TDZ when
  // resolving this module graph.
  class FakeSSEHttpError extends Error {
    constructor(
      public readonly status: number,
      public readonly statusText: string,
    ) {
      super(`HTTP ${status}: ${statusText}`);
      this.name = "SSEHttpError";
    }
  }

  return {
    createSSEStream: vi.fn((options: unknown) => {
      latestCreateOptions = options as typeof latestCreateOptions;
      const stream = new FakeSSEStream();
      fakeStreams.push(stream);
      return stream;
    }),
    SSEHttpError: FakeSSEHttpError,
    ConnectionStatus: {
      CONNECTING: "connecting",
      OPEN: "open",
      PAUSED: "paused",
      CLOSED: "closed",
    },
  };
});

import { SSEHttpError as MockedSSEHttpError } from "@rilldata/web-common/runtime-client/sse";

import { Conversation } from "./conversation";

const INSTANCE_ID = "test-instance";

const mockRuntimeClient = {
  host: "http://localhost:9009",
  instanceId: INSTANCE_ID,
  getJwt: () => undefined,
} as unknown as RuntimeClient;

function latestConnection() {
  return fakeStreams[fakeStreams.length - 1];
}

function getCachedData(conversationId: string) {
  return queryClient.getQueryData<V1GetConversationResponse>(
    getRuntimeServiceGetConversationQueryKey(INSTANCE_ID, {
      conversationId,
    }),
  );
}

describe("Conversation streaming", () => {
  beforeEach(() => {
    queryClient.clear();
    fakeStreams.length = 0;
    latestCreateOptions = undefined;
  });

  afterEach(() => {
    queryClient.clear();
  });

  it("routes untagged frames through the 'message' decoder and updates the message cache", async () => {
    queryClient.setQueryData<V1GetConversationResponse>(
      getRuntimeServiceGetConversationQueryKey(INSTANCE_ID, {
        conversationId: "conv-1",
      }),
      { conversation: { id: "conv-1" }, messages: [], isOwner: true },
    );

    const conversation = new Conversation(mockRuntimeClient, "conv-1");
    conversation.draftMessage.set("hi");

    let emittedMessage: unknown;
    conversation.on("message", (msg) => {
      emittedMessage = msg;
    });

    const sendPromise = conversation.sendMessage({}).catch(() => {});
    // Let startStreaming register its listeners.
    await Promise.resolve();
    await Promise.resolve();

    const stream = latestConnection();
    expect(stream).toBeDefined();
    // Untagged frames would route through the "message" decoder — here we
    // exercise the decoder directly by delivering the decoded payload.
    stream.deliver("message", {
      conversationId: "conv-1",
      message: {
        id: "assistant-msg-1",
        role: "assistant",
        contentData: "hello from the server",
      },
    });

    stream.fireConnection("close");
    await sendPromise;

    expect(emittedMessage).toMatchObject({ id: "assistant-msg-1" });

    // Message was added to the cache.
    const cached = getCachedData("conv-1");
    const cachedIds = (cached?.messages ?? []).map((m) => m.id);
    expect(cachedIds).toContain("assistant-msg-1");
  });

  it("surfaces typed 'error' frames through streamError", async () => {
    const conversation = new Conversation(mockRuntimeClient, "conv-1");
    conversation.draftMessage.set("hi");

    const sendPromise = conversation.sendMessage({}).catch(() => {});
    await Promise.resolve();
    await Promise.resolve();

    const stream = latestConnection();
    stream.deliver("error", { code: "internal", error: "boom" });

    stream.fireConnection("close");
    await sendPromise;

    expect(get(conversation.streamError)).toBe("boom");
  });

  it("surfaces transport errors (SSEHttpError) through streamError", async () => {
    const conversation = new Conversation(mockRuntimeClient, "conv-1");
    conversation.draftMessage.set("hi");

    const sendPromise = conversation.sendMessage({}).catch(() => {});
    await Promise.resolve();
    await Promise.resolve();

    const conn = latestConnection();
    conn.fireConnection(
      "error",
      new (MockedSSEHttpError as unknown as new (
        status: number,
        statusText: string,
      ) => Error)(401, "Unauthorized"),
    );
    conn.fireConnection("close");
    await sendPromise;

    const err = get(conversation.streamError);
    expect(err).toMatch(/Authentication failed/);
  });

  it("surfaces decoder parse errors through streamError", async () => {
    const conversation = new Conversation(mockRuntimeClient, "conv-1");
    conversation.draftMessage.set("hi");

    const sendPromise = conversation.sendMessage({}).catch(() => {});
    await Promise.resolve();
    await Promise.resolve();

    latestCreateOptions?.subscriber?.onParseError?.(new Error("invalid json"), {
      type: "message",
      data: "{invalid",
    });

    latestConnection().fireConnection("close");
    await sendPromise;

    expect(get(conversation.streamError)).toBe(
      "Failed to process server response",
    );
  });
});
