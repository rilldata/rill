<script lang="ts">
  import type { PageData } from "./$types";
  import FileAndResourceWatcher from "@rilldata/web-common/features/entity-management/FileAndResourceWatcher.svelte";

  export let data: PageData;
  $: ({ runtime } = data);
  $: runtimeKey = `${runtime.host ?? ""}__${runtime.instanceId ?? ""}`;
</script>

{#key runtimeKey}
  <FileAndResourceWatcher host={runtime.host} instanceId={runtime.instanceId}>
    <div class="mx-auto my-auto">
      <slot />
    </div>
  </FileAndResourceWatcher>
{/key}
