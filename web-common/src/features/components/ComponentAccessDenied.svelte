<script lang="ts">
  import { Lock } from "lucide-svelte";
</script>

<div class="component-access-denied flex flex-col gap-y-2 size-full items-center justify-center p-4 text-fg-muted bg-surface-secondary/50">
  <div class="p-1.5 bg-surface-secondary rounded-full">
    <Lock size="12px" />
  </div>
  <div class="text-xs font-medium text-center max-w-md">
    You don't have access to this component
  </div>
</div>
