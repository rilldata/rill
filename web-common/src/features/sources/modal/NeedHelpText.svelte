<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { ExternalLinkIcon } from "lucide-svelte";

  export let connector: V1ConnectorDriver;
</script>

<div>
  <div class="text-sm leading-none font-medium mb-4">Help</div>
  <div class="text-sm leading-normal font-medium text-muted-foreground mb-2">
    Need help connecting to {connector.displayName}? Check out our documentation
    for detailed instructions.
  </div>
  <span class="flex flex-row items-center gap-2 group">
    <a
      href={connector.docsUrl || "https://docs.rilldata.com/build/connectors/"}
      rel="noreferrer noopener"
      target="_blank"
      class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium group-hover:underline break-all"
    >
      How to connect to {connector.displayName}
    </a>
    <ExternalLinkIcon size="16px" color="#6366F1" />
  </span>
</div>
