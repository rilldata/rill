<script lang="ts">
  import { useAPI } from "@rilldata/web-admin/features/apis/selectors";
  import MetadataLabel from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataLabel.svelte";
  import MetadataValue from "@rilldata/web-admin/features/scheduled-reports/metadata/MetadataValue.svelte";
  import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";

  export let organization: string;
  export let project: string;
  export let api: string;

  $: ({ instanceId } = $runtime);

  $: apiQuery = useAPI(instanceId, api);
  $: apiSpec = $apiQuery.data?.resource?.api?.spec;
  $: apiMeta = $apiQuery.data?.resource?.meta;

  // TODO: test with a deployed project that has API resources
  $: securityRuleCount = apiSpec?.securityRules?.length ?? 0;
  $: hasRequestSchema = !!apiSpec?.openapiRequestSchemaJson;
  $: hasResponseSchema = !!apiSpec?.openapiResponseSchemaJson;
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
          {api}
        </h1>
        <div class="grow" />
      </div>
    </div>

    <!-- Metadata columns -->
    <div class="flex flex-wrap gap-x-16 gap-y-6">
      <!-- Resolver -->
      <div class="flex flex-col gap-y-3" aria-label="API resolver">
        <MetadataLabel>Resolver</MetadataLabel>
        <MetadataValue>{apiSpec.resolver || "—"}</MetadataValue>
      </div>

      <!-- Security rules -->
      <div class="flex flex-col gap-y-3" aria-label="API security rules">
        <MetadataLabel>Security rules</MetadataLabel>
        <MetadataValue>
          {securityRuleCount === 0
            ? "None"
            : `${securityRuleCount} rule${securityRuleCount > 1 ? "s" : ""}`}
        </MetadataValue>
      </div>

      <!-- TODO: add API endpoint URL once available from the runtime -->

      <!-- Created -->
      {#if apiMeta?.createdOn}
        <div class="flex flex-col gap-y-3" aria-label="API created date">
          <MetadataLabel>Created</MetadataLabel>
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
      {#if apiMeta?.specUpdatedOn}
        <div class="flex flex-col gap-y-3" aria-label="API last updated date">
          <MetadataLabel>Last updated</MetadataLabel>
          <MetadataValue>
            {new Date(apiMeta.specUpdatedOn).toLocaleDateString(undefined, {
              year: "numeric",
              month: "short",
              day: "numeric",
            })}
          </MetadataValue>
        </div>
      {/if}
    </div>

    <!-- Resolver properties -->
    {#if apiSpec.resolverProperties && Object.keys(apiSpec.resolverProperties).length > 0}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Resolver properties</MetadataLabel>
        <pre
          class="text-fg-primary text-xs bg-surface-secondary rounded-md p-3 overflow-x-auto">{JSON.stringify(apiSpec.resolverProperties, null, 2)}</pre>
      </div>
    {/if}

    <!-- OpenAPI schemas -->
    <!-- TODO: test rendering of these sections with real API data -->
    {#if hasRequestSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Request schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs bg-surface-secondary rounded-md p-3 overflow-x-auto">{apiSpec.openapiRequestSchemaJson}</pre>
      </div>
    {/if}

    {#if hasResponseSchema}
      <div class="flex flex-col gap-y-3">
        <MetadataLabel>Response schema</MetadataLabel>
        <pre
          class="text-fg-primary text-xs bg-surface-secondary rounded-md p-3 overflow-x-auto">{apiSpec.openapiResponseSchemaJson}</pre>
      </div>
    {/if}

    <!-- TODO: add a "Try it" section to test the API endpoint inline -->
    <!-- TODO: add reconcile status / error display -->
  </div>
{/if}
