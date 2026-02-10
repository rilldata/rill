<script lang="ts">
  import { useAlertOwnerName } from "../../alerts/selectors";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const ownerName = useAlertOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data || "Via code"}
  </span>
{/if}
