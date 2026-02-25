import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetConversationQueryKey,
  type V1GetConversationResponse,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { get } from "svelte/store";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { Conversation } from "./conversation";
import { NEW_CONVERSATION_ID } from "./utils";

// =============================================================================
// MOCKS
// =============================================================================

vi.mock("@rilldata/web-common/runtime-client", async (importOriginal) => {
  const original =
    await importOriginal<
      typeof import("@rilldata/web-common/runtime-client")
    >();
  return {
    ...original,
    runtimeServiceForkConversation: vi.fn(),
  };
});

import { runtimeServiceForkConversation } from "@rilldata/web-common/runtime-client";

// =============================================================================
// TEST CONSTANTS
// =============================================================================

const INSTANCE_ID = "test-instance";

const mockRuntimeClient = {
  host: "http://localhost:9009",
  instanceId: INSTANCE_ID,
  getJwt: () => undefined,
} as unknown as RuntimeClient;
const ORIGINAL_CONVERSATION_ID = "original-conv-123";
const FORKED_CONVERSATION_ID = "forked-conv-456";

// =============================================================================
// HELPERS
// =============================================================================

function getCacheKey(conversationId: string) {
  return getRuntimeServiceGetConversationQueryKey(INSTANCE_ID, conversationId);
}

function getCachedData(conversationId: string) {
  return queryClient.getQueryData<V1GetConversationResponse>(
    getCacheKey(conversationId),
  );
}

function seedCache(
  conversationId: string,
  options: {
    isOwner: boolean;
    messages?: Partial<V1Message>[];
    title?: string;
    createdOn?: string;
  },
) {
  queryClient.setQueryData<V1GetConversationResponse>(
    getCacheKey(conversationId),
    {
      conversation: {
        id: conversationId,
        title: options.title,
        createdOn: options.createdOn,
      },
      messages: options.messages as V1Message[],
      isOwner: options.isOwner,
    },
  );
}

function mockForkSuccess(forkedId: string = FORKED_CONVERSATION_ID) {
  vi.mocked(runtimeServiceForkConversation).mockResolvedValue({
    conversationId: forkedId,
  });
}

function mockForkFailure(error: Error = new Error("Fork failed")) {
  vi.mocked(runtimeServiceForkConversation).mockRejectedValue(error);
}

function mockForkEmptyResponse() {
  vi.mocked(runtimeServiceForkConversation).mockResolvedValue({});
}

function createConversation(conversationId: string = ORIGINAL_CONVERSATION_ID) {
  return new Conversation(mockRuntimeClient, conversationId);
}

async function sendMessageAndIgnoreStreamError(
  conversation: Conversation,
  message: string,
) {
  conversation.draftMessage.set(message);
  await conversation.sendMessage({}).catch(() => {});
}

// =============================================================================
// TESTS
// =============================================================================

describe("Conversation", () => {
  beforeEach(() => {
    queryClient.clear();
    vi.clearAllMocks();
  });

  afterEach(() => {
    queryClient.clear();
  });

  describe("forkConversation", () => {
    it("forks and copies messages when non-owner sends a message", async () => {
      // Arrange
      seedCache(ORIGINAL_CONVERSATION_ID, {
        isOwner: false,
        title: "Shared conversation",
        createdOn: "2024-01-01T00:00:00Z",
        messages: [
          { id: "msg-1", role: "user", contentData: "Hello" },
          { id: "msg-2", role: "assistant", contentData: "Hi there!" },
        ],
      });
      mockForkSuccess();

      const conversation = createConversation();
      let forkedId: string | null = null;
      conversation.on("conversation-forked", (id) => (forkedId = id));

      // Act
      conversation.draftMessage.set("My follow-up question");
      const sendPromise = conversation.sendMessage({});
      await vi.waitFor(() => expect(forkedId).toBe(FORKED_CONVERSATION_ID));

      // Assert: fork API called correctly
      expect(runtimeServiceForkConversation).toHaveBeenCalledWith(
        INSTANCE_ID,
        ORIGINAL_CONVERSATION_ID,
        {},
      );

      // Assert: cache updated with forked conversation
      const forkedData = getCachedData(FORKED_CONVERSATION_ID);
      expect(forkedData?.conversation?.id).toBe(FORKED_CONVERSATION_ID);
      expect(forkedData?.conversation?.title).toBe("Shared conversation");
      expect(forkedData?.conversation?.createdOn).toBe("2024-01-01T00:00:00Z");
      expect(forkedData?.isOwner).toBe(true);

      // Assert: messages copied + optimistic message added
      expect(forkedData?.messages).toHaveLength(3);
      expect(forkedData?.messages?.[0]?.contentData).toBe("Hello");
      expect(forkedData?.messages?.[1]?.contentData).toBe("Hi there!");
      expect(forkedData?.messages?.[2]?.role).toBe("user");

      // Cleanup
      conversation.cleanup();
      await sendPromise.catch(() => {});
    });

    it("does NOT fork when owner sends a message", async () => {
      // Arrange
      seedCache(ORIGINAL_CONVERSATION_ID, { isOwner: true, messages: [] });
      const conversation = createConversation();

      // Act
      await sendMessageAndIgnoreStreamError(conversation, "My message");

      // Assert
      expect(runtimeServiceForkConversation).not.toHaveBeenCalled();

      conversation.cleanup();
    });

    it("does NOT fork for new conversations", async () => {
      // Arrange
      const conversation = createConversation(NEW_CONVERSATION_ID);

      // Act
      await sendMessageAndIgnoreStreamError(conversation, "First message");

      // Assert
      expect(runtimeServiceForkConversation).not.toHaveBeenCalled();

      conversation.cleanup();
    });

    it("sets error and stops streaming if fork API fails", async () => {
      // Arrange
      seedCache(ORIGINAL_CONVERSATION_ID, { isOwner: false, messages: [] });
      mockForkFailure();
      const conversation = createConversation();

      // Act
      conversation.draftMessage.set("My message");
      await conversation.sendMessage({});

      // Assert
      expect(get(conversation.streamError)).toContain(
        "Failed to create your copy",
      );
      expect(get(conversation.isStreaming)).toBe(false);

      conversation.cleanup();
    });

    it("sets error if fork response is missing conversation ID", async () => {
      // Arrange
      seedCache(ORIGINAL_CONVERSATION_ID, { isOwner: false, messages: [] });
      mockForkEmptyResponse();
      const conversation = createConversation();

      // Act
      conversation.draftMessage.set("My message");
      await conversation.sendMessage({});

      // Assert
      expect(get(conversation.streamError)).toContain(
        "Failed to create your copy",
      );
      expect(get(conversation.isStreaming)).toBe(false);

      conversation.cleanup();
    });

    it("does NOT fork when cache is empty (optimistic ownership)", async () => {
      // Arrange: no cache data seeded - ownership defaults to true
      const conversation = createConversation();

      // Act
      await sendMessageAndIgnoreStreamError(conversation, "Test");

      // Assert
      expect(runtimeServiceForkConversation).not.toHaveBeenCalled();

      conversation.cleanup();
    });
  });
});
