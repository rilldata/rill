<script lang="ts">
  import * as Tooltip from "@rilldata/web-common/components/tooltip-v2";
  import { KeyIcon } from "lucide-svelte";

  export let open = false;
</script>

<Tooltip.Root portal="body" {open}>
  <Tooltip.Trigger>
    <slot />
  </Tooltip.Trigger>
  <Tooltip.Content side="right" sideOffset={12}>
    <div class="flex items-center gap-x-1">
      <KeyIcon size={12} />
      <span>Last used</span>
    </div>
  </Tooltip.Content>
</Tooltip.Root>
