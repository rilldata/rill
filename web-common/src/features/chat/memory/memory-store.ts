/**
 * Queries and mutations for the current user's AI memories.
 *
 * Memories are per user and per project (instance). They are read and written through the runtime,
 * so the same store works in Rill Developer and Rill Cloud.
 */
import { featureFlags } from "@rilldata/web-common/features/feature-flags";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import {
  getRuntimeServiceGetAIMemorySettingsQueryOptions,
  getRuntimeServiceListAIMemoriesQueryOptions,
  runtimeServiceCreateAIMemory,
  runtimeServiceDeleteAIMemory,
  runtimeServiceDeleteAllAIMemories,
  runtimeServiceUpdateAIMemory,
  runtimeServiceUpdateAIMemorySettings,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { createQuery } from "@tanstack/svelte-query";
import { derived, type Readable } from "svelte/store";

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

/** Memory statuses, mirroring runtime/drivers/catalog.go */
export const MemoryStatus = {
  ACTIVE: "active",
  DELETED: "deleted",
} as const;

/** Maximum length of a memory, mirroring runtime/ai/memory.go */
export const MAX_MEMORY_CONTENT_CHARS = 300;

export function useAIMemories(client: RuntimeClient, includeDeleted = false) {
  return createQuery(
    getRuntimeServiceListAIMemoriesQueryOptions(
      client,
      { includeDeleted },
      { query: { enabled: !!client.instanceId } },
    ),
    queryClient,
  );
}

export function useAIMemorySettings(client: RuntimeClient) {
  return createQuery(
    getRuntimeServiceGetAIMemorySettingsQueryOptions(
      client,
      {},
      { query: { enabled: !!client.instanceId } },
    ),
    queryClient,
  );
}

/**
 * True when memory is available to the current user: the feature flag is on and the runtime reports it enabled
 * (it is off for anonymous users and when an admin disabled it for the project).
 */
export function useMemoryEnabled(client: RuntimeClient): Readable<boolean> {
  const settings = useAIMemorySettings(client);
  return derived(
    [featureFlags.chatMemory, settings],
    ([$flag, $settings]) => $flag && !!$settings.data?.enabled,
  );
}

/** Refetch memory lists and settings, for one instance or (when omitted) for all. */
export function invalidateAIMemories(instanceId?: string) {
  return queryClient.invalidateQueries({
    predicate: (query) => {
      const key = query.queryKey;
      return (
        Array.isArray(key) &&
        key[0] === "RuntimeService" &&
        (key[1] === "listAIMemories" || key[1] === "getAIMemorySettings") &&
        (instanceId === undefined || key[2] === instanceId)
      );
    },
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
  changes: { category?: string; content?: string; status?: string },
) {
  const res = await runtimeServiceUpdateAIMemory(client, {
    memoryId,
    ...changes,
  });
  await invalidateAIMemories(client.instanceId);
  return res.memory;
}

export async function deleteAIMemory(client: RuntimeClient, memoryId: string) {
  await runtimeServiceDeleteAIMemory(client, { memoryId });
  await invalidateAIMemories(client.instanceId);
}

export async function deleteAllAIMemories(client: RuntimeClient) {
  await runtimeServiceDeleteAllAIMemories(client, {});
  await invalidateAIMemories(client.instanceId);
}

export async function setAIMemoryPaused(
  client: RuntimeClient,
  paused: boolean,
) {
  await runtimeServiceUpdateAIMemorySettings(client, { paused });
  await invalidateAIMemories(client.instanceId);
}
