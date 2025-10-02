<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import ChatTextMessage from "./ChatTextMessage.svelte";
  import ChatToolMessage from "./ChatToolMessage.svelte";
  import { handleApplyMetricsProps } from "../../utils/apply-metrics-props";
  import Button from "../../../../components/button/Button.svelte";

  export let message: V1Message;
  export let isInDashboard: boolean = false;

  let isApplying = false;
  let applyError: string | null = null;

  // Handle apply button click
  async function handleApply() {
    const metricsProps = getMetricsProps();
    const toolCall = getToolCallInfo();
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

  // Helper to get text content from message
  function getTextContent(message: V1Message): string {
    if (!message.content) return "";
    return message.content
      .filter((block) => block.text)
      .map((block) => block.text)
      .join(" ");
  }

  // Check if message has only text content
  function hasOnlyText(message: V1Message): boolean {
    if (!message.content || message.content.length === 0) return true;
    return message.content.every(
      (block) => block.text && !block.toolCall && !block.toolResult,
    );
  }

  // Check if message contains metricsProps in any tool result
  function hasMetricsProps(): boolean {
    if (!message.content) return false;

    return message.content.some((block) => {
      if (block.toolResult?.content) {
        try {
          const content =
            typeof block.toolResult.content === "string"
              ? JSON.parse(block.toolResult.content)
              : block.toolResult.content;
          return (
            content && typeof content === "object" && "metricsProps" in content
          );
        } catch {
          return false;
        }
      }
      return false;
    });
  }

  // Check if message contains open_url in any tool result
  function hasOpenURL(): boolean {
    if (!message.content) return false;

    return message.content.some((block) => {
      if (block.toolResult?.content) {
        try {
          const content =
            typeof block.toolResult.content === "string"
              ? JSON.parse(block.toolResult.content)
              : block.toolResult.content;
          return (
            content &&
            typeof content === "object" &&
            "open_url" in content &&
            content.open_url
          );
        } catch {
          return false;
        }
      }
      return false;
    });
  }

  // Extract open_url from the first tool result that has it
  function getOpenURL(): string | null {
    if (!message.content) return null;

    for (const block of message.content) {
      if (block.toolResult?.content) {
        try {
          const content =
            typeof block.toolResult.content === "string"
              ? JSON.parse(block.toolResult.content)
              : block.toolResult.content;
          if (
            content &&
            typeof content === "object" &&
            "open_url" in content &&
            content.open_url
          ) {
            return content.open_url;
          }
        } catch {
          continue;
        }
      }
    }
    return null;
  }

  // Handle open URL button click
  function handleOpenURL() {
    const openURL = getOpenURL();
    if (openURL) {
      window.open(openURL, "_blank");
    }
  }

  // Extract metricsProps from the first tool result that has them
  function getMetricsProps(): any {
    if (!message.content) return null;

    for (const block of message.content) {
      if (block.toolResult?.content) {
        try {
          const content =
            typeof block.toolResult.content === "string"
              ? JSON.parse(block.toolResult.content)
              : block.toolResult.content;
          if (
            content &&
            typeof content === "object" &&
            "metricsProps" in content
          ) {
            return content.metricsProps;
          }
        } catch {
          continue;
        }
      }
    }
    return null;
  }

  // Extract tool call info for context
  function getToolCallInfo(): any {
    if (!message.content) return null;

    for (const block of message.content) {
      if (block.toolCall) {
        return block.toolCall;
      }
    }
    return null;
  }

  $: isTextOnly = hasOnlyText(message);
  $: textContent = getTextContent(message);
  $: hasMetricsPropsData = hasMetricsProps();
  $: hasOpenURLData = hasOpenURL();
  $: metricsProps = getMetricsProps();
  $: toolCall = getToolCallInfo();

  // Show apply button only when in dashboard and has metricsProps
  $: showApplyButton = isInDashboard && hasMetricsPropsData;
  // Show open URL button only when outside dashboard and has open_url
  $: showOpenURLButton = !isInDashboard && hasOpenURLData;

  // Debug logging
  $: console.log("ChatMessage Debug:", {
    isTextOnly,
    isInDashboard,
    showApplyButton,
    showOpenURLButton,
    hasMetricsProps: hasMetricsPropsData,
    hasOpenURL: hasOpenURLData,
    metricsProps,
    messageContent: message.content,
    toolResults: message.content
      ?.filter((block) => block.toolResult)
      .map((block) => ({
        hasContent: !!block.toolResult?.content,
        contentType: typeof block.toolResult?.content,
        contentPreview:
          typeof block.toolResult?.content === "string"
            ? block.toolResult.content.substring(0, 200) + "..."
            : block.toolResult?.content,
      })),
  });
</script>

{#if isTextOnly}
  <ChatTextMessage
    {message}
    content={textContent}
    {showApplyButton}
    {metricsProps}
    {toolCall}
    {isInDashboard}
  />
{:else}
  <ChatToolMessage {message} />
  {#if showApplyButton}
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

  {#if showOpenURLButton}
    <div class="open-url-button-container">
      <Button type="secondary" onClick={handleOpenURL}>Open in New Tab</Button>
    </div>
  {/if}
{/if}

<style lang="postcss">
  .apply-button-container,
  .open-url-button-container {
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
