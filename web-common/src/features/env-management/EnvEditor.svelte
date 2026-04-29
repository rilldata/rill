<script lang="ts">
  import { page } from "$app/state";

  let { organization, project } = $derived(page.params);
</script>

<div class="flex h-full w-full items-center justify-center p-8">
  <p class="text-sm text-gray-600">
    You cannot edit .env. Visit <a
      href="/{organization}/{project}/-/settings/environment-variables"
      class="text-blue-500 hover:underline">Settings</a
    > to manage your environment variables.
  </p>
</div>
