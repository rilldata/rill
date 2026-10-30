<script lang="ts">
  import { page } from "$app/stores";
  import ResourceErrorAlerts from "@rilldata/web-admin/features/alerts/resource-errors/ResourceErrorAlerts.svelte";
</script>

<ResourceErrorAlerts
  org={$page.params.organization}
  project={$page.params.project}
  initialResource={$page.url.searchParams.get("resource") ?? undefined}
/>
