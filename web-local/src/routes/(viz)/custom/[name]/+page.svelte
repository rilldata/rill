<script lang="ts">
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";

  export let data;

  $: ({ components = [], columns, gap } = data.dashboard);
</script>

<CustomDashboardEmbed {components} {columns} {gap} />
