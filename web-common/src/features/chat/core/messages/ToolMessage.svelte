<script lang="ts">
  import ChartBlock from "@rilldata/web-common/features/chat/core/messages/ChartBlock.svelte";
  import {
    isChartToolResult,
    isHiddenAgentTool,
    parseChartData,
  } from "@rilldata/web-common/features/chat/core/utils";
  import type { V1Message } from "../../../../runtime-client";
  import TextMessage from "./TextMessage.svelte";
  import ToolCallBlock from "./ToolCallBlock.svelte";

  export let message: V1Message;

  // State for collapsible content blocks
  let expandedBlocks: { [key: number]: boolean } = {};

  function toggleBlock(index: number) {
    expandedBlocks[index] = !expandedBlocks[index];
  }

  // TODO: This correlation logic should be moved to the Conversation class business layer.
  // Currently we have split responsibility: Conversation handles streaming correlation,
  // but UI handles initial fetch correlation. This creates inconsistent architecture.
  // All correlation should happen in the business layer, with UI just displaying pre-correlated data.

  // Helper to create a tool block
  function createToolBlock(toolCall: any, toolResult: any, index: number) {
    return { type: "tool", toolCall, toolResult, index };
  }

  // Helper to create a chart block
  function createChartBlock(
    chartData: any,
    toolCall: any,
    toolResult: any,
    index: number,
  ) {
    return {
      type: "chart",
      chartType: chartData.chartType,
      chartSpec: chartData.chartSpec,
      toolCall,
      toolResult,
      index,
    };
  }

  // Helper to process a tool call block
  function processToolCallBlock(
    block: any,
    index: number,
    toolResults: Map<string, any>,
  ): any[] {
    // Filter out high-level agent invocations
    if (isHiddenAgentTool(block.toolCall?.name)) {
      return [];
    }

    // Streaming merges tool results into toolCall blocks.
    // For initial fetch (GetConversation), calls/results are separate, so we attach via fallback.
    const toolResult = block.toolResult || toolResults.get(block.toolCall.id);

    if (!isChartToolResult(toolResult, block.toolCall)) {
      return [createToolBlock(block.toolCall, toolResult, index)];
    }

    // Try to parse chart data
    const chartData = parseChartData(block.toolCall);
    if (!chartData) {
      // Parsing failed, fallback to regular tool block
      return [createToolBlock(block.toolCall, toolResult, index)];
    }

    // Add both tool block and chart block
    return [
      createToolBlock(block.toolCall, toolResult, index),
      createChartBlock(chartData, block.toolCall, toolResult, index),
    ];
  }

  // Group tool calls with their results within this message
  function groupToolCallsWithResults(content: any[]) {
    // First pass: collect all tool results by ID within this message
    const toolResults = new Map();
    for (const block of content) {
      if (block.toolResult?.id) {
        toolResults.set(block.toolResult.id, block.toolResult);
      }
    }

    // Second pass: process blocks and create groups
    const groups: any[] = [];
    for (let index = 0; index < content.length; index++) {
      const block = content[index];

      if (block.text) {
        groups.push({ type: "text", content: block.text, index });
        continue;
      }

      if (block.toolCall) {
        const toolBlocks = processToolCallBlock(block, index, toolResults);
        groups.push(...toolBlocks);
        continue;
      }

      // Skip standalone toolResult blocks as they should be merged with toolCalls
    }

    return groups;
  }

  $: groupedContent = message.content
    ? groupToolCallsWithResults(message.content)
    : [];
</script>

<div class="complex-message-container">
  {#each groupedContent as group, i (i)}
    {#if group.type === "text"}
      <!-- Text block -->
      <TextMessage {message} content={group.content} />
    {:else if group.type === "chart"}
      <!-- Chart block -->
      <ChartBlock chartType={group.chartType} chartSpec={group.chartSpec} />
    {:else if group.type === "tool"}
      <!-- Tool Call + Result block -->
      <ToolCallBlock
        toolCall={group.toolCall}
        toolResult={group.toolResult}
        isExpanded={expandedBlocks[i] || false}
        onToggle={() => toggleBlock(i)}
      />
    {/if}
  {/each}
</div>

<style lang="postcss">
  .complex-message-container {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    max-width: 90%;
    align-self: flex-start;
    width: 100%;
  }
</style>
