<script lang="ts">
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import { Button } from "../../components/button";
  import { runtime } from "../../runtime-client/runtime-store";
  import { featureFlags } from "../feature-flags";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";

  export let canvasName: string;
  export let inPreviewMode = false;

  $: ({ instanceId } = $runtime);

  $: canvasQuery = useCanvas(instanceId, canvasName);
  $: canvasFilePath = $canvasQuery.data?.filePath ?? "";

  const { dashboardChat, readOnly } = featureFlags;
</script>

{#if $dashboardChat}
  <ChatToggle />
{/if}
{#if !$readOnly && !inPreviewMode}
  <div class="flex gap-2 flex-shrink-0 ml-auto">
    <Button type="secondary" href={`/files${canvasFilePath}`}>Edit</Button>
  </div>
{/if}
