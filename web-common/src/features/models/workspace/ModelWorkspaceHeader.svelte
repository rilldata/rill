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
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import type { LayoutElement } from "../../../layout/workspace/types";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import { renameFileArtifact } from "../../entity-management/actions";
  import {
    getFilePathFromNameAndType,
    getRouteFromName,
  } from "../../entity-management/entity-mappers";
  import { isDuplicateName } from "../../entity-management/name-utils";
  import ModelWorkspaceCTAs from "./ModelWorkspaceCTAs.svelte";

  export let modelName: string;

  const queryClient = useQueryClient();

  $: runtimeInstanceId = $runtime.instanceId;

  $: allNamesQuery = useAllNames(runtimeInstanceId);

  const outputLayout = getContext(
    "rill:app:output-layout",
  ) as Writable<LayoutElement>;
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
    if (
      isDuplicateName(e.target.value, modelName, $allNamesQuery?.data ?? [])
    ) {
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
        runtimeInstanceId,
        modelName,
        toName,
        entityType,
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
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
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
  <svelte:fragment slot="cta" let:width>
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
