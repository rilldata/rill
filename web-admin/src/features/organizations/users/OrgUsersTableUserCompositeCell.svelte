<script lang="ts">
  export let name: string;
  export let email: string;
</script>

<div class="flex flex-col text-left">
  <span class="text-sm font-medium text-gray-900">{name}</span>
  <span class="text-xs text-gray-500">{email}</span>
</div>
