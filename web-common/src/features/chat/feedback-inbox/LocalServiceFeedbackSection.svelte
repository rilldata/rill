<!--
  The "cloud deployment" inbox section in web-local: items are read from the
  project's cloud deployment through the LocalService bridge, which reports
  precise unavailability states (not logged in, not deployed, no permission).
-->
<script lang="ts">
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import ResourceListEmptyState from "@rilldata/web-common/features/resources/ResourceListEmptyState.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { CloudFeedbackState } from "@rilldata/web-common/proto/gen/rill/local/v1/api_pb";
  import { createLocalServiceListProjectAIFeedback } from "@rilldata/web-common/runtime-client/local-service";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
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

  export let status: FeedbackStatus;
  export let canSuggestFix: boolean;
  // Category filter shared across the inbox's sections; clicking a category
  // tag on an item toggles it.
  export let categoryFilter: string | null = null;
  export let onCategoryClick: ((category: string) => void) | undefined =
    undefined;

  // Suggest-fix on cloud items still runs against the local runtime,
  // which has the repo checkout and AI connector.
  const client = useRuntimeClient();

  let expandedId: string | null = null;
  let resolveError: string | null = null;

  $: cloudQuery = createLocalServiceListProjectAIFeedback({ status });
  $: cloudState = $cloudQuery.data?.cloudState;
  $: items = ($cloudQuery.data?.feedback ?? [])
    .map((f) => normalizeFeedback(f, "local-service"))
    .filter((i) => !categoryFilter || i.categories.includes(categoryFilter));

  $: unavailableMessage =
    cloudState === CloudFeedbackState.NOT_LOGGED_IN
      ? m.feedback_inbox_not_logged_in()
      : cloudState === CloudFeedbackState.NOT_DEPLOYED
        ? m.feedback_inbox_not_deployed()
        : cloudState === CloudFeedbackState.NO_PERMISSION
          ? m.feedback_inbox_no_permission()
          : null;

  async function handleResolve(item: InboxItem, newStatus: FeedbackStatus) {
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
  <h2 class="text-sm font-semibold text-fg-secondary">
    {m.feedback_inbox_cloud_section()}
    {#if $cloudQuery.data?.org}
      <span class="font-normal text-fg-muted">
        {$cloudQuery.data.org}/{$cloudQuery.data.project}
      </span>
    {/if}
  </h2>
  {#if resolveError}
    <p class="text-sm text-red-600">
      {m.feedback_inbox_error({ message: resolveError })}
    </p>
  {/if}
  {#if unavailableMessage}
    <div class="py-6">
      <ResourceListEmptyState icon={Inbox} message={unavailableMessage} />
    </div>
  {:else if cloudState === CloudFeedbackState.ERROR}
    <p class="text-sm text-red-600">
      {m.feedback_inbox_error({
        message: $cloudQuery.data?.cloudStateMessage ?? "",
      })}
    </p>
  {:else if $cloudQuery.isLoading}
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
