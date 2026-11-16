<script lang="ts">
  import AlertOpen from "@rilldata/web-admin/features/alerts/open/AlertOpen.svelte";
  import type { PageData } from "./$types";

  let { data }: { data: PageData } = $props();
</script>

<AlertOpen
  alertSpec={data.alertSpec}
  exploreName={data.exploreName}
  executionTime={data.executionTime}
  pathContext={{
    organization: data.organization,
    project: data.project,
    token: data.token,
  }}
  ignoreFilters={!!data.token}
  backHref={`/${data.organization}/${data.project}/-/alerts/${data.alertId}`}
/>
