<script lang="ts">
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { createFileDeleter } from "@rilldata/web-common/features/entity-management/file-actions";
  import {
    useDashboardNames,
    useModel,
    useModelNames,
  } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { createDashboardFromModelCreator } from "@rilldata/web-common/features/models/createDashboardFromModel";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { getLeftPanelParams } from "@rilldata/web-common/metrics/service/metrics-helpers";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const dispatch = createEventDispatcher();

  $: modelNames = useModelNames($runtime.instanceId);
  $: dashboardNames = useDashboardNames($runtime.instanceId);
  $: fileDeleter = createFileDeleter(modelNames);
  $: dashboardFromModelCreator = createDashboardFromModelCreator(
    dashboardNames,
    getLeftPanelParams()
  );

  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: hasError = !!$modelQuery.data.meta.reconcileError;

  const createDashboardFromModel = async (modelName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    await dashboardFromModelCreator($modelQuery.data, modelName);

    // TODO: should this wait till everything is finished?
    overlay.set(null);
    toggleMenu(); // unmount component
  };

  const handleDeleteModel = async (modelName: string) => {
    await fileDeleter(
      modelName,
      EntityType.Model,
      getFilePathFromNameAndType(modelName, EntityType.Model)
    );
    toggleMenu();
  };
</script>

<MenuItem
  disabled={hasError}
  icon
  on:select={() => createDashboardFromModel(modelName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if hasError}
      Model has errors
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
