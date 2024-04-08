<script lang="ts">
  import ChartDisplay from "@rilldata/web-common/features/charts/ChartDisplay.svelte";
  import ChartsEditor from "@rilldata/web-common/features/charts/editor/ChartsEditor.svelte";

  export let filePath: string;
</script>

<div class="flex gap-x-4 w-full">
  <ChartsEditor {filePath} />
  <ChartDisplay {filePath} />
</div>
