<script lang="ts">
  import { useAlertOwnerName } from "../selectors";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  $: ownerName = useAlertOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? `Alert created by ${$ownerName.data}`
      : "Alert created through code"}
  </span>
{/if}
