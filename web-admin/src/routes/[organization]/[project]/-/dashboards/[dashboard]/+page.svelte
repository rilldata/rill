<script lang="ts">
  import { page } from "$app/stores";

  $: customDashboardName = $page.params.dashboard;
</script>

<div>
  A "read-only" version of the Custom Dashboard named <span
    class="font-semibold">{customDashboardName}</span
  > goes here.
</div>
