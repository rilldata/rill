<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { WandIcon } from "lucide-svelte";
  import { createEventDispatcher } from "svelte";
  import { V1ReconcileStatus } from "../../../runtime-client";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { useCreateDashboardFromTableUIAction } from "../../metrics-views/ai-generation/generateMetricsView";
  import { useModel, useModelFileNames } from "../selectors";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  $: modelNames = useModelFileNames($runtime.instanceId);
  $: modelHasError = getFileHasErrors(
    queryClient,
    $runtime.instanceId,
    modelPath,
  );
  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: modelIsIdle =
    $modelQuery.data?.meta?.reconcileStatus ===
    V1ReconcileStatus.RECONCILE_STATUS_IDLE;
  $: disableCreateDashboard = $modelHasError || !modelIsIdle;

  $: createDashboardFromTable = useCreateDashboardFromTableUIAction(
    $runtime.instanceId,
    modelName,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel,
  );

  const handleDeleteModel = async (modelName: string) => {
    if ($modelNames.data) {
      await deleteFileArtifact(
        $runtime.instanceId,
        modelName,
        EntityType.Model,
        $modelNames.data,
      );
    }
    toggleMenu();
  };
</script>

<MenuItem
  disabled={disableCreateDashboard}
  icon
  on:select={() => {
    void createDashboardFromTable();
    toggleMenu();
  }}
  propogateSelect={false}
>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate dashboard with AI
    <WandIcon class="w-3 h-3" />
  </div>
  <svelte:fragment slot="description">
    {#if $modelHasError}
      Model has errors
    {:else if !modelIsIdle}
      Dependencies are being reconciled.
    {/if}
  </svelte:fragment>
</MenuItem>
<MenuItem
  disabled={disableCreateDashboard}
  icon
  on:select={() => {
    dispatch("generate-chart", {
      table: $modelQuery.data?.model?.state?.table,
      connector: $modelQuery.data?.model?.state?.connector,
    });
    toggleMenu();
  }}
  propogateSelect={false}
>
  <Explore slot="icon" />
  <div class="flex gap-x-2 items-center">
    Generate chart with AI
    <WandIcon class="w-3 h-3" />
  </div>
  <svelte:fragment slot="description">
    {#if $modelHasError}
      Model has errors
    {:else if !modelIsIdle}
      Dependencies are being reconciled.
    {/if}
  </svelte:fragment>
</MenuItem>
<Divider />
<MenuItem
  icon
  on:select={() => {
    dispatch("rename-asset");
  }}
>
  <EditIcon slot="icon" />
  Rename...
</MenuItem>
<MenuItem
  icon
  on:select={() => handleDeleteModel(modelName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  Delete
</MenuItem>
