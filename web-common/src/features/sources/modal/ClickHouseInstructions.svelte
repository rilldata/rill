<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import CodeBlock from "../../../components/code-block/CodeBlock.svelte";

  const dispatch = createEventDispatcher();

  let code = `# Connecting to clickhouse local
rill start --db-driver clickhouse --db "clickhouse://localhost:9000"

# Connecting to clickhouse cluster 
rill start --db-driver clickhouse --db "<clickhouse://<host>:<port>?username=<username>&password=<pass>>"
`;
</script>

<div class="h-full w-full flex flex-col mt-1">
  <div class="pb-5 flex-grow flex flex-col gap-y-3 ui-copy text-xs">
    <span>
      To use Rill with ClickHouse, restart Rill with ClickHouse as your database
      driver and include the connection string to your ClickHouse instance.
    </span>
    <CodeBlock {code} />
    <span>
      Launching Rill with these settings will display your ClickHouse tables in
      the sidebar. Then you can build dashboards directly on top of any table.
    </span>
    <span>
      Note: Data modeling is not yet supported for the ClickHouse driver.
    </span>
  </div>
  <div class="flex items-center space-x-2">
    <div class="grow" />
    <Button on:click={() => dispatch("back")} type="secondary">Back</Button>
  </div>
</div>
