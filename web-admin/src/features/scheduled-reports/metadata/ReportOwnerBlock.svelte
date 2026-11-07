<script lang="ts">
  import { useEmbedUserExternalId, useReportOwnerName } from "../selectors";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let project: string;
  export let ownerId = "";
  // Set for reports created from an embed, which are owned by an embed user instead of a Rill user.
  export let embedOwnerId = "";

  $: ownerName = useReportOwnerName(organization, project, ownerId);
  $: embedOwner = useEmbedUserExternalId(organization, project, embedOwnerId);
</script>

{#if embedOwnerId}
  <!-- The embed user may no longer exist, in which case the external user ID can't be shown. -->
  {#if $embedOwner.isSuccess || $embedOwner.isError}
    <span>
      {$embedOwner.data
        ? m.report_meta_created_from_embed_by({ name: $embedOwner.data })
        : m.report_meta_created_from_embed()} •
    </span>
  {/if}
{:else if $ownerName.isSuccess}
  <span>
    {$ownerName.data
      ? m.report_meta_created_by({ name: $ownerName.data })
      : m.report_meta_created_through_code()} •
  </span>
{/if}
