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
    <ReportsTable {organization} {project} />
  </ContentContainer>
</VerticalScrollContainer>
