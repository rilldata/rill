<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/charts/prompt/GenerateVegaSpecPrompt.svelte";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";

  export let chartName: string;

  function onChangeCallback() {
    console.log("edit");
  }
  $: titleInput = chartName;

  let generateOpen = false;
</script>

<WorkspaceHeader {...{ titleInput, onChangeCallback }}>
  <svelte:fragment slot="cta">
    <PanelCTA side="right">
      <Button on:click={() => (generateOpen = true)}>Generate using AI</Button>
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>

<GenerateVegaSpecPrompt bind:open={generateOpen} chart={chartName} />
