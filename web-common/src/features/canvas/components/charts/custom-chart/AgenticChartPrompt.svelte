<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { generateChart } from "./generateChart";
  import type { CustomChartComponent } from "./index";

  export let component: CustomChartComponent;

  const client = useRuntimeClient();

  let prompt = "";
  let generating = false;
  let error: string | null = null;
  let manualMode = false;

  async function handleGenerate() {
    if (!prompt.trim()) return;

    generating = true;
    error = null;

    try {
      const result = await generateChart(client, {
        prompt: prompt.trim(),
      });

      component.updateProperty("prompt", prompt.trim());
      component.updateProperty("metrics_sql", result.metricsSql);
      component.updateProperty("vega_spec", result.vegaSpec);
    } catch (e) {
      error = e instanceof Error ? e.message : "Chart generation failed";
    } finally {
      generating = false;
    }
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
{:else if generating}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <Spinner />
    <span class="text-sm text-gray-500">Generating chart...</span>
  </div>
{:else}
  <div class="flex flex-col items-center justify-center h-full gap-3 p-4">
    <div class="w-full max-w-md flex flex-col gap-3">
      <textarea
        class="w-full px-3 py-2 text-sm border rounded-md border-gray-300 focus:outline-none focus:ring-1 focus:ring-primary-500 focus:border-primary-500 resize-none bg-white"
        rows="3"
        placeholder="Describe the chart you want to see..."
        bind:value={prompt}
        on:keydown={handleKeydown}
      />

      {#if error}
        <div class="text-xs text-red-500">{error}</div>
      {/if}

      <div class="flex items-center justify-between">
        <button
          class="text-xs text-gray-400 hover:text-gray-600 underline"
          on:click={switchToManual}
        >
          Write SQL & Vega-Lite manually
        </button>

        <Button
          type="primary"
          small
          onClick={handleGenerate}
          disabled={!prompt.trim()}
        >
          Generate
        </Button>
      </div>
    </div>
  </div>
{/if}
