<script lang="ts">
  import { useAlertOwnerName } from "../selectors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let ownerId: string | undefined = undefined;
  export let ownerEmail: string | undefined = undefined;

  $: ownerNameQuery = useAlertOwnerName(organization, project, ownerId);
  $: ownerName = $ownerNameQuery.data ?? ownerEmail;
  $: isSuccess = ownerId ? $ownerNameQuery.isSuccess : true;
</script>

{#if isSuccess}
  <span>
    {ownerName
      ? m.alert_meta_created_by({ name: ownerName })
      : m.alert_meta_created_through_code()}
  </span>
{/if}
