<!--
  Routes messages to specialized rendering components based on message type.
-->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { extractMessageText } from "../utils";
  import CallMessage from "./CallMessage.svelte";
  import ProgressMessage from "./ProgressMessage.svelte";
  import TextMessage from "./TextMessage.svelte";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;

  $: isRouterAgent = message.tool === "router_agent";
  $: content = extractMessageText(message);
</script>

{#if isRouterAgent}
  <!-- User prompts and assistant responses -->
  <!-- Note: role should always be set by backend, but fallback to "assistant" for safety -->
  <TextMessage {content} role={message.role || "assistant"} />
{:else if message.type === "progress"}
  <!-- Progress/thinking messages -->
  <ProgressMessage {message} />
{:else if message.type === "call"}
  <!-- Tool call messages (results will be passed in to show together) -->
  <CallMessage {message} {resultMessage} />
{/if}
