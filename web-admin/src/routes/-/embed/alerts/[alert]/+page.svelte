<script lang="ts">
  import { page } from "$app/state";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import AlertHistoryTable from "@rilldata/web-admin/features/alerts/history/AlertHistoryTable.svelte";
  import AlertMetadata from "@rilldata/web-admin/features/alerts/metadata/AlertMetadata.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-common/features/embeds/selectors.ts";

  let alert = $derived(page.params.alert);
  const runtimeClient = useRuntimeClient();
  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<!-- TODO: loading and error states -->
{#if organization && project}
  <ContentContainer>
    <div class="flex justify-center">
      <div class="w-[960px] flex flex-col items-start gap-y-9">
        <AlertMetadata {organization} {project} {alert} />
        <AlertHistoryTable {alert} />
      </div>
    </div>
  </ContentContainer>
{/if}
