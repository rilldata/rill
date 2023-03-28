<script lang="ts">
  import { page } from "$app/stores";
  import { getFilePathFromNameAndType } from "@rilldata/web-common/features/entity-management/entity-mappers";
  import { EntityType } from "@rilldata/web-common/features/entity-management/types";
  import { MetricsWorkspace } from "@rilldata/web-common/features/metrics-views";
  import { useRuntimeServiceGetFile } from "@rilldata/web-common/runtime-client";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
  import { error } from "@sveltejs/kit";
  import { onMount } from "svelte";
  import { featureFlags } from "../../../../../lib/application-state-stores/application-store";
  import { CATALOG_ENTRY_NOT_FOUND } from "../../../../../lib/errors/messages";

  $: metricViewName = $page.params.name;

  onMount(() => {
    if ($featureFlags.readOnly) {
      throw error(404, "Page not found");
    }
  });

  $: fileQuery = useRuntimeServiceGetFile(
    $runtime.instanceId,
    getFilePathFromNameAndType(metricViewName, EntityType.MetricsDefinition),
    {
      query: {
        onError: (err) => {
          if (err.response?.data?.message.includes(CATALOG_ENTRY_NOT_FOUND)) {
            throw error(404, "Dashboard not found");
          }

          //throw error(err.response?.status || 500, err.message);
        },
      },
    }
  );

  $: yaml = $fileQuery.data?.blob || "";

  let nonStandardError: string | undefined;
</script>

<svelte:head>
  <title>Rill Developer | {metricViewName}</title>
</svelte:head>

{#if $fileQuery.data && yaml !== undefined}
  <MetricsWorkspace metricsDefName={metricViewName} {nonStandardError} {yaml} />
{/if}
