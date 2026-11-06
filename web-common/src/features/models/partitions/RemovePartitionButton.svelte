<script lang="ts">
  import { IconButton } from "@rilldata/web-common/components/button";
  import { X } from "lucide-svelte";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let onRemove: () => void;
</script>

<IconButton
  ariaLabel={m.status_backfill_remove()}
  onclick={onRemove}
  rounded
  tooltipLocation="left"
>
  <X size={14} aria-hidden="true" />
  <span slot="tooltip-content">{m.status_backfill_remove()}</span>
</IconButton>
