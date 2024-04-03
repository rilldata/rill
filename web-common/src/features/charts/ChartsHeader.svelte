<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import GenerateVegaSpecPrompt from "@rilldata/web-common/features/charts/prompt/GenerateVegaSpecPrompt.svelte";
  import { renameFileArtifact } from "@rilldata/web-common/features/entity-management/actions";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "@rilldata/web-common/features/entity-management/name-utils";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let chartName: string;

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(VALID_NAME_PATTERN)) {
      notifications.send({
        message: INVALID_NAME_MESSAGE,
      });
      e.target.value = chartName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, chartName, $allNamesQuery.data ?? [])) {
      notifications.send({
        message: `Name ${e.target.value} is already in use`,
      });
      e.target.value = chartName; // resets the input
      return;
    }

    try {
      const toName = e.target.value;
      const type = EntityType.Chart;
      await renameFileArtifact(
        runtimeInstanceId,
        getFileAPIPathFromNameAndType(chartName, type),
        getFileAPIPathFromNameAndType(toName, type),
        type,
      );
      goto(`/chart/${toName}`, { replaceState: true });
    } catch (err) {
      console.error(err.response.data.message);
    }
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
