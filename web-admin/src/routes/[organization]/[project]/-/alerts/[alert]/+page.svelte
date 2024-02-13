<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ReportHistoryTable from "@rilldata/web-admin/features/scheduled-reports/history/ReportHistoryTable.svelte";
  import ReportMetadata from "@rilldata/web-admin/features/scheduled-reports/metadata/ReportMetadata.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alert = $page.params.alert;
</script>

<ContentContainer>
  <div class="flex justify-center">
    <div class="w-[960px] flex flex-col items-start gap-y-9">
      <ReportMetadata {organization} {project} {report} />
      <ReportHistoryTable {report} />
    </div>
  </div>
</ContentContainer>
