<script lang="ts">
  import { useAPI } from "@rilldata/web-admin/features/apis/selectors";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;
  export let api: string;

  $: ({ instanceId, host } = $runtime);

  $: apiQuery = useAPI(instanceId, api);
  $: apiSpec = $apiQuery.data?.resource?.api?.spec;
  $: apiMeta = $apiQuery.data?.resource?.meta;

  $: securityRuleCount = apiSpec?.securityRules?.length ?? 0;
  $: hasRequestSchema = !!apiSpec?.openapiRequestSchemaJson;
  $: hasResponseSchema = !!apiSpec?.openapiResponseSchemaJson;

  // Construct the endpoint URL
  $: endpointPath = `/v1/instances/${instanceId}/api/${api}`;
  $: endpointUrl = host ? `${host}${endpointPath}` : endpointPath;

  // Extract SQL from resolver properties (used by "sql" and "metrics_sql" resolvers)
  $: sql = apiSpec?.resolverProperties?.sql as string | undefined;

  // Filter out "sql" from resolver properties to show the remaining ones
  $: otherResolverProperties = (() => {
    if (!apiSpec?.resolverProperties) return null;
    const { sql: _sql, ...rest } = apiSpec.resolverProperties;
    return Object.keys(rest).length > 0 ? rest : null;
  })();
</script>

{#if apiSpec}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      <!-- Header row 1 -->
      <div class="uppercase text-xs text-fg-secondary font-semibold">
        {#if apiSpec.openapiSummary}
          <span>{apiSpec.openapiSummary}</span>
        {/if}
      </div>
      <div class="flex gap-x-2 items-center">
        <h1 class="text-fg-primary text-lg font-bold" aria-label="API name">
          {apiSpec.displayName || api}
        </h1>
        <div class="grow" />
      </div>
      {#if apiSpec.description}
        <p class="text-fg-secondary text-sm">{apiSpec.description}</p>
      {/if}
    </div>

    <!-- Endpoint URL -->
    <div class="flex flex-col gap-y-3">
      <MetadataLabel>Endpoint</MetadataLabel>
      <div class="flex items-center gap-x-3">
        <span
          class="text-xs font-semibold px-1.5 py-0.5 rounded bg-surface-secondary text-fg-secondary"
          >GET / POST</span
        >
        <code
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md px-3 py-1.5 overflow-x-auto"
          >{endpointUrl}</code
        >
      </div>
    </div>

    <!-- Metadata columns -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Resolver -->
      <div class="flex flex-col gap-y-3" aria-label="API resolver">
        <MetadataLabel>Resolver</MetadataLabel>
        <MetadataValue>{apiSpec.resolver || "—"}</MetadataValue>
      </div>

      <!-- Authentication -->
      <div class="flex flex-col gap-y-3" aria-label="API authentication">
        <MetadataLabel>Authentication</MetadataLabel>
        <MetadataValue>Bearer token</MetadataValue>
      </div>

      <!-- Security rules -->
      {#if securityRuleCount > 0}
        <div class="flex flex-col gap-y-3" aria-label="API security rules">
          <MetadataLabel>Security rules</MetadataLabel>
          <MetadataValue>
            {securityRuleCount} rule{securityRuleCount > 1 ? "s" : ""}
          </MetadataValue>
        </div>
      {/if}

      <!-- Created -->
      {#if apiMeta?.createdOn}
        <div class="flex flex-col gap-y-3" aria-label="API created date">
          <MetadataLabel>Created on</MetadataLabel>
          <MetadataValue>
            {new Date(apiMeta.createdOn).toLocaleDateString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
            })}
          </MetadataValue>
        </div>
      {/if}

      <!-- Last updated -->
      {#if apiMeta?.stateUpdatedOn}
        <div class="flex flex-col gap-y-3" aria-label="API last executed date">
          <MetadataLabel>Last executed on</MetadataLabel>
          <MetadataValue>
            {new Date(apiMeta.stateUpdatedOn).toLocaleDateString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
            })}
          </MetadataValue>
        </div>
      {/if}
    </div>

    <!-- SQL -->
    {#if sql}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>SQL</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{sql.trim()}</pre>
      </div>
    {/if}

    <!-- Other resolver properties -->
    {#if otherResolverProperties}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Resolver properties</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{JSON.stringify(otherResolverProperties, null, 2)}</pre>
      </div>
    {/if}

    <!-- OpenAPI schemas -->
    {#if hasRequestSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Request schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{JSON.stringify(JSON.parse(apiSpec.openapiRequestSchemaJson ?? ""), null, 2)}</pre>
      </div>
    {/if}

    {#if hasResponseSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Response schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{JSON.stringify(JSON.parse(apiSpec.openapiResponseSchemaJson ?? ""), null, 2)}</pre>
      </div>
    {/if}

    <!-- TODO: add a "Try it" section to test the API endpoint inline -->
  </div>
{/if}
