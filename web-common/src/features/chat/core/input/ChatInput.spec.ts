import { fireEvent, render } from "@testing-library/svelte";
import type { Editor } from "@tiptap/core";
import { tick } from "svelte";
import { readable, writable } from "svelte/store";
import { beforeAll, describe, expect, it, vi } from "vitest";
import type { Conversation } from "../conversation";
import type { ConversationManager } from "../conversation-manager";
import type { ChatConfig } from "../types";
import ChatInput from "./ChatInput.svelte";

// A chat whose agent has no skills, so the input does not need a runtime client for the "/" picker.
const config: ChatConfig = {
  agent: "test_agent",
  additionalContextStoreGetter: () => readable({}),
  emptyChatLabel: "",
  placeholder: "",
  minChatHeight: "",
};

// jsdom has no layout. ProseMirror reads these to scroll the selection into view when the editor is focused.
beforeAll(() => {
  Range.prototype.getClientRects = () => [] as unknown as DOMRectList;
  Range.prototype.getBoundingClientRect = () => new DOMRect();
});

/**
 * Renders the input for a conversation whose response the test controls:
 * `sendMessage` resolves once the test finishes the response.
 */
function renderInput() {
  let finishResponse = () => {};
  const sendMessage = vi.fn<Conversation["sendMessage"]>(
    () =>
      new Promise<void>((resolve) => {
        finishResponse = resolve;
      }),
  );
  const conversation = {
    draftMessage: writable(""),
    isStreaming: writable(false),
    getConversationQuery: () => readable({ isLoading: false }),
    sendMessage,
    cancelStream: vi.fn(),
  } as unknown as Conversation;
  const conversationManager = {
    getCurrentConversation: () => readable(conversation),
  } as unknown as ConversationManager;

  const { container, unmount } = render(ChatInput, {
    props: { conversationManager, config },
  });
  // TipTap exposes the editor on its root element.
  const tiptap = container.querySelector(".tiptap") as HTMLElement & {
    editor: Editor;
  };

  return {
    tiptap,
    editor: tiptap.editor,
    unmount,
    // Types a prompt and presses Enter, then returns the hooks the input handed to the conversation.
    send: async (prompt: string) => {
      tiptap.editor.commands.insertContent(prompt);
      await tick();
      await fireEvent.keyDown(tiptap, { key: "Enter" });
      expect(sendMessage).toHaveBeenCalledOnce();
      return {
        onStreamStart: sendMessage.mock.calls[0][1]!.onStreamStart!,
        finishResponse: () => finishResponse(),
      };
    },
  };
}

describe("ChatInput", () => {
  it("clears the input when the response starts streaming and focuses it when the response ends", async () => {
    const { tiptap, send } = renderInput();
    const { onStreamStart, finishResponse } = await send("hello");
    expect(tiptap.textContent).toBe("hello");

    onStreamStart();
    expect(tiptap.textContent).toBe("");

    finishResponse();
    await vi.waitFor(() => expect(document.activeElement).toBe(tiptap));
  });

  it("leaves the editor alone when the input is destroyed while the response streams", async () => {
    const { editor, unmount, send } = renderInput();
    const { onStreamStart, finishResponse } = await send("hello");

    // Leaving the page destroys the input, and with it the editor, while the message is still in flight.
    unmount();
    expect(editor.isDestroyed).toBe(true);
    // A destroyed editor has no command manager, so running any command on it throws.
    const commands = vi.spyOn(editor, "commands", "get");

    onStreamStart();
    finishResponse();
    // Let the pending send run to completion.
    await new Promise((resolve) => setTimeout(resolve));
    expect(commands).not.toHaveBeenCalled();
  });
});
