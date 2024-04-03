<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/charts/prompt/GenerateVegaSpecPrompt.svelte";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { useQueryClient } from "@tanstack/svelte-query";

  export let filePath: string;
  $: chartName = extractFileName(filePath);

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  const onChangeCallback = async (e) => {
    return handleEntityRename(
      queryClient,
      runtimeInstanceId,
      e,
      filePath,
      EntityType.Chart,
    );
  };

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
