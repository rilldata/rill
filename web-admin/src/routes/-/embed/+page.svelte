<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
  import PersonalCanvasesList from "@rilldata/web-admin/features/personal-files/canvas/PersonalCanvasesList.svelte";
  import { embedPersonalCanvasesEnabled } from "@rilldata/web-admin/features/embeds/embed-personal-canvases.ts";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  const runtimeClient = useRuntimeClient();
  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  $: ({ organization, project } = $orgAndProjectQuery);
</script>

<ContentContainer>
  <div class="flex flex-col gap-y-4">
    {#if $embedPersonalCanvasesEnabled && organization && project}
      <PersonalCanvasesList org={organization} {project} />
    {/if}
    <div class="flex flex-col items-center gap-y-4">
      <DashboardsTable isEmbedded />
    </div>
  </div>
</ContentContainer>
