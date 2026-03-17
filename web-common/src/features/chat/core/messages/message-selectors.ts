// Utilities to select messages similar to backend utils.

import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageType } from "@rilldata/web-common/features/chat/core/types.ts";

export type MessageSelectorPredicate = (messages: V1Message) => boolean;
export const MessageSelectors = {
  ById: (id: string) => (m: V1Message) => m.id === id,
  ByType: (type: MessageType) => (m: V1Message) => m.type === type,
  ByToolName: (toolName: string) => (m: V1Message) => m.tool === toolName,
};

export function getMessage(
  messages: V1Message[],
  predicates: MessageSelectorPredicate[],
) {
  for (const message of messages) {
    if (predicates.every((predicate) => predicate(message))) return message;
  }
  return null;
}

export function getLastMessage(
  messages: V1Message[],
  predicates: MessageSelectorPredicate[],
) {
  for (let i = messages.length - 1; i >= 0; i--) {
    const message = messages[i];
    if (predicates.every((predicate) => predicate(message))) return message;
  }
  return null;
}

export function getMessages(
  messages: V1Message[],
  predicates: MessageSelectorPredicate[],
) {
  return messages.filter((message) =>
    predicates.every((predicate) => predicate(message)),
  );
}

export function getNearestUserMessage(
  messages: V1Message[],
  fromMessage: V1Message,
) {
  const index = messages.indexOf(fromMessage);
  if (index === -1) return null;
  for (let i = index - 1; i >= 0; i--) {
    const message = messages[i];
    if (message.role === "user") return message;
  }
  return null;
}
