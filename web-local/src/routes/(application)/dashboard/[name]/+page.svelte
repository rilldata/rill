<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";
  import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/dashboard-stores";
  import { fromUrl } from "@rilldata/web-common/features/dashboards/proto-state/fromProto";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { WorkspaceContainer } from "@rilldata/web-common/layout/workspace";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { featureFlags } from "../../../../lib/application-state-stores/application-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";

  $: metricViewName = $page.params.name;
  $: metricsExplorer = $metricsExplorerStore.entities[metricViewName];

  let protoState: string;
  let urlState: string;
  let updating = false;

  function handleStateChange() {
    if (protoState === metricsExplorer.proto) return;
    protoState = metricsExplorer.proto;

    // if state didn't change do not call goto. this avoids adding unnecessary urls to history stack
    if (protoState !== urlState) {
      goto(`/dashboard/${metricViewName}/?state=${protoState}`);
      updating = true;
    }
  }

  function handleUrlChange() {
    const newUrlState = $page.url.searchParams.get("state");
    if (urlState === newUrlState) return;
    urlState = newUrlState;

    // run sync if we didn't change the url through a state change
    // this can happen when url is updated directly by the user
    if (!updating && urlState && urlState !== protoState) {
      const partialDashboard = fromUrl($page.url);
      if (partialDashboard) {
        metricsExplorerStore.syncFromUrl(metricViewName, partialDashboard);
      }
    }
    updating = false;
  }

  $: if (metricsExplorer) {
    handleStateChange();
  }
  $: if ($page) {
    handleUrlChange();
  }

  $: fileQuery = useRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricViewName, EntityType.MetricsDefinition),
    {
      query: {
        onError: (err) => {
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            throw error(404, "Dashboard not found");
          }

          throw error(err.response?.status || 500, err.message);
        },
      },
    }
  );

  $: catalogQuery = useRuntimeServiceGetCatalogEntry(
    $runtime.instanceId,
    metricViewName,
    {
      query: {
        onError: () => {
          if (!metricViewName) return;

          // When the catalog entry doesn't exist, the dashboard config is invalid
          if ($featureFlags.readOnly) {
            throw error(400, "Invalid dashboard");
          }

          goto(`/dashboard/${metricViewName}/edit`);
        },
      },
    }
  );
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && $catalogQuery.data}
  <WorkspaceContainer
    top="0px"
    assetID={metricViewName}
    bgClass="bg-white"
    inspector={false}
  >
    <Dashboard {metricViewName} slot="body" />
  </WorkspaceContainer>
{/if}
