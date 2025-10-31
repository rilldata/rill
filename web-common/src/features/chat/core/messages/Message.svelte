<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import TextMessage from "./TextMessage.svelte";
  import ToolMessage from "./ToolMessage.svelte";

  export let message: V1Message;

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

  $: isTextOnly = hasOnlyText(message);
  $: textContent = getTextContent(message);
</script>

{#if isTextOnly}
  <TextMessage {message} content={textContent} />
{:else}
  <ToolMessage {message} />
{/if}
