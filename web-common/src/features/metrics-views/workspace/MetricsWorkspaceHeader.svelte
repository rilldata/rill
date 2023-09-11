<script lang="ts">
  import { validateAndRenameEntity } from "@rilldata/web-common/features/entity-management/rename-entity";
  import { useAllNames } from "@rilldata/web-common/features/entity-management/selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createRuntimeServiceRenameFile } from "@rilldata/web-common/runtime-client";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import GoToDashboardButton from "./GoToDashboardButton.svelte";

  export let metricsDefName;

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllNames(runtimeInstanceId);
  const renameMetricsDef = createRuntimeServiceRenameFile();

  const onChangeCallback = async (e) => {
    if (
      !(await validateAndRenameEntity(
        runtimeInstanceId,
        e.target.value,
        metricsDefName,
        $allNamesQuery.data,
        EntityType.MetricsDefinition,
        renameMetricsDef
      ))
    ) {
      e.target.value = metricsDefName; // resets the input
    }
  };

  $: titleInput = metricsDefName;
</script>

<WorkspaceHeader
  {...{ titleInput, onChangeCallback }}
  appRunning={$appQueryStatusStore}
>
  <GoToDashboardButton {metricsDefName} slot="cta" />
</WorkspaceHeader>
