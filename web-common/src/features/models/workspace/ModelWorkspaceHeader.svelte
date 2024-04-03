<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { handleEntityRename } from "@rilldata/web-common/features/entity-management/ui-actions";
  import { extractFileName } from "@rilldata/web-common/features/sources/extract-file-name";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import ModelWorkspaceCTAs from "./ModelWorkspaceCTAs.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";

  export let filePath: string;

  const queryClient = useQueryClient();

  $: modelName = extractFileName(filePath);

  $: runtimeInstanceId = $runtime.instanceId;

  $: workspaceLayout = $workspaces;

  $: tableVisible = workspaceLayout.table.visible;

  $: fileArtifact = fileArtifacts.getFileArtifact(filePath);
  $: modelHasError = fileArtifact.getHasErrors(queryClient, runtimeInstanceId);

  let contextMenuOpen = false;

  $: availableDashboards = useGetDashboardsForModel(
    runtimeInstanceId,
    modelName,
  );

  function formatModelName(str: string) {
    return str.replace(/\.sql/, "");
  }

  async function onChangeCallback(e) {
    return handleEntityRename(
      queryClient,
      runtimeInstanceId,
      e,
      filePath,
      EntityType.Model,
    );
  }

  $: titleInput = modelName;
</script>

<WorkspaceHeader
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
>
  <svelte:fragment slot="workspace-controls">
    <IconButton on:click={workspaceLayout.table.toggle}
      ><span class="text-gray-500"><HideBottomPane size="18px" /></span>
      <svelte:fragment slot="tooltip-content">
        <SlidingWords active={$tableVisible} reverse
          >results preview</SlidingWords
        >
      </svelte:fragment>
    </IconButton>
  </svelte:fragment>
  <svelte:fragment let:width slot="cta">
    {@const collapse = width < 800}
    <PanelCTA side="right">
      <ModelWorkspaceCTAs
        availableDashboards={$availableDashboards?.data ?? []}
        {collapse}
        modelHasError={$modelHasError}
        {modelName}
        suppressTooltips={contextMenuOpen}
      />
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>
