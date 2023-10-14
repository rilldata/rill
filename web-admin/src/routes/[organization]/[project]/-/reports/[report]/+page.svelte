<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ReportHistoryTable from "../../../../../../features/scheduled-reports/ReportHistoryTable.svelte";
  import ReportMetadata from "../../../../../../features/scheduled-reports/ReportMetadata.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: report = $page.params.report;
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex flex-col gap-y-8">
      <ReportMetadata {organization} {project} {report} />
      <ReportHistoryTable {organization} {project} />
    </div>
  </ContentContainer>
</VerticalScrollContainer>
