<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import Textarea from "@rilldata/web-common/components/forms/Textarea.svelte";
  import type { Conversation } from "../conversation";
  import CheckboxCard from "@rilldata/web-common/components/forms/CheckboxCard.svelte";
  import {
    type FeedbackCategory,
    getCategoriesForAgent,
  } from "./feedback-categories";

  export let open: boolean;
  export let messageId: string | null;
  export let conversation: Conversation;
  export let agent: string;
  export let onClose: () => void;

  let selectedCategories: FeedbackCategory[] = [];
  let comment = "";

  $: categories = getCategoriesForAgent(agent);
  $: isSubmitDisabled = selectedCategories.length === 0;

  function handleCategoryToggle(categoryId: FeedbackCategory) {
    if (selectedCategories.includes(categoryId)) {
      selectedCategories = selectedCategories.filter((c) => c !== categoryId);
    } else {
      selectedCategories = [...selectedCategories, categoryId];
    }
  }

  function handleSubmit() {
    if (!messageId || isSubmitDisabled) return;
    // Submit before closing: handleClose resets the form and clears messageId
    // (via the parent), so the arguments must be read first. The call captures
    // them synchronously and streams in the background, so the modal still
    // closes immediately. Negative feedback is always flagged for admin review.
    void conversation.submitFeedback(
      messageId,
      "negative",
      selectedCategories,
      comment || undefined,
      { requestReview: true },
    );
    handleClose();
  }

  function handleSkip() {
    if (!messageId) return;
    // Submit with no categories before closing (see handleSubmit for ordering).
    void conversation.submitFeedback(messageId, "negative", [], undefined, {
      requestReview: true,
    });
    handleClose();
  }

  function handleClose() {
    onClose();
    resetForm();
  }

  function resetForm() {
    selectedCategories = [];
    comment = "";
  }
</script>

<Dialog.Root
  {open}
  onOpenChange={(isOpen) => {
    if (!isOpen) handleClose();
  }}
>
  <Dialog.Content class="max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{m.chat_feedback_title()}</Dialog.Title>
      <Dialog.Description>{m.chat_feedback_select_all()}</Dialog.Description>
    </Dialog.Header>

    <div class="feedback-form">
      <div class="categories">
        {#each categories as category}
          <CheckboxCard
            checked={selectedCategories.includes(category.id)}
            label={category.label}
            onCheckedChange={() => handleCategoryToggle(category.id)}
          />
        {/each}
      </div>

      <Textarea
        id="feedback-comment"
        label={m.chat_feedback_comments()}
        optional
        bind:value={comment}
        placeholder={m.chat_feedback_placeholder()}
        rows={3}
      />
    </div>

    <Dialog.Footer class="gap-x-2">
      <Button type="secondary" onClick={handleSkip}
        >{m.chat_feedback_skip()}</Button
      >
      <Button type="primary" onClick={handleSubmit} disabled={isSubmitDisabled}>
        {m.chat_feedback_submit()}
      </Button>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

<style lang="postcss">
  .feedback-form {
    @apply flex flex-col gap-4 py-4;
  }

  .categories {
    @apply flex flex-wrap gap-2;
  }
</style>
