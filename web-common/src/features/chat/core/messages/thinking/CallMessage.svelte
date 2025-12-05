<!--
  Renders tool invocations with their results in a collapsible interface.
  
  Architecture: Tool calls and results are rendered together in one component,
  with results correlated via parent_id.
-->
<script lang="ts">
  import CodeBlock from "../../../../../components/code-block/CodeBlock.svelte";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import LoadingSpinner from "../../../../../components/icons/LoadingSpinner.svelte";
  import type { V1Message, V1Tool } from "../../../../../runtime-client";
  import { isHiddenAgentTool, MessageContentType } from "../../types";
  import { getToolDisplayName } from "./tool-display-names";
  import { getToolIcon } from "./tool-icons";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;
  export let tools: V1Tool[] | undefined = undefined;

  let isExpanded = false;
  let activeTab: "request" | "response" = "request";

  // Call message properties
  $: toolInput = formatContentData(message);
  $: isHidden = isHiddenAgentTool(message.tool);

  // Result message properties
  $: hasResult = !!resultMessage;
  $: isError = resultMessage?.contentType === MessageContentType.ERROR;
  $: resultContent = resultMessage ? formatContentData(resultMessage) : "";

  // Display name from API metadata (tools passed from parent)
  $: toolDisplayName = getToolDisplayName(
    message.tool || "Unknown Tool",
    hasResult,
    tools,
  );

  // Icon based on tool type
  $: ToolIcon = getToolIcon(message.tool);

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }

  function formatContentData(msg: V1Message): string {
    const rawContent = msg.contentData || "";

    switch (msg.contentType) {
      case MessageContentType.JSON:
        try {
          const obj = JSON.parse(rawContent);
          return JSON.stringify(obj, null, 2);
        } catch {
          return rawContent;
        }

      case MessageContentType.TEXT:
        return rawContent;

      case MessageContentType.ERROR:
        return rawContent;

      default:
        // Fallback for unknown content types
        return rawContent;
    }
  }
</script>

{#if !isHidden}
  <div class="tool-container">
    <button class="tool-header" on:click={toggleExpanded}>
      <div class="tool-icon">
        {#if isExpanded}
          <CaretDownIcon size="14" />
        {:else}
          <svelte:component this={ToolIcon} size="14" />
        {/if}
      </div>
      <div class="tool-name">
        {toolDisplayName}
      </div>
    </button>

    {#if isExpanded}
      <div class="tool-content">
        <!-- Tabs -->
        <div class="tool-tabs">
          <button
            class="tool-tab"
            class:active={activeTab === "request"}
            on:click={() => (activeTab = "request")}
          >
            Request
          </button>
          <button
            class="tool-tab"
            class:active={activeTab === "response"}
            on:click={() => (activeTab = "response")}
          >
            {isError ? "Error" : "Response"}
          </button>
        </div>

        <!-- Tab Content -->
        <div class="tool-tab-content">
          {#if activeTab === "request"}
            <div class="tool-code">
              <CodeBlock code={toolInput} language="json" showCopyButton />
            </div>
          {:else if hasResult}
            <div class="tool-code">
              <CodeBlock code={resultContent} language="json" showCopyButton />
            </div>
          {:else}
            <div class="tool-loading">
              <LoadingSpinner size="0.875rem" />
              <span>Waiting for response...</span>
            </div>
          {/if}
        </div>
      </div>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .tool-container {
    @apply w-full max-w-full self-start;
  }

  .tool-header {
    @apply w-full flex items-center gap-1.5 px-1 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-gray-500 transition-colors;
    @apply sticky top-0;
    background: var(--surface);
    z-index: 2;
  }

  .tool-header:hover {
    @apply text-gray-600;
  }

  .tool-icon {
    @apply flex items-center;
  }

  .tool-name {
    @apply font-normal flex-1 text-left;
  }

  .tool-content {
    @apply pt-1 ml-5;
  }

  .tool-tabs {
    @apply flex gap-1 pb-2;
    @apply sticky top-6;
    background: var(--surface);
    z-index: 1;
  }

  .tool-tab {
    @apply px-2 py-1 text-[0.625rem] font-normal text-gray-500;
    @apply bg-transparent border-none cursor-pointer;
    @apply transition-colors rounded;
  }

  .tool-tab:hover {
    @apply text-gray-600 bg-gray-50;
  }

  .tool-tab.active {
    @apply text-gray-700 bg-gray-100;
  }

  .tool-tab-content {
    @apply bg-gray-50/50 border border-gray-100;
    @apply rounded-md overflow-hidden;
  }

  .tool-code {
    @apply text-[0.625rem];
  }

  .tool-code :global(pre) {
    @apply font-mono text-[0.625rem] leading-snug text-gray-500;
    @apply m-0 overflow-x-auto whitespace-pre-wrap break-all;
  }

  .tool-loading {
    @apply p-2 flex items-center gap-2 text-[0.625rem] text-gray-500;
  }
</style>
