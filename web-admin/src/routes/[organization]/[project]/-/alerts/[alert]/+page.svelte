<script lang="ts">
  import { page } from "$app/stores";
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import AlertHistoryTable from "@rilldata/web-admin/features/alerts/history/AlertHistoryTable.svelte";
  import AlertMetadata from "@rilldata/web-admin/features/alerts/metadata/AlertMetadata.svelte";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
  $: alert = $page.params.alert;
</script>

<ContentContainer maxWidth={1100}>
  <div class=" flex flex-col items-start gap-y-9">
    <AlertMetadata {alert} {organization} {project} />
    <AlertHistoryTable {alert} />
  </div>
</ContentContainer>
