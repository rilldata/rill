<script lang="ts">
  import type { RestoreChangesBlock } from "@rilldata/web-common/features/chat/core/messages/restore/restore-block.ts";

  export let block: RestoreChangesBlock;
</script>

<div>Restored last {block.revertedConversations.length} conversation(s).</div>
