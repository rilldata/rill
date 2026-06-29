<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import AnimatedDots from "@rilldata/web-common/features/chat/core/messages/AnimatedDots.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getAgentStreamingStore, sendToDevAgent } from "./chart-ai-agent";
  import type { CustomChartComponent } from "./index";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

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
      {m.canvas_ai_manual_hint()}
    </div>
    <Button type="text" onClick={() => (manualMode = false)}>
      {m.canvas_back_to_prompt()}
    </Button>
  </div>
{:else if isStreaming}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="text-sm text-fg-secondary">
      <AnimatedDots>{m.canvas_ai_generating_chart()}</AnimatedDots>
    </div>
  </div>
{:else}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="w-full max-w-md flex flex-col gap-3">
      <textarea
        class="w-full px-3 py-2 text-sm border rounded-md border-gray-300 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 resize-none bg-white text-fg-primary placeholder:text-fg-muted"
        rows="3"
        placeholder={m.canvas_describe_chart_prompt()}
        bind:value={prompt}
      >
      </textarea>

      <div class="flex items-center justify-between">
        <Button type="text" onClick={() => (manualMode = true)}>
          {m.canvas_ai_write_manually()}
        </Button>

        <Button
          type="primary"
          disabled={!prompt.trim()}
          onClick={handleGenerate}
        >
          {m.canvas_generate()}
        </Button>
      </div>
    </div>
  </div>
{/if}
