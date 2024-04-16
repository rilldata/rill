<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import Dialog from "@rilldata/web-common/components/dialog/Dialog.svelte";
  import InputV2 from "@rilldata/web-common/components/forms/InputV2.svelte";
  import ChartPromptHistoryDisplay from "@rilldata/web-common/features/charts/prompt/ChartPromptHistoryDisplay.svelte";
  import { createFullChartGenerator } from "@rilldata/web-common/features/charts/prompt/generateChart";
  import { useChartFileNames } from "@rilldata/web-common/features/charts/selectors";
  import { getName } from "@rilldata/web-common/features/entity-management/name-utils";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let open: boolean;
  export let table: string = "";
  export let connector: string = "";
  export let metricsView: string = "";

  let prompt: string;

  $: chartFileNames = useChartFileNames($runtime.instanceId);
  $: generateVegaConfig = createFullChartGenerator($runtime.instanceId);

  async function createVegaConfig() {
    open = false;
    await generateVegaConfig(
      prompt,
      {
        table,
        connector,
        metricsView,
      },
      getName(`${table || metricsView}_chart`, $chartFileNames.data ?? []),
    );
  }
</script>

<Dialog on:close={() => (open = false)} {open}>
  <svelte:fragment slot="title">
    Generate chart yaml for "{table || metricsView}" using AI
  </svelte:fragment>
  <svelte:fragment slot="body">
    <InputV2 bind:value={prompt} error="" label="Prompt" />
    <ChartPromptHistoryDisplay
      entityName={table || metricsView}
      on:reuse-prompt={({ detail }) => {
        prompt = detail;
      }}
    />
  </svelte:fragment>
  <div class="pt-2" slot="footer">
    <Button on:click={createVegaConfig}>Generate</Button>
  </div>
</Dialog>
