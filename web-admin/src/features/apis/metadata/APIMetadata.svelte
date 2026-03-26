<script lang="ts">
  import { createAdminServiceGetProject } from "@rilldata/web-admin/client";
  import { useAPI } from "@rilldata/web-admin/features/apis/selectors";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { CANONICAL_ADMIN_API_URL } from "@rilldata/web-admin/client/http-client";
  import { copyToClipboard } from "@rilldata/web-common/lib/actions/copy-to-clipboard";
  import { useRuntimeClient } from "@rilldata/web-common/runtime-client/v2";
  import CopyIcon from "@rilldata/web-common/components/icons/CopyIcon.svelte";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";

  export let organization: string;
  export let project: string;
  export let api: string;

  const runtimeClient = useRuntimeClient();

  $: apiQuery = useAPI(runtimeClient, api);
  $: apiSpec = $apiQuery.data?.resource?.api?.spec;
  $: apiMeta = $apiQuery.data?.resource?.meta;

  $: securityRules = apiSpec?.securityRules ?? [];

  $: projectQuery = createAdminServiceGetProject(organization, project);
  $: canViewPolicy = !!$projectQuery.data?.projectPermissions?.manageProd;

  // Construct the endpoint URL via the admin API gateway
  $: endpointUrl = `${CANONICAL_ADMIN_API_URL}/v1/organizations/${organization}/projects/${project}/runtime/api/${api}`;

  // Extract SQL from resolver properties (used by "sql" and "metrics_sql" resolvers)
  $: sql = apiSpec?.resolverProperties?.sql as string | undefined;

  // Extract connector from resolver properties
  $: connector = apiSpec?.resolverProperties?.connector as string | undefined;

  // Filter out "sql" and "connector" from resolver properties to show the remaining ones
  $: otherResolverProperties = (() => {
    if (!apiSpec?.resolverProperties) return null;
    const props = { ...apiSpec.resolverProperties };
    delete props.sql;
    delete props.connector;
    return Object.keys(props).length > 0 ? props : null;
  })();

  // Safely parse JSON schemas
  $: requestSchema = safeParseJson(apiSpec?.openapiRequestSchemaJson);
  $: responseSchema = safeParseJson(apiSpec?.openapiResponseSchemaJson);

  function safeParseJson(json: string | undefined): unknown | null {
    if (!json) return null;
    try {
      return JSON.parse(json);
    } catch {
      return null;
    }
  }

  let copied = false;
  function copyEndpointUrl() {
    copyToClipboard(endpointUrl, "Copied endpoint URL to clipboard");
    copied = true;
    setTimeout(() => (copied = false), 2500);
  }
</script>

{#if apiSpec}
  <div class="flex flex-col gap-y-9 w-full max-w-full 2xl:max-w-[1200px]">
    <div class="flex flex-col gap-y-2">
      {#if apiSpec.openapiSummary}
        <div class="uppercase text-xs text-fg-secondary font-semibold">
          <span>{apiSpec.openapiSummary}</span>
        </div>
      {/if}
      <h1 class="text-fg-primary text-lg font-bold" aria-label="API name">
        {apiSpec.displayName || api}
      </h1>
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
          class="text-fg-primary text-xs font-mono bg-surface-subtle border border-border rounded-md px-3 py-1.5 overflow-x-auto"
          >{endpointUrl}</code
        >
        <button
          type="button"
          class="p-1 rounded hover:bg-surface-secondary cursor-pointer text-fg-secondary hover:text-fg-primary"
          on:click={copyEndpointUrl}
          aria-label="Copy endpoint URL"
        >
          {#if copied}
            <Check size="14px" />
          {:else}
            <CopyIcon size="14px" />
          {/if}
        </button>
      </div>
    </div>

    <!-- Metadata columns -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Resolver -->
      <div class="flex flex-col gap-y-3" aria-label="API resolver">
        <MetadataLabel>Resolver</MetadataLabel>
        {#if apiSpec.resolver}
          <MetadataValue>{apiSpec.resolver}</MetadataValue>
        {:else}
          <MetadataValue>—</MetadataValue>
        {/if}
      </div>

      <!-- Connector -->
      {#if connector}
        <div class="flex flex-col gap-y-3" aria-label="API connector">
          <MetadataLabel>Connector</MetadataLabel>
          <MetadataValue>{connector}</MetadataValue>
        </div>
      {/if}

      <!-- Security rules -->
      <div class="flex flex-col gap-y-3" aria-label="API security rules">
        <MetadataLabel>Access policy</MetadataLabel>
        {#if securityRules.length === 0}
          <MetadataValue>No additional rules</MetadataValue>
        {:else if canViewPolicy}
          {#each securityRules as rule}
            <MetadataValue>
              <code class="font-mono text-xs"
                >{rule.access?.allow ? "allow" : "deny"}: {rule.access
                  ?.conditionExpression ?? "—"}</code
              >
            </MetadataValue>
          {/each}
        {:else}
          <MetadataValue>Enabled</MetadataValue>
        {/if}
      </div>

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
        <div class="flex flex-col gap-y-3" aria-label="API last updated date">
          <MetadataLabel>Last updated</MetadataLabel>
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
          class="text-fg-primary text-xs font-mono bg-surface-subtle border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{sql.trim()}</pre>
      </div>
    {/if}

    <!-- Other resolver properties -->
    {#if otherResolverProperties}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Resolver properties</MetadataLabel>
        <div
          class="bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto flex flex-col gap-y-2"
        >
          {#each Object.entries(otherResolverProperties) as [key, value]}
            <div class="flex gap-x-2 text-xs font-mono">
              <span class="text-fg-secondary shrink-0">{key}:</span>
              <span class="text-fg-primary whitespace-pre-wrap"
                >{typeof value === "object"
                  ? JSON.stringify(value, null, 2)
                  : value}</span
              >
            </div>
          {/each}
        </div>
      </div>
    {/if}

    <!-- OpenAPI schemas -->
    {#if requestSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Request schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{JSON.stringify(
            requestSchema,
            null,
            2,
          )}</pre>
      </div>
    {/if}

    {#if responseSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Response schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs font-mono bg-surface-secondary border border-border rounded-md p-4 overflow-x-auto whitespace-pre-wrap">{JSON.stringify(
            responseSchema,
            null,
            2,
          )}</pre>
      </div>
    {/if}
  </div>
{/if}
