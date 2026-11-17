<script lang="ts">
  import AlertOpen from "@rilldata/web-admin/features/alerts/open/AlertOpen.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
</script>

<!-- Embeds have no magic token, so the alert's filters must be applied by the dashboard. -->
<AlertOpen
  alertSpec={data.alertSpec}
  exploreName={data.exploreName}
  executionTime={data.executionTime}
  pathContext={{ embedded: true }}
  ignoreFilters={false}
  backHref={`/-/embed/alerts/${data.alertId}`}
/>
