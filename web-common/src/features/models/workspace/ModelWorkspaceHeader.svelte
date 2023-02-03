<script lang="ts">
  import { goto } from "$app/navigation";
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import { notifications } from "@rilldata/web-common/components/notifications";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { useRuntimeServiceRenameFileAndReconcile } from "@rilldata/web-common/runtime-client";
  import {
    appQueryStatusStore,
    runtimeStore,
  } from "@rilldata/web-local/lib/application-state-stores/application-store";
  import type { LayoutElement } from "@rilldata/web-local/lib/types";
  import { useQueryClient } from "@sveltestack/svelte-query";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import { renameFileArtifact } from "../../entity-management/actions";
  import {
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "../../entity-management/entity-mappers";
  import { isDuplicateName } from "../../entity-management/name-utils";
  import { useAllNames } from "../../entity-management/selectors";
  import ModelWorkspaceCTAs from "./ModelWorkspaceCTAs.svelte";

  export let modelName: string;

  $: runtimeInstanceId = $runtimeStore.instanceId;

  $: allNamesQuery = useAllNames(runtimeInstanceId);
  const queryClient = useQueryClient();
  const renameModel = useRuntimeServiceRenameFileAndReconcile();

  const outputLayout = getContext(
    "rill:app:output-layout"
  ) as Writable<LayoutElement>;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);
  $: modelError = $fileArtifactsStore.entities[modelPath]?.errors[0]?.message;
  $: modelHasError = !!modelError;

  let contextMenuOpen = false;

  $: availableDashboards = useGetDashboardsForModel(
    runtimeInstanceId,
    modelName
  );

  function formatModelName(str) {
    return str?.trim().replaceAll(" ", "_").replace(/\.sql/, "");
  }

  const onChangeCallback = async (e) => {
    if (!e.target.value.match(/^[a-zA-Z_][a-zA-Z0-9_]*$/)) {
      notifications.send({
        message:
          "Model name must start with a letter or underscore and contain only letters, numbers, and underscores",
      });
      e.target.value = modelName; // resets the input
      return;
    }
    if (isDuplicateName(e.target.value, modelName, $allNamesQuery.data)) {
      notifications.send({
        message: `Name ${e.target.value} is already in use`,
      });
      e.target.value = modelName; // resets the input
      return;
    }

    try {
      const toName = e.target.value;
      const entityType = EntityType.Model;
      await renameFileArtifact(
        queryClient,
        runtimeInstanceId,
        modelName,
        toName,
        entityType,
        $renameModel
      );
      goto(getRouteFromName(toName, entityType), {
        replaceState: true,
      });
    } catch (err) {
      console.error(err.response.data.message);
    }
  };

  $: titleInput = modelName;
</script>

<WorkspaceHeader
  let:width
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
  appRunning={$appQueryStatusStore}
>
  <svelte:fragment slot="workspace-controls">
    <IconButton
      on:click={() => {
        outputLayout.update((state) => {
          state.visible = !state.visible;
          return state;
        });
      }}
      ><span class="text-gray-500"><HideBottomPane size="18px" /></span>
      <svelte:fragment slot="tooltip-content">
        <SlidingWords active={$outputLayout?.visible} reverse
          >results preview</SlidingWords
        >
      </svelte:fragment>
    </IconButton>
  </svelte:fragment>
  <svelte:fragment slot="cta">
    {@const collapse = width < 800}
    <PanelCTA side="right">
      <ModelWorkspaceCTAs
        availableDashboards={$availableDashboards?.data}
        suppressTooltips={contextMenuOpen}
        {modelName}
        {collapse}
        {modelHasError}
      />
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>
