<script lang="ts">
  import { WorkspaceHeader } from "@rilldata/web-common/layout/workspace";
  export let chartName: string;

  function onChangeCallback() {
    console.log("edit");
  }
  $: titleInput = chartName;
</script>

<WorkspaceHeader {...{ titleInput, onChangeCallback }}></WorkspaceHeader>
