<script lang="ts">
  import ContentContainer from "@rilldata/web-admin/components/layout/ContentContainer.svelte";
  import DashboardsTable from "@rilldata/web-admin/features/dashboards/listing/DashboardsTable.svelte";
</script>

<ContentContainer>
  <div class="flex flex-col items-center gap-y-4">
    <DashboardsTable isEmbedded />
  </div>
</ContentContainer>
