<script lang="ts">
  import { useReportOwnerName } from "../selectors";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const ownerName = useReportOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data || "Via code"}
  </span>
{/if}
