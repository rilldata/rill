<script lang="ts">
  import { Chip } from "@rilldata/web-common/components/chip";
  import {
    Tabs,
    UnderlineTabsList,
    UnderlineTabsTrigger,
  } from "@rilldata/web-common/components/tabs";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { feedbackCategoryLabel } from "../core/feedback/feedback-categories";
  import {
    feedbackStatusLabels,
    feedbackStatusOptions,
    type FeedbackStatus,
  } from "./inbox-data";

  export let status: FeedbackStatus;
  export let onStatusChange: (status: FeedbackStatus) => void;
  // Active category filter, set by clicking a category tag on an item.
  export let categoryFilter: string | null = null;
  export let onClearCategory: () => void = () => {};
</script>

<p class="text-sm text-fg-muted">{m.feedback_inbox_description()}</p>

<div class="flex flex-wrap items-center justify-between gap-2">
  <Tabs
    value={status}
    onValueChange={(value) => onStatusChange(value as FeedbackStatus)}
  >
    <UnderlineTabsList>
      {#each feedbackStatusOptions as option (option)}
        <UnderlineTabsTrigger value={option}>
          {feedbackStatusLabels[option]()}
        </UnderlineTabsTrigger>
      {/each}
    </UnderlineTabsList>
  </Tabs>

  {#if categoryFilter}
    <Chip
      type="dimension"
      compact
      removable
      label={feedbackCategoryLabel(categoryFilter)}
      removeTooltipText={m.feedback_inbox_clear_category()}
      onRemove={onClearCategory}
    >
      <span slot="body">{feedbackCategoryLabel(categoryFilter)}</span>
    </Chip>
  {/if}
</div>
