<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import AlertsTable from "@rilldata/web-admin/features/alerts/listing/AlertsTable.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<ContentContainer title="Project alerts">
  <div class="flex flex-col items-center gap-y-4">
    <AlertsTable {organization} {project} />
  </div>
</ContentContainer>
