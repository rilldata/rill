<script lang="ts">
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import InputLabel from "@rilldata/web-common/components/forms/InputLabel.svelte";
  import Spinner from "@rilldata/web-common/features/entity-management/Spinner.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import {
    CustomChartComponent,
    type CustomChart,
  } from "../components/charts/custom-chart/index";
  import { generateChart } from "../components/charts/custom-chart/generateChart";
  import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";
  import { writable, type Readable } from "svelte/store";

  export let component: BaseCanvasComponent;

  const client = useRuntimeClient();

  let generating = false;
  let error: string | null = null;
  let localPrompt = "";

  const emptySpec = writable<CustomChart>({
    metrics_sql: [""],
    vega_spec: "",
  });

  $: customChart = component instanceof CustomChartComponent ? component : null;
  $: specStore = (customChart?.specStore ?? emptySpec) as Readable<CustomChart>;
  $: spec = $specStore;

  $: prompt = spec.prompt ?? "";
  $: hasSpec =
    !!spec.vega_spec &&
    Array.isArray(spec.metrics_sql) &&
    spec.metrics_sql.some((s) => s.trim().length > 0);

  // Initialize localPrompt from spec only once, then let user edit freely
  let initialized = false;
  $: if (!initialized && prompt) {
    localPrompt = prompt;
    initialized = true;
  }

  async function handleRegenerate() {
    if (!customChart || !localPrompt.trim()) return;

    generating = true;
    error = null;

    try {
      const result = await generateChart(client, {
        prompt: localPrompt.trim(),
        previousSql: spec.metrics_sql,
        previousSpec: spec.vega_spec,
      });

      customChart.updateProperty("prompt", localPrompt.trim());
      customChart.updateProperty("metrics_sql", result.metricsSql);
      customChart.updateProperty("vega_spec", result.vegaSpec);
    } catch (e) {
      error = e instanceof Error ? e.message : "Generation failed";
    } finally {
      generating = false;
    }
  }

  function handleBlur() {
    if (customChart && localPrompt !== prompt) {
      customChart.updateProperty("prompt", localPrompt);
    }
  }
</script>

{#if customChart}
  <div class="flex flex-col gap-y-2">
    <InputLabel small label="AI Prompt" id="ai-prompt" />
    <textarea
      class="w-full p-2 text-xs border border-gray-300 rounded-sm resize-none"
      rows="3"
      placeholder="Describe the chart you want to see..."
      bind:value={localPrompt}
      on:blur={handleBlur}
    />

    {#if error}
      <div class="text-xs text-red-500">{error}</div>
    {/if}

    <div class="flex items-center justify-end gap-2">
      {#if generating}
        <div class="flex items-center gap-1 text-xs text-gray-500">
          <Spinner size="14px" />
          Generating...
        </div>
      {:else}
        <Button
          type="primary"
          small
          onClick={handleRegenerate}
          disabled={!localPrompt.trim()}
        >
          {hasSpec ? "Regenerate" : "Generate"}
        </Button>
      {/if}
    </div>
  </div>
{/if}
