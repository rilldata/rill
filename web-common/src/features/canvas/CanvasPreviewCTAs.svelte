<script lang="ts">
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import { Button } from "../../components/button";
  import { runtime } from "../../runtime-client/runtime-store";
  import { featureFlags } from "../feature-flags";

  export let canvasName: string;

  $: ({ instanceId } = $runtime);

  $: canvasQuery = useCanvas(instanceId, canvasName);
  $: canvasFilePath = $canvasQuery.data?.filePath ?? "";

  const { readOnly } = featureFlags;
</script>

{#if !$readOnly}
  <div class="flex gap-2 flex-shrink-0 ml-auto">
    <Button type="secondary" href={`/files${canvasFilePath}`}>Edit</Button>
  </div>
{/if}
