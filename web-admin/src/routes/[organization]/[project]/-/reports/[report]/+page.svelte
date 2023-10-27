<script lang="ts">
  import { page } from "$app/stores";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ContentContainer from "../../../../../../components/layout/ContentContainer.svelte";
  import ReportHistoryTable from "../../../../../../features/scheduled-reports/history/ReportHistoryTable.svelte";
  import ReportMetadata from "../../../../../../features/scheduled-reports/metadata/ReportMetadata.svelte";

  $: report = $page.params.report;
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex justify-center">
      <div class="w-[960px] flex flex-col items-start gap-y-8">
        <ReportMetadata {report} />
        <ReportHistoryTable {report} />
      </div>
    </div>
  </ContentContainer>
</VerticalScrollContainer>
