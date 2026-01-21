<script lang="ts">
  import { Share, Link, Check } from "lucide-svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import IconButton from "@rilldata/web-common/components/button/IconButton.svelte";
  import {
    Popover,
    PopoverContent,
    PopoverTrigger,
  } from "@rilldata/web-common/components/popover";
  import {
    createRuntimeServiceShareConversation,
    type RpcStatus,
  } from "@rilldata/web-common/runtime-client";

  export let conversationId: string;
  export let instanceId: string;
  export let organization: string;
  export let project: string;

  let isOpen = false;
  let copied = false;
  let isSharing = false;
  let shareError: string | null = null;

  const shareConversationMutation = createRuntimeServiceShareConversation();

  async function handleCreateLink() {
    if (isSharing) return;

    isSharing = true;
    shareError = null;

    try {
      // Call the share API to set the sharing boundary
      await $shareConversationMutation.mutateAsync({
        instanceId,
        conversationId,
        data: {},
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
      }, 1500);
    } catch (error) {
      console.error("[ShareChatPopover] Share failed:", error);
      shareError =
        (error as { response?: { data?: RpcStatus } })?.response?.data
          ?.message ?? "Failed to create share link";
    } finally {
      isSharing = false;
    }
  }
</script>

<Popover bind:open={isOpen}>
  <PopoverTrigger>
    <IconButton
      ariaLabel="Share conversation"
      bgGray
      active={isOpen}
      disableTooltip={isOpen}
    >
      <Share class="text-gray-500" size="16px" />
      <svelte:fragment slot="tooltip-content"
        >Share conversation</svelte:fragment
      >
    </IconButton>
  </PopoverTrigger>
  <PopoverContent align="end" class="w-[320px] p-4">
    <div class="flex flex-col gap-y-3">
      <h3 class="text-sm font-medium text-gray-800">Share conversation</h3>
      <p class="text-xs text-gray-600">
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
