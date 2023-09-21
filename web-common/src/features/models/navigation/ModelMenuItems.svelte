<script lang="ts">
  import { goto } from "$app/navigation";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import {
    createConnectorServiceOLAPGetTable,
    createRuntimeServicePutFile,
    V1ModelV2,
  } from "@rilldata/web-common/runtime-client";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { getName } from "../../entity-management/name-utils";
  import { generateDashboardYAMLForModel } from "../../metrics-views/metrics-internal-store";
  import { useModel, useModelFileNames } from "../selectors";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;

  const dispatch = createEventDispatcher();

  const createFileMutation = createRuntimeServicePutFile();

  $: modelNames = useModelFileNames($runtime.instanceId);
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);
  $: modelQuery = useModel($runtime.instanceId, modelName);
  let model: V1ModelV2;
  $: model = $modelQuery.data?.model;
  $: modelHasError = !$modelQuery.data?.meta?.reconcileError;

  $: modelSchema = createConnectorServiceOLAPGetTable({
    instanceId: $runtime.instanceId,
    table: model?.state?.table,
    connector: model?.state?.connector,
  });

  const createDashboardFromModel = (modelName: string) => {
    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    const dashboardYAML = generateDashboardYAMLForModel(
      modelName,
      $modelSchema.data?.schema,
      newDashboardName
    );
    $createFileMutation.mutate(
      {
        instanceId: $runtime.instanceId,
        path: getFileAPIPathFromNameAndType(
          newDashboardName,
          EntityType.MetricsDefinition
        ),
        data: {
          blob: dashboardYAML,
          create: true,
          createOnly: true,
        },
      },
      {
        onSuccess: () => {
          goto(`/dashboard/${newDashboardName}`);
          const previousActiveEntity = $appScreen?.type;
          behaviourEvent.fireNavigationEvent(
            newDashboardName,
            BehaviourEventMedium.Menu,
            MetricsEventSpace.LeftPanel,
            previousActiveEntity,
            MetricsEventScreenName.Dashboard
          );
        },
        onError: (err) => {
          console.error(err);
        },
        onSettled: () => {
          overlay.set(null);
          toggleMenu(); // unmount component
        },
      }
    );
  };

  const handleDeleteModel = async (modelName: string) => {
    await deleteFileArtifact(
      $runtime.instanceId,
      modelName,
      EntityType.Model,
      $modelNames.data
    );
    toggleMenu();
  };
</script>

<MenuItem
  disabled={modelHasError}
  icon
  on:select={() => createDashboardFromModel(modelName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if modelHasError}
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
