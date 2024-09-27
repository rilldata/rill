<script lang="ts">
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";

  export let name: string;
  export let email: string;
</script>

<div class="flex items-center gap-2">
  <Avatar size="h-7 w-7" alt={name} />
  <div class="flex flex-col text-left">
    <span class="text-sm font-medium text-gray-900">{name}</span>
    <span class="text-xs text-gray-500">{email}</span>
  </div>
</div>
