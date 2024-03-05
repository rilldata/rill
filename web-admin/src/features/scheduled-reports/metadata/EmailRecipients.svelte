<script lang="ts">
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import { Tag } from "@rilldata/web-common/components/tag/index.js";

  export let emailRecipients: string[];
</script>

<div class="flex flex-col gap-y-3">
  <MetadataLabel>Recipients ({emailRecipients.length})</MetadataLabel>
  <div class="flex flex-wrap gap-2">
    {#each emailRecipients as recipient}
      <Tag>
        {recipient}
      </Tag>
    {/each}
  </div>
</div>
