<script lang="ts">
  import { createFileValidatorAndRenamer } from "@rilldata/web-common/features/entity-management/rename-entity";
  import { useAllEntityNames } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appQueryStatusStore } from "@rilldata/web-common/runtime-client/application-store";
  import { WorkspaceHeader } from "../../../layout/workspace";
  import { runtime } from "../../../runtime-client/runtime-store";
  import GoToDashboardButton from "./GoToDashboardButton.svelte";

  export let metricsDefName: string;

  $: runtimeInstanceId = $runtime.instanceId;
  $: allNamesQuery = useAllEntityNames(runtimeInstanceId);
  $: fileValidatorAndRenamer = createFileValidatorAndRenamer(allNamesQuery);

  const onChangeCallback = async (e) => {
    if (
      !(await fileValidatorAndRenamer(
        metricsDefName,
        e.target.value,
        EntityType.MetricsDefinition
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
