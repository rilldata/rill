<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import HideBottomPane from "@rilldata/web-common/components/icons/HideBottomPane.svelte";
  import PanelCTA from "@rilldata/web-common/components/panel/PanelCTA.svelte";
  import SlidingWords from "@rilldata/web-common/components/tooltip/SlidingWords.svelte";
  import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
  import { createFileValidatorAndRenamer } from "@rilldata/web-common/features/entity-management/rename-entity";
  import { useAllEntityNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { getContext } from "svelte";
  import type { Writable } from "svelte/store";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import type { LayoutElement } from "../../../layout/workspace/types";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { useGetDashboardsForModel } from "../../dashboards/selectors";
  import { getFilePathFromNameAndType } from "../../entity-management/entity-mappers";
  import ModelWorkspaceCTAs from "./ModelWorkspaceCTAs.svelte";

  export let modelName: string;

  $: runtimeInstanceId = $runtime.instanceId;

  $: allNamesQuery = useAllEntityNames(runtimeInstanceId);
  $: fileValidatorAndRenamer = createFileValidatorAndRenamer(allNamesQuery);

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
    if (
      !(await fileValidatorAndRenamer(
        e.target.value,
        modelName,
        EntityType.Model
      ))
    ) {
      e.target.value = modelName; // resets the input
    }
  };

  $: titleInput = modelName;
</script>

<WorkspaceHeader
  {...{ titleInput: formatModelName(titleInput), onChangeCallback }}
  appRunning={$appQueryStatusStore}
  let:width
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
        {collapse}
        {modelHasError}
        {modelName}
        suppressTooltips={contextMenuOpen}
      />
    </PanelCTA>
  </svelte:fragment>
</WorkspaceHeader>
