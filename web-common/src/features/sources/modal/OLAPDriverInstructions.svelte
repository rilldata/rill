<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import CodeBlock from "../../../components/code-block/CodeBlock.svelte";
  import { V1ConnectorDriver } from "../../../runtime-client";

  const dispatch = createEventDispatcher();

  export let connector: V1ConnectorDriver;

  $: displayName = connector?.displayName;
  $: code = getCodeSnippet(connector?.name as string);
  $: docsUrl = getDocsUrl(connector?.name as string); // The docsUrl should be available in the connector object, but it's not.

  function getCodeSnippet(connectorName: string) {
    switch (connectorName) {
      case "clickhouse":
        return `# Connecting to clickhouse local
rill start --db-driver clickhouse --db "clickhouse://localhost:9000"

# Connecting to clickhouse cluster 
rill start --db-driver clickhouse --db "<clickhouse://<host>:<port>?username=<username>&password=<pass>>"
`;
      case "pinot":
        return `# Connecting to pinot
rill start --db-driver pinot --db "http(s)://username:password@localhost:9000"
`;
      default:
        throw new Error(`Unsupported connector: ${connectorName}`);
    }
  }

  function getDocsUrl(connectorName: string) {
    switch (connectorName) {
      case "clickhouse":
        return "https://docs.rilldata.com/reference/olap-engines/clickhouse";
      case "pinot":
        return "https://docs.rilldata.com/reference/olap-engines/pinot";
      default:
        throw new Error(`Unsupported connector: ${connectorName}`);
    }
  }
</script>

<div class="wrapper">
  <div class="content">
    <span>
      To use Rill with {displayName}, restart Rill with {displayName}
      as your database driver and include the connection string to your {displayName}
      instance.
    </span>
    <CodeBlock {code} />
    <span>
      Launching Rill with these settings will display your {displayName}
      tables in the sidebar. Then you can build dashboards directly on top of any
      table.
    </span>
    <span>
      Note: Data modeling is not yet supported for the {displayName} driver.
    </span>
    <span>
      Need help? Refer to our <a
        href={docsUrl}
        target="_blank"
        rel="noreferrer noopener">{displayName} docs</a
      >
      for more information.
    </span>
  </div>
  <div class="button-wrapper">
    <div class="grow" />
    <Button on:click={() => dispatch("back")} type="secondary">Back</Button>
  </div>
</div>

<style lang="postcss">
  .wrapper {
    @apply h-full w-full mt-1;
    @apply flex flex-col;
  }

  .content {
    @apply flex-grow pb-5;
    @apply flex flex-col gap-y-3;
    @apply text-xs;
  }

  .button-wrapper {
    @apply flex items-center space-x-2;
  }
</style>
