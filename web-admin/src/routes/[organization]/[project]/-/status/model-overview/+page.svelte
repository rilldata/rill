<script lang="ts">
  import { page } from "$app/stores";

  $: organization = $page.params.organization;
  $: project = $page.params.project;
</script>

<div class="flex flex-col gap-y-4 size-full">
  <h2 class="text-lg font-semibold text-gray-900">Model Overview</h2>
  <p class="text-gray-500">Coming soon...</p>
</div>
