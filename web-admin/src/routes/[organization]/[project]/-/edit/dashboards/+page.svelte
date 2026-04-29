<script lang="ts">
  import ContentContainer from "@rilldata/web-common/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
</script>

<ContentContainer maxWidth={900} title="Dashboards">
  <DashboardsTable isPreview />
</ContentContainer>
