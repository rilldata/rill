<script lang="ts">
  import CustomDashboardEmbed from "@rilldata/web-common/features/custom-dashboards/CustomDashboardEmbed.svelte";
  import type { PageData } from "./$types";

  export let data: PageData;

  $: ({ items = [], columns, gap } = data.dashboard);
</script>

<CustomDashboardEmbed {items} {columns} {gap} />
