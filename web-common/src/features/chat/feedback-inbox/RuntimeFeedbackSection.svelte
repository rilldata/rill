<!--
  One inbox section backed directly by a runtime's ai_feedback store.
  The client decides which store: the local instance in web-local, or the
  edit session / production deployment in cloud editing.
-->
<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { createRuntimeServiceListAIFeedback } from "@rilldata/web-common/runtime-client";
  import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Inbox } from "lucide-svelte";
  import {
    feedbackStatusLabels,
    invalidateFeedbackLists,
    normalizeFeedback,
    resolveFeedback,
    type FeedbackStatus,
    type InboxItem,
  } from "./inbox-data";
  import FeedbackInboxItem from "./FeedbackInboxItem.svelte";

  export let client: RuntimeClient | null;
  export let title: string;
  export let status: FeedbackStatus;
  export let canSuggestFix: boolean;
  // Category filter shared across the inbox's sections; clicking a category
  // tag on an item toggles it.
  export let categoryFilter: string | null = null;
  export let onCategoryClick: ((category: string) => void) | undefined =
    undefined;
  // Shown instead of the list when no client is available (e.g. no production
  // deployment yet). Empty while the availability is still being determined.
  export let unavailableMessage = "";

  let expandedId: string | null = null;
  let resolveError: string | null = null;

  $: listQuery = client
    ? createRuntimeServiceListAIFeedback(client, { status })
    : null;
  $: items = ($listQuery?.data?.feedback ?? [])
    .map((f) => normalizeFeedback(f, "runtime"))
    .filter((i) => !categoryFilter || i.categories.includes(categoryFilter));
  $: listError = $listQuery?.error;

  async function handleResolve(item: InboxItem, newStatus: FeedbackStatus) {
    if (!client) return;
    resolveError = null;
    try {
      await resolveFeedback(client, item, newStatus);
    } catch (e) {
      resolveError = e instanceof Error ? e.message : String(e);
      return;
    }
    await invalidateFeedbackLists();
  }
</script>

<section class="flex flex-col gap-y-2">
  <h2 class="text-sm font-semibold text-fg-secondary">{title}</h2>
  {#if resolveError}
    <p class="text-sm text-red-600">
      {m.feedback_inbox_error({ message: resolveError })}
    </p>
  {/if}
  {#if !client}
    {#if unavailableMessage}
      <div class="py-6">
        <ResourceListEmptyState icon={Inbox} message={unavailableMessage} />
      </div>
    {/if}
  {:else if listError}
    <p class="text-sm text-red-600">
      {m.feedback_inbox_error({ message: listError.message })}
    </p>
  {:else if $listQuery?.isLoading}
    <div class="flex justify-center py-6">
      <DelayedSpinner isLoading size="20px" />
    </div>
  {:else if items.length === 0}
    <div class="py-6">
      <ResourceListEmptyState
        icon={Inbox}
        message={m.feedback_inbox_empty({
          status: feedbackStatusLabels[status](),
        })}
      />
    </div>
  {:else}
    <div class="flex flex-col gap-y-2">
      {#each items as item (item.id)}
        <FeedbackInboxItem
          {item}
          {client}
          {canSuggestFix}
          {onCategoryClick}
          activeCategory={categoryFilter}
          expanded={expandedId === item.id}
          onToggle={() => (expandedId = expandedId === item.id ? null : item.id)}
          onResolve={(newStatus) => handleResolve(item, newStatus)}
        />
      {/each}
    </div>
  {/if}
</section>
