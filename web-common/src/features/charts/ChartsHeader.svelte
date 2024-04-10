<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/charts/prompt/GenerateVegaSpecPrompt.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let filePath: string;
  $: chartName = extractFileName(filePath);

  $: runtimeInstanceId = $runtime.instanceId;

  async function handleNameChange(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    const newRoute = await handleEntityRename(
      runtimeInstanceId,
      e.currentTarget,
      filePath,
      EntityType.Chart,
    );

    if (newRoute) await goto(newRoute);
  }

  $: titleInput = chartName;

  let generateOpen = false;
</script>

<WorkspaceHeader on:change={handleNameChange} {titleInput}>
  <svelte:fragment slot="cta">
    <PanelCTA side="right">
      <Button on:click={() => (generateOpen = true)}>Generate using AI</Button>
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>

<GenerateVegaSpecPrompt bind:open={generateOpen} chart={chartName} />
