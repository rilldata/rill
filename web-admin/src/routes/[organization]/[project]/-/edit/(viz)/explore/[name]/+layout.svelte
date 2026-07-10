<script lang="ts">
  import DashboardChat from "@rilldata/web-common/features/chat/DashboardChat.svelte";
</script>

<div class="flex flex-1 overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <slot />
  </div>
  <DashboardChat />
</div>
