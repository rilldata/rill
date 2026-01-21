<script lang="ts">
  import MappingWorkspace from "@rilldata/web-common/features/workspaces/MappingWorkspace.svelte";
  import DeveloperChat from "@rilldata/web-common/features/chat/DeveloperChat.svelte";
</script>

<svelte:head>
  <title>Rill Developer | New Mapping File</title>
</svelte:head>

<div class="flex h-full overflow-hidden">
  <div class="flex-1 overflow-hidden">
    <MappingWorkspace />
  </div>
  <DeveloperChat />
</div>
