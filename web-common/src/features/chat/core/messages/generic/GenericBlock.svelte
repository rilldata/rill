<!--
  Renders a file diff block with collapsible tool call header.
  Shows the diff visualization with expandable request/response details.
-->
<script lang="ts">
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { GenericBlock } from "@rilldata/web-common/features/chat/core/messages/generic/generic-block.ts";

  export let block: GenericBlock;
  export let tools: V1Tool[] | undefined = undefined;
</script>

<div class="generic-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />
</div>

<style lang="postcss">
  .generic-block {
    @apply w-full max-w-full self-start;
  }
</style>
