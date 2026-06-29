<script lang="ts">
  import { useAlertOwnerName } from "../../alerts/selectors";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const ownerName = useAlertOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? m.alert_created_by({ name: $ownerName.data })
      : m.alert_created_through_code()}
  </span>
{/if}
