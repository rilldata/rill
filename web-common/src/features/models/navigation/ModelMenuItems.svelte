<script lang="ts">
  import { goto } from "$app/navigation";
  import Cancel from "@rilldata/web-common/components/icons/Cancel.svelte";
  import EditIcon from "@rilldata/web-common/components/icons/EditIcon.svelte";
  import Explore from "@rilldata/web-common/components/icons/Explore.svelte";
  import { Divider, MenuItem } from "@rilldata/web-common/components/menu";
  import { useDashboardFileNames } from "@rilldata/web-common/features/dashboards/selectors";
  import {
    getFileAPIPathFromNameAndType,
    getFilePathFromNameAndType,
  } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { useSchemaForTable } from "@rilldata/web-common/features/entity-management/resource-selectors";
  import { waitForResource } from "@rilldata/web-common/features/entity-management/resource-status-utils";
  import { getFileHasErrors } from "@rilldata/web-common/features/entity-management/resources-store";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { appScreen } from "@rilldata/web-common/layout/app-store";
  import { overlay } from "@rilldata/web-common/layout/overlay-store";
  import { waitUntil } from "@rilldata/web-common/lib/waitUtils";
  import { behaviourEvent } from "@rilldata/web-common/metrics/initMetrics";
  import { BehaviourEventMedium } from "@rilldata/web-common/metrics/service/BehaviourEventTypes";
  import {
    MetricsEventScreenName,
    MetricsEventSpace,
  } from "@rilldata/web-common/metrics/service/MetricsTypes";
  import { createRuntimeServicePutFile } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { runtime } from "../../../runtime-client/runtime-store";
  import { deleteFileArtifact } from "../../entity-management/actions";
  import { getName } from "../../entity-management/name-utils";
  import { generateDashboardYAMLForModel } from "../../metrics-views/metrics-internal-store";
  import { useModel, useModelFileNames } from "../selectors";

  export let modelName: string;
  // manually toggle menu to workaround: https://stackoverflow.com/questions/70662482/react-query-mutate-onsuccess-function-not-responding
  export let toggleMenu: () => void;
  $: modelPath = getFilePathFromNameAndType(modelName, EntityType.Model);

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  const createFileMutation = createRuntimeServicePutFile();

  $: modelNames = useModelFileNames($runtime.instanceId);
  $: dashboardNames = useDashboardFileNames($runtime.instanceId);
  $: modelQuery = useModel($runtime.instanceId, modelName);
  $: modelHasError = getFileHasErrors(
    queryClient,
    $runtime.instanceId,
    modelPath
  );

  $: modelSchema = useSchemaForTable(
    $runtime.instanceId,
    $modelQuery.data?.model
  );

  const createDashboardFromModel = async (modelName: string) => {
    if (!$modelQuery.data?.model) {
      return;
    }

    overlay.set({
      title: "Creating a dashboard for " + modelName,
    });
    const newDashboardName = getName(
      `${modelName}_dashboard`,
      $dashboardNames.data
    );
    await waitUntil(() => !!$modelSchema.data?.schema);
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
        onSuccess: async () => {
          await waitForResource(
            queryClient,
            $runtime.instanceId,
            getFilePathFromNameAndType(
              newDashboardName,
              EntityType.MetricsDefinition
            )
          );
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
  disabled={$modelHasError}
  icon
  on:select={() => createDashboardFromModel(modelName)}
  propogateSelect={false}
>
  <Explore slot="icon" />
  Autogenerate dashboard
  <svelte:fragment slot="description">
    {#if $modelHasError}
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
