<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import ReportsTable from "@rilldata/web-admin/features/scheduled-reports/listing/ReportsTable.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ContentContainer title="Project reports">
  <div class="flex flex-col items-center gap-y-4">
    <ReportsTable {organization} {project} />
  </div>
</ContentContainer>
