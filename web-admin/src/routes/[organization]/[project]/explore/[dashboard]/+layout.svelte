<script lang="ts">
  import Chat from "@rilldata/web-common/features/chat/Chat.svelte";
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <Chat />
</div>
