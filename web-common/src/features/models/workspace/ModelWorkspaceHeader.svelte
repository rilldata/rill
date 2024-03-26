<script lang="ts">
  import { goto } from "$app/navigation";
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import { renameFileArtifact } from "../../entity-management/actions";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "../../entity-management/entity-mappers";
  import {
    INVALID_NAME_MESSAGE,
    VALID_NAME_PATTERN,
    isDuplicateName,
  } from "../../entity-management/name-utils";
  import ModelWorkspaceCTAs from "./ModelWorkspaceCTAs.svelte";
  import { workspaces } from "@rilldata/web-common/layout/workspace/workspace-stores";

  export let modelName: string;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  $: workspaceLayout = $workspaces;

  $: tableVisible = workspaceLayout.table.visible;

  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelHasError = getFileHasErrors(
    queryClient,
    runtimeInstanceId,
    modelPath,
  );

  let contextMenuOpen = false;

  $: availableDashboards = useGetDashboardsForModel(
    runtimeInstanceId,
    modelName,
  );

  function formatModelName(str: string) {
    return str.replace(/\.sql/, "");
  }

  async function onChangeCallback(
    e: Event & {
      currentTarget: EventTarget & HTMLInputElement;
    },
  ) {
    if (!e.currentTarget.value.match(VALID_NAME_PATTERN)) {
      notifications.send({
        message: INVALID_NAME_MESSAGE,
      });
      e.currentTarget.value = modelName; // resets the input
      return;
    }
    if (
      isDuplicateName(
        e.currentTarget.value,
        modelName,
        $allNamesQuery?.data ?? [],
      )
    ) {
      notifications.send({
        message: `Name ${e.currentTarget.value} is already in use`,
      });
      e.currentTarget.value = modelName; // resets the input
      return;
    }

    try {
      const toName = e.currentTarget.value;
      const entityType = EntityType.Model;
      await renameFileArtifact(
        runtimeInstanceId,
        getFileAPIPathFromNameAndType(modelName, entityType),
        getFileAPIPathFromNameAndType(toName, entityType),
        entityType,
      );
      await goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
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
