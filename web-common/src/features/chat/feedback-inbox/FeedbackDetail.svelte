<script lang="ts">
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { EntityStatus } from "@rilldata/web-common/features/entity-management/types";
  import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ThumbsDown, ThumbsUp } from "lucide-svelte";
  import { feedbackCategoryLabel } from "../core/feedback/feedback-categories";
  import { fetchFeedbackDetail, type InboxItem } from "./inbox-data";
  import ReadOnlyTranscript from "./ReadOnlyTranscript.svelte";
  import SuggestFixCard from "./SuggestFixCard.svelte";

  export let item: InboxItem;
  export let client: RuntimeClient;
  // Whether this surface can write project files (dev workspaces: web-local
  // outside readonly mode, cloud edit sessions). The global readOnly flag is
  // not a usable signal here because web-admin sets it even in edit sessions.
  export let canSuggestFix: boolean;
  export let onResolve: (status: "open" | "addressed" | "dismissed") => void;

  // Fetch keyed on the item's identity, not the object reference: list refetches rebuild the
  // item objects, and re-fetching here would tear down the transcript and any in-progress
  // suggest-fix state below. loadDetail is a plain function so `item` isn't a reactive dependency.
  let loadedKey: string | null = null;
  let detailPromise: ReturnType<typeof fetchFeedbackDetail>;
  $: itemKey = `${item.origin}:${item.id}`;
  $: if (itemKey !== loadedKey) {
    loadedKey = itemKey;
    detailPromise = loadDetail();
  }

  function loadDetail() {
    return fetchFeedbackDetail(client, item);
  }
</script>

<div class="detail">
  <div class="user-feedback">
    <span class="label">{m.feedback_inbox_user_feedback()}</span>
    <div class="user-feedback-body">
      {#if item.sentiment === "negative"}
        <ThumbsDown size="14px" class="shrink-0 text-fg-muted" />
      {:else if item.sentiment === "positive"}
        <ThumbsUp size="14px" class="shrink-0 text-fg-muted" />
      {/if}
      <p class:no-comment={!item.comment}>
        {item.comment || m.feedback_inbox_no_comment()}
      </p>
    </div>
    {#if item.categories.length > 0}
      <div class="user-feedback-categories">
        {#each item.categories as category (category)}
          <Tag color="gray" height={18} text={feedbackCategoryLabel(category)} />
        {/each}
      </div>
    {/if}
  </div>

  {#if item.suggestedAction}
    <div class="suggested-action">
      <span class="label">{m.feedback_inbox_suggested_action()}</span>
      <p>{item.suggestedAction}</p>
    </div>
  {/if}

  {#await detailPromise}
    <div class="loading">
      <Spinner status={EntityStatus.Running} size="16px" />
    </div>
  {:then detail}
    {#if detail.transcriptAvailable}
      <ReadOnlyTranscript messages={detail.messages} />
    {:else}
      <p class="unavailable">{m.feedback_inbox_transcript_unavailable()}</p>
    {/if}
    {#if canSuggestFix}
      <SuggestFixCard {item} messages={detail.messages} {onResolve} />
    {/if}
  {:catch error}
    <p class="unavailable">{error.message}</p>
  {/await}
</div>

<style lang="postcss">
  .detail {
    @apply flex flex-col gap-2 border-t pt-2 mt-2;
  }

  .user-feedback {
    @apply flex flex-col gap-1 text-sm rounded-md bg-surface-hover px-3 py-2;
  }

  .user-feedback-body {
    @apply flex items-start gap-2;
  }

  .user-feedback .no-comment {
    @apply italic text-fg-muted;
  }

  .user-feedback-categories {
    @apply flex flex-wrap items-center gap-1.5;
  }

  .suggested-action {
    @apply text-sm rounded-md bg-surface-hover px-3 py-2;
  }

  .label {
    @apply text-xs font-semibold uppercase text-fg-muted;
  }

  .loading {
    @apply flex justify-center py-4;
  }

  .unavailable {
    @apply text-sm text-fg-muted italic;
  }
</style>
