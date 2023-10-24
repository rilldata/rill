<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import VerticalScrollContainer from "@rilldata/web-common/layout/VerticalScrollContainer.svelte";
  import ReportsTable from "../../../../../features/scheduled-reports/ReportsTable.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<VerticalScrollContainer>
  <ContentContainer>
    <div class="flex flex-col items-center">
      <ReportsTable {organization} {project} />
    </div>
  </ContentContainer>
</VerticalScrollContainer>
