<script lang="ts">
  import { useAlertOwnerName } from "../../alerts/selectors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let ownerId: string;
  // Set for alerts created from an embed, which have no Rill owner.
  export let embedOwnerId: string | undefined = undefined;

  const ownerName = useAlertOwnerName(organization, project, ownerId);
</script>

{#if embedOwnerId}
  <span>{m.alert_created_from_embed()}</span>
{:else if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? m.alert_created_by({ name: $ownerName.data })
      : m.alert_created_through_code()}
  </span>
{/if}
