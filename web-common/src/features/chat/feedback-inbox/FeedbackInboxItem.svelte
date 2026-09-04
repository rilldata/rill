<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Tag from "@rilldata/web-common/components/tag/Tag.svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { ThumbsDown, ThumbsUp } from "lucide-svelte";
  import { feedbackCategoryLabel } from "../core/feedback/feedback-categories";
  import type { InboxItem } from "./inbox-data";
  import FeedbackDetail from "./FeedbackDetail.svelte";

  export let item: InboxItem;
  // The runtime client "runtime"-origin items were listed with; unused for
  // "local-service" items, whose detail/resolve calls go through the bridge.
  export let client: RuntimeClient;
  export let canSuggestFix: boolean;
  export let expanded: boolean;
  export let onToggle: () => void;
  export let onResolve: (status: "open" | "addressed" | "dismissed") => void;
  // Clicking a category tag filters the inbox by that category.
  export let onCategoryClick: ((category: string) => void) | undefined =
    undefined;
  export let activeCategory: string | null = null;

  $: isOpen = item.status === "open";
</script>

<div class="item" class:expanded>
  <div class="item-row">
    <div class="item-main">
      <button class="item-toggle" onclick={onToggle}>
        <div class="item-header">
          <Tag
            color={item.kind === "review_request" ? "blue" : "gray"}
            height={18}
            text={item.kind === "review_request"
              ? m.feedback_inbox_kind_review_request()
              : m.feedback_inbox_kind_rating()}
          />
          {#if item.sentiment === "negative"}
            <ThumbsDown size="12px" class="text-fg-muted" />
          {:else if item.sentiment === "positive"}
            <ThumbsUp size="12px" class="text-fg-muted" />
          {/if}
          <span class="title">
            {item.conversationTitle || m.feedback_inbox_conversation_untitled()}
          </span>
        </div>
        {#if item.comment}
          <p class="comment">{item.comment}</p>
        {/if}
      </button>
      <div class="item-meta">
        {#if item.predictedAttribution}
          <Tag
            color="gray"
            height={18}
            text={m.feedback_inbox_attribution({
              attribution: item.predictedAttribution,
            })}
          />
        {/if}
        {#each item.categories as category (category)}
          {#if onCategoryClick}
            <button
              class="category-button"
              onclick={() => onCategoryClick?.(category)}
              title={m.feedback_inbox_filter_by_category()}
            >
              <Tag
                color={activeCategory === category ? "blue" : "gray"}
                height={18}
                text={feedbackCategoryLabel(category)}
              />
            </button>
          {:else}
            <Tag
              color="gray"
              height={18}
              text={feedbackCategoryLabel(category)}
            />
          {/if}
        {/each}
        {#if item.ownerEmail}
          <span class="meta-text">{item.ownerEmail}</span>
        {/if}
        {#if item.createdOn}
          <span class="meta-text">{item.createdOn.toLocaleString()}</span>
        {/if}
      </div>
    </div>
    <div class="item-actions">
      {#if isOpen}
        <Button type="secondary" small onClick={() => onResolve("addressed")}>
          {m.feedback_inbox_addressed()}
        </Button>
        <Button type="text" small onClick={() => onResolve("dismissed")}>
          {m.feedback_inbox_dismiss()}
        </Button>
      {:else}
        <Button type="text" small onClick={() => onResolve("open")}>
          {m.feedback_inbox_reopen()}
        </Button>
      {/if}
    </div>
  </div>

  {#if expanded}
    <FeedbackDetail {item} {client} {canSuggestFix} {onResolve} />
  {/if}
</div>

<style lang="postcss">
  .item {
    @apply rounded-md border px-3 py-2;
  }

  .item-row {
    @apply flex w-full items-start justify-between gap-2;
  }

  .item-main {
    @apply flex flex-col gap-1 min-w-0 flex-1;
  }

  .item-toggle {
    @apply flex flex-col gap-1 text-left;
  }

  .item-header {
    @apply flex items-center gap-2;
  }

  .title {
    @apply text-sm font-medium truncate;
  }

  .comment {
    @apply border-l-2 pl-2 text-sm text-fg-secondary;
  }

  .item-meta {
    @apply flex flex-wrap items-center gap-2;
  }

  .category-button {
    @apply cursor-pointer;
  }

  .meta-text {
    @apply text-xs text-fg-muted;
  }

  .item-actions {
    @apply flex shrink-0 items-center gap-1;
  }
</style>
