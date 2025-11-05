<!--
  Renders tool invocations with their results in a collapsible interface.
  
  Architecture: Tool calls and results are rendered together in one component,
  with results correlated via parent_id. Charts are rendered separately below.
-->
<script lang="ts">
  import CaretDownIcon from "../../../../components/icons/CaretDownIcon.svelte";
  import ChevronRight from "../../../../components/icons/ChevronRight.svelte";
  import type { V1Message } from "../../../../runtime-client";
  import { isHiddenAgentTool, parseChartData } from "../utils";
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
  $: isError = resultMessage?.contentType === "error";
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
      case "json":
        try {
          const obj = JSON.parse(rawContent);
          return JSON.stringify(obj, null, 2);
        } catch {
          return rawContent;
        }

      case "text":
        return rawContent;

      case "error":
        return rawContent;

      default:
        // Fallback for unknown content types
        return rawContent;
    }
  }

  function isChartCall(message: V1Message): boolean {
    return message.tool === "create_chart";
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
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    background: #fafafa;
    width: 100%;
    max-width: 90%;
    align-self: flex-start;
  }

  .chart-container {
    width: 100%;
    max-width: 100%;
    margin-top: 0.5rem;
    align-self: flex-start;
  }

  .tool-header {
    width: 100%;
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem;
    background: none;
    border: none;
    cursor: pointer;
    font-size: 0.875rem;
    transition: background-color 0.15s ease;
  }

  .tool-header:hover {
    background: #f3f4f6;
  }

  .tool-icon {
    color: #6b7280;
    display: flex;
    align-items: center;
  }

  .tool-name {
    font-weight: 500;
    color: #374151;
    flex: 1;
    text-align: left;
  }

  .tool-content {
    border-top: 1px solid #e5e7eb;
    background: #ffffff;
    border-radius: 0 0 0.5rem 0.5rem;
  }

  .tool-section {
    padding: 0.5rem;
  }

  .tool-section:not(:last-child) {
    border-bottom: 1px solid #f3f4f6;
  }

  .tool-section-title {
    font-size: 0.625rem;
    font-weight: 600;
    color: #6b7280;
    margin-bottom: 0.375rem;
    text-transform: uppercase;
    letter-spacing: 0.025em;
  }

  .tool-section-content {
    background: #f9fafb;
    border: 1px solid #e5e7eb;
    border-radius: 0.375rem;
    overflow: hidden;
  }

  .tool-json {
    font-family:
      "SF Mono", Monaco, "Cascadia Code", "Roboto Mono", Consolas,
      "Courier New", monospace;
    font-size: 0.75rem;
    line-height: 1.4;
    color: #374151;
    padding: 0.5rem;
    margin: 0;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }
</style>
