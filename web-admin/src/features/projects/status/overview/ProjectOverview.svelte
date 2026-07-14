<script lang="ts">
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import { useParserReconcileError } from "../selectors";
  import DeploymentSection from "./DeploymentSection.svelte";
  import ResourcesSection from "./ResourcesSection.svelte";
  import TablesSection from "./TablesSection.svelte";
  import ErrorsSection from "./ErrorsSection.svelte";

  export let organization: string;
  export let project: string;

  const runtimeClient = useRuntimeClient();
  $: parserErrorQuery = useParserReconcileError(runtimeClient);
  $: hasProjectError = !!($parserErrorQuery.data ?? "");
</script>

<div class="flex flex-col gap-6">
  <DeploymentSection {organization} {project} />
  {#if !hasProjectError}
    <ResourcesSection />
    <TablesSection />
    <ErrorsSection />
  {/if}
</div>
