// Utilities to select messages similar to backend utils.

import type { V1Message } from "@rilldata/web-common/runtime-client";

export type MessageSelectorPredicate = (messages: V1Message) => boolean;
export const MessageSelectors = {
  ById: (id: string) => (m: V1Message) => m.id === id,
  ByType: (type: string) => (m: V1Message) => m.type === type,
  ByToolName: (toolName: string) => (m: V1Message) => m.tool === toolName,
  ByRoleName: (roleName: string) => (m: V1Message) => m.role === roleName,
};

export function getMessage(
  messages: V1Message[],
  predicates: MessageSelectorPredicate[],
  fromIndex: number = 0,
) {
  for (let i = fromIndex; i < messages.length; i++) {
    const message = messages[i];
    if (predicates.every((predicate) => predicate(message))) return message;
  }
  return null;
}

export function getLastMessage(
  messages: V1Message[],
  predicates: MessageSelectorPredicate[],
  fromIndex: number = messages.length - 1,
) {
  for (let i = fromIndex; i >= 0; i--) {
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
