<script lang="ts">
  import ExploreChat from "@rilldata/web-common/features/chat/ExploreChat.svelte";
</script>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <ExploreChat />
</div>
