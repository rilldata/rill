<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { ExternalLinkIcon } from "lucide-svelte";

  export let connector: V1ConnectorDriver;
</script>

<div>
  <div class="text-sm leading-none font-medium text-fg-secondary mb-4">
    Help
  </div>
  <div class="text-sm leading-normal font-medium text-fg-muted mb-2">
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
  {#if connector.displayName === "DuckDB" || connector.displayName === "SQLite"}
    <div class="mt-8">
      <div class="text-sm leading-none font-medium text-fg-secondary mb-4">
        Additional Information
      </div>

      <div class="text-sm leading-normal font-medium text-fg-muted mb-2">
        External {connector.displayName} files are meant for local development only.
        They may run fine on your machine, but aren't reliably supported in production
        deployments—especially if the file is large (100MB) or outside the data directory.
      </div>
    </div>
  {/if}
  {#if connector.displayName === "Public URL"}
    <div class="mt-8">
      <div class="text-sm leading-none font-medium text-fg-secondary mb-4">
        Supported Sources
      </div>

      <div class="text-sm leading-normal font-medium text-fg-muted mb-2">
        Connect to any publicly accessible dataset via HTTP/HTTPS URLs, including
        public files from GCS, S3, and Azure blob storage.
      </div>
    </div>
  {/if}
</div>
