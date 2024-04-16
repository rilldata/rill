<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import ChartPromptHistoryDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptHistoryDisplay.svelte";
  import { createChartGenerator } from "@rilldata/web-common/features/charts/prompt/generateChart";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let open: boolean;
  export let chart: string;
  export let filePath: string;

  let prompt: string;

  $: generateVegaConfig = createChartGenerator(
    $runtime.instanceId,
    chart,
    filePath,
  );

  async function createVegaConfig() {
    open = false;
    await generateVegaConfig(prompt);
  }
</script>

<Dialog on:close={() => (open = false)} {open}>
  <svelte:fragment slot="title">Generate vega config using AI</svelte:fragment>
  <svelte:fragment slot="body">
    <InputV2 bind:value={prompt} error="" label="Prompt" />
    <ChartPromptHistoryDisplay
      entityName={chart}
      on:reuse-prompt={({ detail }) => {
        prompt = detail;
      }}
    />
  </svelte:fragment>
  <div class="pt-2" slot="footer">
    <Button on:click={createVegaConfig}>Generate</Button>
  </div>
</Dialog>
