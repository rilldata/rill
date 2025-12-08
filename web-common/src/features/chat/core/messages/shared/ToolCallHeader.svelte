<!--
  Reusable tool call header with collapsible request/response details.
  Used by block-level tools (ChartBlock, FileDiffBlock) to show tool invocation info.
-->
<script lang="ts">
  import CodeBlock from "../../../../../components/code-block/CodeBlock.svelte";
  import CaretDownIcon from "../../../../../components/icons/CaretDownIcon.svelte";
  import type { V1Message, V1Tool } from "../../../../../runtime-client";
  import { MessageContentType } from "../../types";
  import { getToolDisplayName } from "../thinking/tool-display-names";
  import { getToolIcon } from "../thinking/tool-icons";

  export let message: V1Message;
  export let resultMessage: V1Message;
  export let tools: V1Tool[] | undefined = undefined;

  let isExpanded = false;
  let activeTab: "request" | "response" = "request";

  // Tool display info
  $: toolDisplayName = getToolDisplayName(message.tool || "", true, tools);
  $: ToolIcon = getToolIcon(message.tool);

  // Format request/response JSON
  $: requestJson = formatJson(message.contentData);
  $: responseJson = formatJson(resultMessage.contentData);
  $: isError = resultMessage.contentType === MessageContentType.ERROR;

  function toggleExpanded() {
    isExpanded = !isExpanded;
  }

  function formatJson(data: string | undefined): string {
    if (!data) return "";
    try {
      return JSON.stringify(JSON.parse(data), null, 2);
    } catch {
      return data;
    }
  }
</script>

<div class="tool-call-header">
  <button class="header-button" on:click={toggleExpanded}>
    <div class="header-icon">
      {#if isExpanded}
        <CaretDownIcon size="14" />
      {:else}
        <svelte:component this={ToolIcon} size="14" />
      {/if}
    </div>
    <div class="header-name">
      {toolDisplayName}
    </div>
  </button>

  {#if isExpanded}
    <div class="header-content">
      <div class="header-tabs">
        <button
          class="header-tab"
          class:active={activeTab === "request"}
          on:click={() => (activeTab = "request")}
        >
          Request
        </button>
        <button
          class="header-tab"
          class:active={activeTab === "response"}
          on:click={() => (activeTab = "response")}
        >
          {isError ? "Error" : "Response"}
        </button>
      </div>
      <div class="header-tab-content">
        {#if activeTab === "request"}
          <div class="header-code">
            <CodeBlock code={requestJson} language="json" showCopyButton />
          </div>
        {:else}
          <div class="header-code">
            <CodeBlock code={responseJson} language="json" showCopyButton />
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style lang="postcss">
  .tool-call-header {
    @apply mb-2;
  }

  .header-button {
    @apply w-full flex items-center gap-1.5 py-1;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-xs text-gray-500 transition-colors;
  }

  .header-button:hover {
    @apply text-gray-600;
  }

  .header-icon {
    @apply flex items-center;
  }

  .header-name {
    @apply font-normal flex-1 text-left;
  }

  .header-content {
    @apply pt-1 mb-2;
  }

  .header-tabs {
    @apply flex gap-1 pb-2;
  }

  .header-tab {
    @apply px-2 py-1 text-[0.625rem] font-normal text-gray-500;
    @apply bg-transparent border-none cursor-pointer;
    @apply transition-colors rounded;
  }

  .header-tab:hover {
    @apply text-gray-600 bg-gray-50;
  }

  .header-tab.active {
    @apply text-gray-700 bg-gray-100;
  }

  .header-tab-content {
    @apply bg-gray-50/50 border border-gray-100;
    @apply rounded-md overflow-hidden;
  }

  .header-code {
    @apply text-[0.625rem];
  }

  .header-code :global(pre) {
    @apply font-mono text-[0.625rem] leading-snug text-gray-500;
    @apply m-0 overflow-x-auto whitespace-pre-wrap break-all;
  }
</style>
