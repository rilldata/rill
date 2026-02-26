<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import { createRuntimeServiceShareConversationMutation } from "@rilldata/web-common/runtime-client";
  import { isHTTPError } from "@rilldata/web-common/lib/errors";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { Check, Link, Share } from "lucide-svelte";

  export let conversationId: string | undefined = undefined;
  export let organization: string | undefined = undefined;
  export let project: string | undefined = undefined;
  export let disabled = false;

  const DISABLED_TOOLTIP = "Start a conversation to share";
  const COPIED_FEEDBACK_DURATION_MS = 1500;

  let isOpen = false;
  let copied = false;
  let isSharing = false;
  let shareError: string | null = null;

  const runtimeClient = useRuntimeClient();
  const shareConversationMutation =
    createRuntimeServiceShareConversationMutation(runtimeClient);

  async function handleCreateLink() {
    if (copied || isSharing || !conversationId || !organization || !project)
      return;

    isSharing = true;
    shareError = null;

    try {
      // Call the share API to set the sharing boundary
      await $shareConversationMutation.mutateAsync({
        conversationId,
      });

      // Construct the share URL
      const shareUrl = `${window.location.origin}/${organization}/${project}/-/ai/${conversationId}`;

      // Copy to clipboard
      await navigator.clipboard.writeText(shareUrl);

      // Show success state, then close popover
      copied = true;
      setTimeout(() => {
        copied = false;
        isOpen = false;
      }, COPIED_FEEDBACK_DURATION_MS);
    } catch (error) {
      console.error("[ShareChatPopover] Share failed:", error);
      shareError = isHTTPError(error)
        ? error.response.data.message
        : "Failed to create share link";
    } finally {
      isSharing = false;
    }
  }
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger {disabled}>
    <IconButton
      ariaLabel="Share conversation"
      bgGray
      active={isOpen}
      {disabled}
      disableTooltip={isOpen}
    >
      <Share class="text-fg-muted" size="16px" />
      <svelte:fragment slot="tooltip-content">
        {disabled ? DISABLED_TOOLTIP : "Share conversation"}
      </svelte:fragment>
    </IconButton>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[320px] p-4">
    <div class="flex flex-col gap-y-3">
      <h3 class="text-sm font-medium text-fg-primary">Share conversation</h3>
      <p class="text-xs text-fg-secondary">
        Share this conversation with other project members. They can view and
        continue the conversation.
      </p>
      {#if shareError}
        <p class="text-xs text-red-600">{shareError}</p>
      {/if}
      <Button type="secondary" onClick={handleCreateLink} disabled={isSharing}>
        {#if copied}
          <Check size="16px" class="text-green-600" />
          Copied!
        {:else if isSharing}
          Creating link...
        {:else}
          <Link size="16px" class="text-primary-500" />
          Create link
        {/if}
      </Button>
    </div>
  </PopoverContent>
</Popover>
