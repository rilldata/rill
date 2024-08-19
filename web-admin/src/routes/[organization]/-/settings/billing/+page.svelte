<script lang="ts">
  import { page } from "$app/stores";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";

  $: organization = $page.params.organization;
</script>

<Plan {organization} />
<Payment {organization} />
