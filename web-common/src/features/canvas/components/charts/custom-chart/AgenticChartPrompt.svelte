<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
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
</script>

{#if manualMode}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="text-sm text-fg-secondary text-center">
      Use the inspector panel to write Metrics SQL and Vega-Lite spec manually.
    </div>
    <Button type="text" onClick={() => (manualMode = false)}>
      ← Back to prompt
    </Button>
  </div>
{:else if isStreaming}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="text-sm text-fg-secondary">
      <AnimatedDots>AI is generating chart</AnimatedDots>
    </div>
  </div>
{:else}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="w-full max-w-md flex flex-col gap-3">
      <textarea
        class="w-full px-3 py-2 text-sm border rounded-md border-gray-300 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 resize-none bg-white text-fg-primary placeholder:text-fg-muted"
        rows="3"
        placeholder="Describe the chart you want to see..."
        bind:value={prompt}
      >
      </textarea>

      <div class="flex items-center justify-between">
        <Button type="text" onClick={() => (manualMode = true)}>
          Write SQL & Vega-Lite manually
        </Button>

        <Button
          type="primary"
          disabled={!prompt.trim()}
          onClick={handleGenerate}
        >
          Generate
        </Button>
      </div>
    </div>
  </div>
{/if}
