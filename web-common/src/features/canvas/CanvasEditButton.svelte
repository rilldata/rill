<script lang="ts">
  import * as m from "@rilldata/web-common/paraglide/messages.js";
  import { getFileHref } from "@rilldata/web-common/layout/navigation/editor-routing.ts";
  import { Button } from "@rilldata/web-common/components/button/index.ts";
  import { useCanvas } from "@rilldata/web-common/features/canvas/selector.ts";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";

  let { canvasName }: { canvasName: string } = $props();

  const runtimeClient = useRuntimeClient();

  let canvasQuery = $derived(useCanvas(runtimeClient, canvasName));
  let canvasFilePath = $derived($canvasQuery.data?.filePath ?? "");
</script>

<Button type="secondary" href={getFileHref(canvasFilePath)}
  >{m.canvas_edit()}</Button
>
