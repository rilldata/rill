<script lang="ts">
  import { createAdminServiceGetPaymentsPortalURL } from "@rilldata/web-admin/client";
  import SettingsItemContainer from "@rilldata/web-admin/features/settings/SettingsItemContainer.svelte";
  import { Button } from "@rilldata/web-common/components/button";

  export let organization: string;

  $: paymentUrl = createAdminServiceGetPaymentsPortalURL(organization);
</script>

<SettingsItemContainer title="Payment Method">
  <div slot="description">Your payment method is valid and good to go.</div>
  <Button slot="action" type="link" href={$paymentUrl.data?.url}>Manage</Button>
</SettingsItemContainer>
