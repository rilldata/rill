<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

  export let connector: V1ConnectorDriver;
</script>

<div class="pb-1 text-slate-500">
  Need help? Refer to our
  <a
    href={connector.docsUrl || "https://docs.rilldata.com/build/connect/"}
    rel="noreferrer noopener"
    target="_blank">docs</a
  >
  for more information.
</div>
