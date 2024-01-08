<script lang="ts">
  import {_} from "svelte-i18n";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    useCreateDashboardFromModelUIAction,
    useModelSchemaIsReady,
  } from "@rilldata/web-common/features/models/createDashboardFromModel";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { useModelFileNames } from "../selectors";

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
    modelPath
  );
  $: modelSchemaIsReady = useModelSchemaIsReady(
    queryClient,
    $runtime.instanceId,
    modelName
  );
  $: disableCreateDashboard = $modelHasError || !$modelSchemaIsReady;

  $: createDashboardFromModel = useCreateDashboardFromModelUIAction(
    $runtime.instanceId,
    modelName,
    queryClient,
    BehaviourEventMedium.Menu,
    MetricsEventSpace.LeftPanel
  );

  async function createDashboardFromModelHandler() {
    await createDashboardFromModel();
    toggleMenu();
  }

  const handleDeleteModel = async (modelName: string) => {
    if ($modelNames.data) {
      await deleteFileArtifact(
        $runtime.instanceId,
        modelName,
        EntityType.Model,
        $modelNames.data
      );
    }
    toggleMenu();
  };
</script>

<MenuItem
  disabled={disableCreateDashboard}
  icon
  on:select={() => createDashboardFromModelHandler()}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if $modelHasError}
      {$_('model-has-errors')}
    {:else if !$modelSchemaIsReady}
      {$_('dependencies-are-being-reconciled')}
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
  {$_("rename")}
</MenuItem>
<MenuItem
  icon
  on:select={() => handleDeleteModel(modelName)}
  propogateSelect={false}
>
  <Cancel slot="icon" />
  {$_("delete")}
</MenuItem>
