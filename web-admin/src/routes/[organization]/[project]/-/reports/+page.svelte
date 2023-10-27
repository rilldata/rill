<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ReportsTable from "../../../../../features/scheduled-reports/listing/ReportsTable.svelte";
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex flex-col items-center">
      <ReportsTable />
    </div>
  </ContentContainer>
</VerticalScrollContainer>
