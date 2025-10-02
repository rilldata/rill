<script lang="ts">
  import Markdown from "../../../../components/markdown/Markdown.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import { handleApplyMetricsProps } from "../../utils/apply-metrics-props";
  import Button from "../../../../components/button/Button.svelte";

  export let message: V1Message;
  export let content: string;
  export let showApplyButton: boolean = false;
  export let metricsProps: any = null;
  export let toolCall: any = null;
  export let isInDashboard: boolean = false;

  let isApplying = false;
  let applyError: string | null = null;

  $: shouldShowApplyButton = showApplyButton && isInDashboard;

  async function handleApply() {
    if (!metricsProps) return;

    isApplying = true;
    applyError = null;

    try {
      await handleApplyMetricsProps(metricsProps, toolCall);
    } catch (error) {
      applyError =
        error instanceof Error
          ? error.message
          : "Failed to apply metrics props";
      console.error("Failed to apply metrics props:", error);
    } finally {
      isApplying = false;
    }
  }

  $: role = message.role;
</script>

<div class="chat-message chat-message--{role}">
  <div class="chat-message-content">
    {#if role === "assistant"}
      <Markdown {content} />
    {:else}
      {content}
    {/if}
  </div>

  {#if shouldShowApplyButton && role === "assistant"}
    <div class="apply-button-container">
      <Button
        type="primary"
        loading={isApplying}
        onClick={handleApply}
        disabled={isApplying}
        loadingCopy="Applying..."
      >
        Apply to Dashboard
      </Button>
      {#if applyError}
        <div class="apply-error">
          {applyError}
        </div>
      {/if}
    </div>
  {/if}
</div>

<style lang="postcss">
  .chat-message {
    max-width: 90%;
  }

  .chat-message--user {
    align-self: flex-end;
  }

  .chat-message--assistant {
    align-self: flex-start;
  }

  .chat-message-content {
    padding: 0.375rem 0.5rem;
    border-radius: 1rem;
    font-size: 0.875rem;
    line-height: 1.5;
    word-break: break-word;
  }

  .chat-message--user .chat-message-content {
    @apply bg-primary-400 text-white rounded-br-lg;
  }

  .chat-message--assistant .chat-message-content {
    color: #374151;
  }

  .apply-button-container {
    margin-top: 0.75rem;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .apply-error {
    padding: 0.5rem;
    background: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 0.375rem;
    color: #dc2626;
    font-size: 0.875rem;
  }
</style>
