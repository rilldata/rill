<script lang="ts">
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
  import { Button } from "../../components/button";
  import { useRuntimeClient } from "../../runtime-client/v2";
  import { featureFlags } from "../feature-flags";
  import ChatToggle from "@rilldata/web-common/features/chat/layouts/sidebar/ChatToggle.svelte";

  const client = useRuntimeClient();

  export let canvasName: string;

  $: canvasQuery = useCanvas(client, canvasName);
  $: canvasFilePath = $canvasQuery.data?.filePath ?? "";

  const { dashboardChat, readOnly } = featureFlags;
</script>

{#if $dashboardChat}
  <ChatToggle />
{/if}
{#if !$readOnly}
  <div class="flex gap-2 flex-shrink-0 ml-auto">
    <Button type="secondary" href={`/files${canvasFilePath}`}>Edit</Button>
  </div>
{/if}
