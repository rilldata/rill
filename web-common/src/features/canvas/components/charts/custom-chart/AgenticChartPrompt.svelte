<script lang="ts">
  import AnimatedDots from "@rilldata/web-common/features/chat/core/messages/AnimatedDots.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getAgentStreamingStore, sendToDevAgent } from "./chart-ai-agent";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;

  const client = useRuntimeClient();

  let prompt = "";
  let manualMode = false;

  $: streamingStore = getAgentStreamingStore(client, component.id);
  $: isStreaming = $streamingStore;

  function handleGenerate() {
    if (!prompt.trim() || isStreaming) return;
    sendToDevAgent(client, component, prompt.trim());
  }

  function handleKeydown(event: KeyboardEvent) {
    if (event.key === "Enter" && !event.shiftKey) {
      event.preventDefault();
      handleGenerate();
    }
  }

  function switchToManual() {
    manualMode = true;
  }
</script>

{#if manualMode}
  <div
    class="flex items-center justify-center h-full text-gray-400 text-sm p-4"
  >
    Use the inspector panel to write Metrics SQL and Vega-Lite spec manually.
  </div>
{:else if isStreaming}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="text-sm text-gray-500">
      <AnimatedDots>AI is generating chart</AnimatedDots>
    </div>
  </div>
{:else}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="w-full max-w-md flex flex-col gap-3">
      <textarea
        class="w-full px-3 py-2 text-sm border rounded-md border-gray-300 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 resize-none bg-white"
        rows="3"
        placeholder="Describe the chart you want to see..."
        bind:value={prompt}
        onkeydown={handleKeydown}
      >
      </textarea>

      <div class="flex items-center justify-between">
        <button
          class="text-xs text-gray-400 hover:text-gray-600 underline"
          onclick={switchToManual}
        >
          Write SQL & Vega-Lite manually
        </button>

        <button
          class="px-3 py-1.5 text-xs font-medium text-white bg-primary-500 rounded-md hover:bg-primary-600 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          onclick={handleGenerate}
          disabled={!prompt.trim()}
        >
          Generate
        </button>
      </div>
    </div>
  </div>
{/if}
