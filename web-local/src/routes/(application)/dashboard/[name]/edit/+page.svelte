<script lang="ts">
  import { page } from "$app/stores";
  import { initLocalUserPreferenceStore } from "@rilldata/web-common/features/dashboards/user-preferences";
  import { getFileAPIPathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { featureFlags } from "@rilldata/web-common/features/feature-flags";
  import { MetricsWorkspace } from "@rilldata/web-common/features/metrics-views";
  import { createRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";
  import DeployDashboardCta from "@rilldata/web-common/features/dashboards/workspace/DeployDashboardCTA.svelte";

  let showDeployDashboardModal = false;
  $: metricViewName = $page.params.name;
  $: filePath = getFileAPIPathFromNameAndType(
    metricViewName,
    EntityType.MetricsDefinition,
  );

  const { readOnly } = featureFlags;

  onMount(() => {
    if ($readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: fileQuery = createRuntimeServiceGetFile($runtime.instanceId, filePath, {
    query: {
      onError: (err) => {
        if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
          throw error(404, "Dashboard not found");
        }

        throw error(err.response?.status || 500, err.message);
      },
      // this will ensure that any changes done outside our app is pulled in.
      refetchOnWindowFocus: true,
    },
  });

  $: yaml = $fileQuery.data?.blob || "";

  $: initLocalUserPreferenceStore(metricViewName);
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <MetricsWorkspace {filePath} />
{/if}

<DeployDashboardCta
  on:close={() => (showDeployDashboardModal = false)}
  open={showDeployDashboardModal}
/>
