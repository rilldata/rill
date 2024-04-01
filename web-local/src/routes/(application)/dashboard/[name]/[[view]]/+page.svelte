<script lang="ts">
  import { page } from "$app/stores";
  import { Dashboard } from "@rilldata/web-common/features/dashboards";

  $: metricViewName = $page.params.name;
</script>

<Dashboard {metricViewName} />
