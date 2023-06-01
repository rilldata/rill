<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";

  export let filePath: string;

  $: fileName = filePath.split("/").pop();
  $: leftPadding = getLeftPaddingFromDirectoryPath(filePath);
  $: isCurrentFile = filePath === $page.params.file;

  function navigate(filePath: string) {
    goto(`/files/${filePath}`);
  }

  function getLeftPaddingFromDirectoryPath(path: string) {
    const dirLevel = path.split("/").length;
    console.log(filePath, "dirLevel", dirLevel);
    return 2 + (dirLevel - 1) * 6;
  }

  // IDEA: Add a tag so the user can navigate directly to the Catalog view
</script>

<button
  on:click={() => navigate(filePath)}
  class="w-full pl-{leftPadding} pr-4 text-left py-0.5 text-gray-500 hover:text-gray-900 hover:bg-gray-100 {isCurrentFile
    ? 'bg-gray-100 text-gray-900'
    : ''}"
>
  {fileName}
</button>
