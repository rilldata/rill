<script lang="ts">
  import { goto } from "$app/navigation";
  import { page } from "$app/stores";
  import File from "../../components/icons/File.svelte";

  export let filePath: string;

  $: fileName = filePath.split("/").pop();
  $: leftPadding = getLeftPaddingForDirectoryLevel(
    getDirectoryLevelFromPath(filePath),
  );
  $: isCurrentFile = filePath === $page.params.file;

  async function navigate(filePath: string) {
    await goto(`/files/${filePath}`);
  }

  function getDirectoryLevelFromPath(path: string) {
    return path.split("/").length;
  }

  function getLeftPaddingForDirectoryLevel(dirLevel: number) {
    return 4 + (dirLevel - 1) * 4;
  }

  $: if (leftPadding === 2) {
    console.log("filePath", filePath, "leftPadding", leftPadding);
  }

  $: console.log("filePath", filePath, "leftPadding", leftPadding);
</script>

<button
  on:click={() => navigate(filePath)}
  class="w-full pl-{leftPadding} pr-4 text-left py-1 flex gap-x-1 items-center text-gray-900 font-medium hover:text-gray-900 hover:bg-slate-100 {isCurrentFile
    ? 'bg-slate-100 text-gray-900'
    : ''}"
>
  <File size="14px" className="shrink-0" />
  <span class="truncate">{fileName}</span>
</button>
