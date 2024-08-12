<script lang="ts">
  import { page } from "$app/stores";
  import Plan from "@rilldata/web-admin/features/billing/Plan.svelte";

  $: organization = $page.params.organization;
</script>

<Plan {organization} />
