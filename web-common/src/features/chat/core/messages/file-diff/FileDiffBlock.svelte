<!--
  Renders a file diff block with collapsible tool call header.
  Shows the diff visualization with expandable request/response details.
-->
<script lang="ts">
  import Diff2HtmlView from "@rilldata/web-common/components/diff/Diff2HtmlView.svelte";
  import type { V1Tool } from "../../../../../runtime-client";
  import ToolCall from "../tools/ToolCall.svelte";
  import type { FileDiffBlock } from "./file-diff-block";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing";

  export let block: FileDiffBlock;
  export let tools: V1Tool[] | undefined = undefined;
</script>

<div class="file-diff-block">
  <ToolCall
    message={block.message}
    resultMessage={block.resultMessage}
    {tools}
    variant="block"
  />

  <div class="diff-container">
    <div class="diff-header">
      <a href={getFileHref(block.filePath)} class="file-path-link">
        {block.filePath}
      </a>
      {#if block.isNewFile}
        <span class="new-badge">new</span>
      {/if}
    </div>
    <Diff2HtmlView diff={block.diff} scrollX>
      <div slot="empty" class="no-changes-message">No changes detected</div>
    </Diff2HtmlView>
  </div>
</div>

<style lang="postcss">
  .file-diff-block {
    @apply w-full max-w-full self-start;
  }

  /* Diff container */
  .diff-container {
    @apply border border-gray-200 rounded-md overflow-hidden;
  }

  .diff-header {
    @apply flex items-center gap-2 px-3 py-1.5;
    @apply text-xs border-b border-gray-200 bg-gray-100;
  }

  .file-path-link {
    @apply text-fg-secondary font-mono;
  }

  .file-path-link:hover {
    @apply text-fg-primary underline;
  }

  .new-badge {
    @apply text-[0.5rem] px-1 py-0.5 rounded;
    @apply bg-green-100 text-green-700 font-medium;
  }

  .no-changes-message {
    @apply p-3 text-xs text-fg-secondary italic;
  }
</style>
