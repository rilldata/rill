<script lang="ts">
  import { useReportOwnerName } from "../selectors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let ownerId: string;

  const ownerName = useReportOwnerName(organization, project, ownerId);
</script>

{#if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? m.report_created_by({ name: $ownerName.data })
      : m.report_created_through_code()}
  </span>
{/if}
