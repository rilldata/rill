<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import * as Dialog from "@rilldata/web-common/components/dialog";
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
    // Close immediately - let the chat show the feedback response
    handleClose();
    // Submit asynchronously - conversation will handle streaming
    void conversation.submitFeedback(
      messageId,
      "negative",
      selectedCategories,
      comment || undefined,
    );
  }

  function handleSkip() {
    if (!messageId) return;
    // Close immediately - let the chat show the feedback response
    handleClose();
    // Submit asynchronously with no categories
    void conversation.submitFeedback(messageId, "negative");
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
  portal="#rill-portal"
>
  <Dialog.Content class="max-w-md">
    <Dialog.Header>
      <Dialog.Title>Give feedback</Dialog.Title>
      <Dialog.Description>Select all that apply.</Dialog.Description>
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
        label="Comments"
        optional
        bind:value={comment}
        placeholder="Type here..."
        rows={3}
      />
    </div>

    <Dialog.Footer class="gap-x-2">
      <Button type="secondary" onClick={handleSkip}>Skip</Button>
      <Button type="primary" onClick={handleSubmit} disabled={isSubmitDisabled}>
        Submit
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
