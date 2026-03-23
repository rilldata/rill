<script lang="ts">
  import AnimatedDots from "@rilldata/web-common/features/chat/core/messages/AnimatedDots.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { CustomChartComponent } from "../components/charts/custom-chart/index";
  import {
    sendToDevAgent,
    getAgentStreamingStore,
  } from "../components/charts/custom-chart/chart-ai-agent";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";

  export let component: BaseCanvasComponent;

  const client = useRuntimeClient();

  let prompt = "";

  $: customChart = component instanceof CustomChartComponent ? component : null;
  $: streamingStore = customChart
    ? getAgentStreamingStore(client, customChart.id)
    : null;
  $: isStreaming = streamingStore ? $streamingStore : false;

  function handleGenerate() {
    if (!customChart || !prompt.trim() || isStreaming) return;
    sendToDevAgent(client, customChart, prompt.trim());
    prompt = "";
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleGenerate();
    }
  }
</script>

{#if customChart}
  <div class="ai-generate">
    {#if isStreaming}
      <div class="status">
        <AnimatedDots>AI is editing</AnimatedDots>
      </div>
    {:else}
      <div class="input-row">
        <textarea
          class="ai-input"
          rows="2"
          placeholder="Describe chart changes..."
          bind:value={prompt}
          on:keydown={handleKeydown}
        />
        <button
          class="generate-btn"
          on:click={handleGenerate}
          disabled={!prompt.trim()}
          aria-label="Edit with AI"
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none">
            <path
              d="M14.5 1.5L7 9M14.5 1.5L10 14.5L7 9M14.5 1.5L1.5 6L7 9"
              stroke="currentColor"
              stroke-width="1.5"
              stroke-linecap="round"
              stroke-linejoin="round"
            />
          </svg>
        </button>
      </div>
      <div class="hint">Opens the AI assistant to edit this chart</div>
    {/if}
  </div>
{/if}

<style lang="postcss">
  .ai-generate {
    @apply flex flex-col gap-1;
  }

  .input-row {
    @apply flex items-end gap-1.5;
  }

  .ai-input {
    @apply flex-1 px-2.5 py-1.5 text-xs;
    @apply border border-gray-200 rounded-md;
    @apply resize-none outline-none;
    min-height: 28px;
    max-height: 64px;
    field-sizing: content;
  }

  .ai-input:focus {
    @apply border-primary-400 ring-1 ring-primary-200;
  }

  .generate-btn {
    @apply p-1.5 rounded-md;
    @apply text-gray-400 transition-colors;
  }

  .generate-btn:hover {
    @apply text-primary-500 bg-gray-100;
  }

  .generate-btn:disabled {
    @apply opacity-30 cursor-not-allowed;
  }

  .hint {
    @apply text-[10px] text-gray-400;
  }

  .status {
    @apply text-xs text-gray-500 py-1;
  }
</style>
