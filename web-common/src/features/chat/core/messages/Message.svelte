<!--
  Routes messages to specialized rendering components based on message type.
-->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import CallMessage from "./CallMessage.svelte";
  import ProgressMessage from "./ProgressMessage.svelte";
  import TextMessage from "./TextMessage.svelte";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;

  $: effectiveRole = getEffectiveRole(message);
  $: isRouterAgent = isRouterAgentMessage(message);
  $: content = extractTextContent(message);

  function getEffectiveRole(message: V1Message): string {
    if (message.type === "call" && message.tool === "router_agent") {
      return "user";
    }
    if (message.type === "result" && message.tool === "router_agent") {
      return "assistant";
    }
    return message.role || "";
  }

  function extractTextContent(message: V1Message): string {
    const rawContent = message.contentData || "";

    switch (message.contentType) {
      case "json":
        // For router_agent, parse JSON and extract prompt/response field
        if (message.tool === "router_agent") {
          try {
            const parsed = JSON.parse(rawContent);
            return parsed.prompt || parsed.response || rawContent;
          } catch {
            return rawContent;
          }
        }
        // Non-router_agent JSON messages shouldn't reach TextMessage (they go to CallMessage)
        // But if they do, return raw content as fallback
        return rawContent;

      case "text":
        return rawContent;

      case "error":
        return rawContent;

      default:
        // Fallback for unknown content types
        return rawContent;
    }
  }

  function isRouterAgentMessage(message: V1Message): boolean {
    return message.tool === "router_agent";
  }
</script>

{#if isRouterAgent}
  <!-- User prompts and assistant responses -->
  <TextMessage {content} role={effectiveRole} />
{:else if message.type === "progress"}
  <!-- Progress/thinking messages -->
  <ProgressMessage {message} />
{:else if message.type === "call"}
  <!-- Tool call messages (results will be passed in to show together) -->
  <CallMessage {message} {resultMessage} />
{/if}
