<script lang="ts">
  import { page } from "$app/stores";
  import { DashboardWorkspace } from "@rilldata/web-common/features/dashboards";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import {
    useRuntimeServiceGetCatalogEntry,
    useRuntimeServiceGetFile,
  } from "@rilldata/web-common/runtime-client";
  import { error, redirect } from "@sveltejs/kit";
  import { featureFlags } from "../../../../lib/application-state-stores/application-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../lib/errors/messages";

  const metricViewName: string = $page.params.name;

  const fileQuery = useRuntimeServiceGetFile(
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

  const catalogQuery = useRuntimeServiceGetCatalogEntry(metricViewName, {
    query: {
      onError: () => {
        // When the catalog entry doesn't exist, the dashboard config is invalid
        if ($featureFlags.readOnly) {
          throw error(400, "Invalid dashboard");
        }

        throw redirect(307, `/dashboard/${metricViewName}/edit`);
      },
    },
  });
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && $catalogQuery.data}
  <DashboardWorkspace {metricViewName} />
{/if}
