<script lang="ts">
  import type { V1Resource } from "@rilldata/web-common/runtime-client";

  export let resource: V1Resource;

  function getResourceSpec(res: V1Resource): string {
    const kindKeys = [
      "source",
      "model",
      "metricsView",
      "explore",
      "theme",
      "component",
      "canvas",
      "api",
      "connector",
      "report",
      "alert",
    ] as const;

    for (const key of kindKeys) {
      if (res[key]) {
        return JSON.stringify(res[key], null, 2);
      }
    }

    const { meta: _meta, ...rest } = res;
    return JSON.stringify(rest, null, 2);
  }

  $: specContent = getResourceSpec(resource);
</script>

<pre
  class="text-xs font-mono whitespace-pre-wrap bg-surface-subtle rounded-md p-4 overflow-auto max-h-[50vh]">{specContent}</pre>
