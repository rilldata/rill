<!--
  Routes messages to specialized rendering components based on message type.
-->
<script lang="ts">
  import type { V1Message } from "../../../../runtime-client";
  import { MessageType, ToolName } from "../types";
  import CallMessage from "./CallMessage.svelte";
  import ProgressMessage from "./ProgressMessage.svelte";
  import TextMessage from "./TextMessage.svelte";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;

  $: isRouterAgent = message.tool === ToolName.ROUTER_AGENT;
</script>

{#if isRouterAgent}
  <TextMessage {message} />
{:else if message.type === MessageType.PROGRESS}
  <!-- Progress/thinking messages -->
  <ProgressMessage {message} />
{:else if message.type === MessageType.CALL}
  <!-- Tool call messages (results will be passed in to show together) -->
  <CallMessage {message} {resultMessage} />
{/if}
