import { Timestamp, type PartialMessage } from "@bufbuild/protobuf";
import type {
  AIFeedback,
  Message,
} from "@rilldata/web-common/proto/gen/rill/runtime/v1/api_pb";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  runtimeServiceGetAIFeedback,
  runtimeServiceUpdateAIFeedbackStatus,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  localServiceGetProjectAIFeedback,
  localServiceResolveProjectAIFeedback,
} from "@rilldata/web-common/runtime-client/local-service";

// FeedbackOrigin identifies the path an item was loaded through, which is also the path
// its detail/resolve calls must take: "local-service" items live in a cloud deployment
// reached via the web-local LocalService bridge; "runtime" items are read directly from
// a runtime (the local instance, an edit session, or a production deployment) using the
// RuntimeClient they were listed with.
export type FeedbackOrigin = "local-service" | "runtime";

export type FeedbackStatus = "open" | "addressed" | "dismissed";

export const feedbackStatusOptions: FeedbackStatus[] = [
  "open",
  "addressed",
  "dismissed",
];

export const feedbackStatusLabels: Record<FeedbackStatus, () => string> = {
  open: m.feedback_status_open,
  addressed: m.feedback_status_addressed,
  dismissed: m.feedback_status_dismissed,
};

// InboxItem is a normalized feedback item. Cloud items arrive as protobuf class instances
// (via the LocalService connect client), local items as JSON-shaped plain objects
// (via the generated runtime hooks); this type unifies the two.
export interface InboxItem {
  id: string;
  origin: FeedbackOrigin;
  conversationId: string;
  targetMessageId: string;
  kind: string;
  sentiment: string;
  categories: string[];
  comment: string;
  predictedAttribution: string;
  suggestedAction: string;
  status: string;
  ownerEmail: string;
  createdOn: Date | undefined;
  conversationTitle: string;
}

export function normalizeFeedback(
  f: AIFeedback | PartialMessage<AIFeedback>,
  origin: FeedbackOrigin,
): InboxItem {
  return {
    id: f.id ?? "",
    origin,
    conversationId: f.conversationId ?? "",
    targetMessageId: f.targetMessageId ?? "",
    kind: f.kind ?? "",
    sentiment: f.sentiment ?? "",
    categories: (f.categories ?? []) as string[],
    comment: f.comment ?? "",
    predictedAttribution: f.predictedAttribution ?? "",
    suggestedAction: f.suggestedAction ?? "",
    status: f.status ?? "",
    ownerEmail: f.ownerEmail ?? "",
    createdOn: toDate(f.createdOn),
    conversationTitle: f.conversationTitle ?? "",
  };
}

// toDate converts a timestamp that may be a protobuf Timestamp (connect client)
// or an ISO string (JSON-shaped runtime hooks).
function toDate(v: unknown): Date | undefined {
  if (!v) return undefined;
  if (typeof v === "string") return new Date(v);
  if (v instanceof Timestamp) return v.toDate();
  return undefined;
}

export interface FeedbackDetailData {
  transcriptAvailable: boolean;
  messages: V1Message[];
}

// fetchFeedbackDetail loads the conversation transcript for a feedback item
// from the store it lives in (cloud deployment or the local runtime).
export async function fetchFeedbackDetail(
  client: RuntimeClient,
  item: InboxItem,
): Promise<FeedbackDetailData> {
  if (item.origin === "local-service") {
    const res = await localServiceGetProjectAIFeedback(item.id);
    return {
      transcriptAvailable: res.conversation !== undefined,
      messages: res.messages.map(protoMessageToV1),
    };
  }
  const res = await runtimeServiceGetAIFeedback(client, {
    feedbackId: item.id,
  });
  return {
    transcriptAvailable: res.conversation !== undefined,
    messages: (res.messages ?? []) as V1Message[],
  };
}

// resolveFeedback updates the item's status in the store it lives in.
export async function resolveFeedback(
  client: RuntimeClient,
  item: InboxItem,
  status: FeedbackStatus,
): Promise<void> {
  if (item.origin === "local-service") {
    await localServiceResolveProjectAIFeedback({
      feedbackId: item.id,
      status,
    });
    return;
  }
  await runtimeServiceUpdateAIFeedbackStatus(client, {
    feedbackId: item.id,
    status,
  });
}

// invalidateFeedbackLists refetches every feedback list, regardless of the store it
// reads from: the LocalService bridge query and all per-runtime list queries.
export async function invalidateFeedbackLists(): Promise<void> {
  await Promise.all([
    queryClient.invalidateQueries({ queryKey: ["/v1/local/ai-feedback"] }),
    queryClient.invalidateQueries({
      queryKey: ["RuntimeService", "listAIFeedback"],
    }),
  ]);
}

// protoMessageToV1 converts a protobuf Message class instance to the JSON shape
// that the chat rendering pipeline (transformToBlocks) expects.
function protoMessageToV1(m: Message): V1Message {
  return m.toJson({ emitDefaultValues: true }) as unknown as V1Message;
}
