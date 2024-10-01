<script lang="ts">
  import { page } from "$app/stores";
  import Payment from "@rilldata/web-admin/features/billing/Payment.svelte";
  import Plan from "@rilldata/web-admin/features/billing/plans/Plan.svelte";

  $: organization = $page.params.organization;
</script>

<div class="flex flex-col gap-y-5">
  <Plan {organization} />
  <Payment {organization} />
</div>
