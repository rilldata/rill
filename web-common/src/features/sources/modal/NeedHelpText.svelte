<script lang="ts">
  import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";
  import { ExternalLinkIcon } from "lucide-svelte";
  import { connectorStepStore } from "./connectorStepStore";

  export let connector: V1ConnectorDriver;

  $: stepState = $connectorStepStore;
</script>

<div>
  <div class="text-sm leading-none font-medium mb-4">Help</div>
  {#if stepState.step === "connector"}
    <div class="text-sm leading-normal font-medium text-muted-foreground mb-2">
      Need help connecting to {connector.displayName}? Check out our
      documentation for detailed instructions.
    </div>
    <span class="flex flex-row items-center gap-2 group mb-4">
      <a
        href={connector.docsUrl ||
          "https://docs.rilldata.com/build/connectors/"}
        rel="noreferrer noopener"
        target="_blank"
        class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium group-hover:underline break-all"
      >
        How to connect to {connector.displayName}
      </a>
      <ExternalLinkIcon size="16px" color="#6366F1" />
    </span>
  {:else}
    <div class="text-sm leading-normal font-medium text-muted-foreground mb-2">
      Check out our <a
        href="https://docs.rilldata.com/build/models/source-models/"
        rel="noreferrer noopener"
        target="_blank"
        class="text-sm leading-normal text-primary-500 hover:text-primary-600 font-medium group-hover:underline break-all"
      >
        source model documentation
      </a> for detailed instructions on how to customize up your data source ingestion.
    </div>
  {/if}

  {#if connector.displayName === "DuckDB" || connector.displayName === "SQLite"}
    <div class="mt-8">
      <div class="text-sm leading-none font-medium mb-4">
        Additional Information
      </div>

      <div
        class="text-sm leading-normal font-medium text-muted-foreground mb-2"
      >
        External {connector.displayName} files are meant for local development only.
        They may run fine on your machine, but aren’t reliably supported in production
        deployments—especially if the file is large (100MB) or outside the data directory.
      </div>
    </div>
  {/if}
</div>
