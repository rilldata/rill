<!--
  Renders a simple tool call block with collapsible tool call header.
  It will only have raw tool call request and responses.
-->
<script lang="ts">
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { SimpleToolCall } from "@rilldata/web-common/features/chat/core/messages/simple-tool-call/simple-tool-call.ts";

  export let block: SimpleToolCall;
  export let tools: V1Tool[] | undefined = undefined;
</script>

<div class="simple-tool-call">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />
</div>

<style lang="postcss">
  .simple-tool-call {
    @apply w-full max-w-full self-start;
  }
</style>
