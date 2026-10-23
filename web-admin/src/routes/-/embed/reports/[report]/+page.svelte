<script lang="ts">
  import { page } from "$app/state";
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import ReportHistoryTable from "@rilldata/web-admin/features/scheduled-reports/history/ReportHistoryTable.svelte";
  import ReportMetadata from "@rilldata/web-admin/features/scheduled-reports/metadata/ReportMetadata.svelte";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { getOrgAndProjectForEmbed } from "@rilldata/web-admin/features/embeds/selectors.ts";

  let report = $derived(page.params.report);
  const runtimeClient = useRuntimeClient();
  const orgAndProjectQuery = getOrgAndProjectForEmbed(runtimeClient);
  let { organization, project } = $derived($orgAndProjectQuery);
</script>

<!-- TODO: loading and error states -->
{#if organization && project}
  <ContentContainer>
    <div class="flex justify-center">
      <div class="w-[960px] flex flex-col items-start gap-y-9">
        <ReportMetadata {organization} {project} {report} />
        <ReportHistoryTable {report} />
      </div>
    </div>
  </ContentContainer>
{/if}
