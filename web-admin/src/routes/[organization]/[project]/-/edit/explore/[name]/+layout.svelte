<script lang="ts">
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <DeveloperChat />
</div>
