<!--
  Renders tool invocations with their results in a collapsible interface.
  
  Architecture: Tool calls and results are rendered together in one component,
  with results correlated via parent_id. Charts are rendered separately below.
-->
<script lang="ts">
  import CaretDownIcon from "../../../../components/icons/CaretDownIcon.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import { isHiddenAgentTool, MessageContentType, ToolName } from "../types";
  import { parseChartData } from "../utils";
  import ChartBlock from "./ChartBlock.svelte";

  export let message: V1Message;
  export let resultMessage: V1Message | undefined = undefined;

  let isExpanded = false;

  // Call message properties
  $: toolName = message.tool || "Unknown Tool";
  $: toolInput = formatContentData(message);
  $: isHidden = isHiddenAgentTool(message.tool);

  // Result message properties
  $: hasResult = !!resultMessage;
  $: isError = resultMessage?.contentType === MessageContentType.ERROR;
  $: resultContent = resultMessage ? formatContentData(resultMessage) : "";

  // Chart detection and parsing
  $: isChart = isChartCall(message);
  $: chartData = isChart
    ? parseChartData({ input: message.contentData })
    : null;

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

  function isChartCall(message: V1Message): boolean {
    return message.tool === ToolName.CREATE_CHART;
  }
</script>

{#if !isHidden}
  <div class="tool-container">
    <button class="tool-header" on:click={toggleExpanded}>
      <div class="tool-icon">
        {#if isExpanded}
          <CaretDownIcon size="16" />
        {:else}
          <ChevronRight size="16" />
        {/if}
      </div>
      <div class="tool-name">
        {toolName}
      </div>
    </button>

    {#if isExpanded}
      <div class="tool-content">
        <div class="tool-section">
          <div class="tool-section-title">Request</div>
          <div class="tool-section-content">
            <pre class="tool-json">{toolInput}</pre>
          </div>
        </div>

        {#if hasResult}
          <div class="tool-section">
            <div class="tool-section-title">
              {isError ? "Error" : "Response"}
            </div>
            <div class="tool-section-content">
              <pre class="tool-json">{resultContent}</pre>
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>

  {#if isChart && chartData && hasResult && !isError}
    <!-- Chart visualization (shown below the collapsible tool details) -->
    <div class="chart-container">
      <ChartBlock
        chartType={chartData.chartType}
        chartSpec={chartData.chartSpec}
      />
    </div>
  {/if}
{/if}

<style lang="postcss">
  .tool-container {
    @apply w-full max-w-[90%] self-start;
    @apply border border-gray-200 rounded-lg bg-gray-50;
  }

  .chart-container {
    @apply w-full max-w-full mt-2 self-start;
  }

  .tool-header {
    @apply w-full flex items-center gap-2 p-2;
    @apply bg-transparent border-none cursor-pointer;
    @apply text-sm transition-colors;
  }

  .tool-header:hover {
    @apply bg-gray-100;
  }

  .tool-icon {
    @apply text-gray-500 flex items-center;
  }

  .tool-name {
    @apply font-medium text-gray-700 flex-1 text-left;
  }

  .tool-content {
    @apply border-t border-gray-200 bg-white rounded-b-lg;
  }

  .tool-section {
    @apply p-2;
  }

  .tool-section:not(:last-child) {
    @apply border-b border-gray-50;
  }

  .tool-section-title {
    @apply text-[0.625rem] font-semibold text-gray-500 mb-1.5;
    @apply uppercase tracking-wide;
  }

  .tool-section-content {
    @apply bg-gray-50 border border-gray-200;
    @apply rounded-md overflow-hidden;
  }

  .tool-json {
    @apply font-mono text-xs leading-snug text-gray-700;
    @apply p-2 m-0 overflow-x-auto whitespace-pre-wrap break-all;
  }
</style>
