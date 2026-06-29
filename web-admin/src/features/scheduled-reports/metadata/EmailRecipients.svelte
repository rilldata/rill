<script lang="ts">
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import { Tag } from "@rilldata/web-common/components/tag/index.js";
  import * as m from "@rilldata/web-common/paraglide/messages.js";

  export let emailRecipients: string[];
</script>

<div class="flex flex-col gap-y-3">
  <MetadataLabel>{m.report_recipients({ count: String(emailRecipients.length) })}</MetadataLabel>
  <div class="flex flex-wrap gap-2">
    {#each emailRecipients as recipient}
      <Tag>
        {recipient}
      </Tag>
    {/each}
  </div>
</div>
