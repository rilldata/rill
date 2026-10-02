<script lang="ts">
  import { page } from "$app/state";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-admin/features/embeds/selectors.ts";
  import AlertMetadata from "@rilldata/web-admin/features/alerts/metadata/AlertMetadata.svelte";
  import AlertHistoryTable from "@rilldata/web-admin/features/alerts/history/AlertHistoryTable.svelte";

  let alert = $derived(page.params.alert);
  const runtimeClient = useRuntimeClient();
  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<!-- TODO: loading and error states -->
{#if organization && project}
  <ContentContainer maxWidth={1100}>
    <div class=" flex flex-col items-start gap-y-9">
      <AlertMetadata {alert} {organization} {project} />
      <AlertHistoryTable {alert} />
    </div>
  </ContentContainer>
{/if}
