<script lang="ts">
  import { page } from "$app/stores";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";

  $: organization = $page.params.organization;
</script>

<Plan {organization} />
