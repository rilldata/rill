/**
 * Queries and mutations for the current user's AI memories.
 *
 * Memories are per user and per project (instance). They are read and written through the runtime,
 * so the same store works in Rill Developer and Rill Cloud.
 */
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceListAIMemoriesQueryKey,
  getRuntimeServiceListAIMemoriesQueryOptions,
  runtimeServiceCreateAIMemory,
  runtimeServiceDeleteAIMemory,
  runtimeServiceUpdateAIMemory,
  runtimeServiceUpdateAIMemorySettings,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery } from "@tanstack/svelte-query";

/** Memory categories, mirroring runtime/ai/memory.go */
export const MemoryCategory = {
  PREFERENCE: "preference",
  DEFINITION: "definition",
  CONTEXT: "context",
  FEEDBACK: "feedback",
} as const;
export type MemoryCategory =
  (typeof MemoryCategory)[keyof typeof MemoryCategory];

export const MEMORY_CATEGORIES: MemoryCategory[] = [
  MemoryCategory.PREFERENCE,
  MemoryCategory.DEFINITION,
  MemoryCategory.CONTEXT,
  MemoryCategory.FEEDBACK,
];

/** Limits mirroring runtime/ai/memory.go */
export const MAX_MEMORIES = 50;
export const MAX_MEMORY_CONTENT_CHARS = 300;

/**
 * The user's memories plus whether memory is available to them (feature flag on, logged-in user)
 * and whether they have it turned on.
 */
export function useAIMemories(client: RuntimeClient) {
  return createQuery(
    getRuntimeServiceListAIMemoriesQueryOptions(
      client,
      {},
      { query: { enabled: !!client.instanceId } },
    ),
    queryClient,
  );
}

/** Refetch the memory list after a change, whether made here or by the AI during a chat turn. */
export function invalidateAIMemories(instanceId: string) {
  return queryClient.invalidateQueries({
    queryKey: getRuntimeServiceListAIMemoriesQueryKey(instanceId),
  });
}

export async function createAIMemory(
  client: RuntimeClient,
  category: string,
  content: string,
) {
  const res = await runtimeServiceCreateAIMemory(client, {
    category,
    content,
  });
  await invalidateAIMemories(client.instanceId);
  return res.memory;
}

export async function updateAIMemory(
  client: RuntimeClient,
  memoryId: string,
  category: string,
  content: string,
) {
  const res = await runtimeServiceUpdateAIMemory(client, {
    memoryId,
    category,
    content,
  });
  await invalidateAIMemories(client.instanceId);
  return res.memory;
}

export async function deleteAIMemory(client: RuntimeClient, memoryId: string) {
  await runtimeServiceDeleteAIMemory(client, { memoryId });
  await invalidateAIMemories(client.instanceId);
}

export async function setAIMemoryEnabled(
  client: RuntimeClient,
  enabled: boolean,
) {
  await runtimeServiceUpdateAIMemorySettings(client, { enabled });
  await invalidateAIMemories(client.instanceId);
}
