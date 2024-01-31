<script lang="ts">
  import { useReportOwnerName } from "../selectors";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  $: ownerName = useReportOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? `Report created by ${$ownerName.data}`
      : "Report created through code"} •
  </span>
{/if}
