<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import BedrockIcon from "../../../components/icons/connectors/BedrockIcon.svelte";
  import ClaudeIcon from "../../../components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "../../../components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "../../../components/icons/connectors/OpenAIIcon.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    getBackendConnectorName,
    getConnectorSchema,
  } from "../../sources/modal/connector-schemas";
  import {
    getRequiredFieldsForValues,
    getSchemaInitialValues,
    isVisibleForValues,
    radioOptions,
  } from "../../templates/schema-utils";
  import type { MultiStepFormSchema } from "../../templates/schemas/types";
  import { saveAiConnector } from "./saveAiConnector";
  import { ExternalLinkIcon } from "lucide-svelte";
  import type { ComponentType, SvelteComponent } from "svelte";
  import { useRuntimeClient } from "../../../runtime-client/v2";
  import {
    getRuntimeServiceGetInstanceQueryKey,
    runtimeServiceGetInstance,
  } from "../../../runtime-client";
  import { behaviourEvent } from "../../../metrics/initMetrics";
  import {
    BehaviourEventAction,
    BehaviourEventMedium,
  } from "../../../metrics/service/BehaviourEventTypes";
  import { MetricsEventSpace } from "../../../metrics/service/MetricsTypes";
  import { getScreenNameFromPage } from "../../file-explorer/telemetry";
  import { getEnvFileStore } from "@rilldata/web-common/features/env-management/env-file-store.ts";
  import { EnvEditSession } from "@rilldata/web-common/features/env-management/env-edit-session.ts";

  export let open = false;

  const queryClient = useQueryClient();
  const runtimeClient = useRuntimeClient();
  const envStore = getEnvFileStore();

  /** Expected API key prefixes per provider, used for soft validation. */
  const API_KEY_PREFIXES: Record<string, { prefix: string; label: string }> = {
    claude: { prefix: "sk-ant-", label: "Claude" },
    openai: { prefix: "sk-", label: "OpenAI" },
    gemini: { prefix: "AIza", label: "Gemini" },
  };

  const providerOptions: Array<{
    value: string;
    label: string;
    icon: ComponentType<SvelteComponent>;
  }> = [
    { value: "claude", label: "Claude", icon: ClaudeIcon },
    { value: "gemini", label: "Gemini", icon: GeminiIcon },
    { value: "openai", label: "OpenAI", icon: OpenAIIcon },
    { value: "bedrock", label: "Amazon Bedrock", icon: BedrockIcon },
  ];

  const defaultSchemaName = "claude";

  let schemaName = defaultSchemaName;
  let values: Record<string, string> = {};
  let saving = false;
  let error = "";
  let existingAiConnector = "";

  $: schema = schemaName ? getConnectorSchema(schemaName) : null;
  // Render the provider's connector fields in schema order (e.g. an API key for
  // Claude, or an auth method choice, credentials, and a region for Bedrock).
  $: fields = schema
    ? Object.entries(schema.properties ?? {}).filter(
        ([key, prop]) =>
          !prop["x-advanced"] && isVisibleForValues(schema, key, values),
      )
    : [];
  $: required = schema
    ? getRequiredFieldsForValues(schema, values)
    : new Set<string>();
  $: canSave = fields.every(([key]) => !required.has(key) || !!values[key]);
  $: selectedOption = providerOptions.find((o) => o.value === schemaName);
  $: docsUrl = schemaName
    ? `https://docs.rilldata.com/developers/build/connectors/services/${getBackendConnectorName(schemaName)}`
    : "";

  $: envEditSession = new EnvEditSession(envStore, schemaName, schema);

  // Soft validation: warn when the API key doesn't match the expected prefix
  $: apiKeyWarning = getApiKeyWarning(schemaName, values.api_key ?? "");

  function getApiKeyWarning(provider: string, key: string): string {
    if (!key) return "";
    const expected = API_KEY_PREFIXES[provider];
    if (!expected) return "";
    if (key.startsWith(expected.prefix)) {
      // Guard against subset matches: OpenAI's "sk-" prefix also matches
      // Claude keys ("sk-ant-"). Reject keys that match a more-specific provider.
      const moreSpecific = Object.entries(API_KEY_PREFIXES).find(
        ([p, spec]) =>
          p !== provider &&
          spec.prefix.startsWith(expected.prefix) &&
          key.startsWith(spec.prefix),
      );
      if (moreSpecific) {
        return `This looks like a ${moreSpecific[1].label} API key, not ${expected.label}`;
      }
      return "";
    }
    return `This doesn't look like a ${expected.label} API key`;
  }

  // Reset form fields when dialog opens; also fetch the current AI connector.
  // Note: this block must not read `schemaName` or `values`, or it would re-run
  // (and reset the form) whenever the user changes the provider or a field.
  $: if (open) {
    schemaName = defaultSchemaName;
    values = initialValues(getConnectorSchema(defaultSchemaName));
    saving = false;
    error = "";
    existingAiConnector = "";
    fetchExistingAiConnector();
    behaviourEvent?.fireSourceTriggerEvent(
      BehaviourEventAction.SourceModal,
      BehaviourEventMedium.Button,
      getScreenNameFromPage(),
      MetricsEventSpace.Modal,
    );
  }

  // Clear inputs when the provider changes
  function handleProviderChange(value: string) {
    schemaName = value;
    values = initialValues(getConnectorSchema(value));
    error = "";
  }

  // Defaults such as the initial auth method choice.
  function initialValues(
    schema: MultiStepFormSchema | null,
  ): Record<string, string> {
    if (!schema) return {};
    return Object.fromEntries(
      Object.entries(getSchemaInitialValues(schema)).map(([k, v]) => [
        k,
        String(v),
      ]),
    );
  }

  async function fetchExistingAiConnector() {
    try {
      const instance = await queryClient.fetchQuery({
        queryKey: getRuntimeServiceGetInstanceQueryKey(
          runtimeClient.instanceId,
        ),
        queryFn: () => runtimeServiceGetInstance(runtimeClient, {}),
      });
      existingAiConnector = instance?.instance?.aiConnector ?? "";
    } catch {
      existingAiConnector = "";
    }
  }

  async function handleSave() {
    if (!schemaName || !canSave) return;
    saving = true;
    error = "";
    try {
      // Wait for the first .env pull so the edit session allocates names against
      // the persisted secrets rather than an empty store, which would let the
      // commit overwrite an existing secret instead of suffixing it.
      await envStore.whenReady();
      // Only submit visible, non-empty connector properties (e.g. skip the
      // UI-only auth method and credentials for the unselected auth method).
      const formValues = Object.fromEntries(
        fields
          .filter(([key, prop]) => !prop["x-ui-only"] && values[key])
          .map(([key]) => [key, values[key]]),
      );
      await saveAiConnector(
        runtimeClient,
        queryClient,
        schemaName,
        formValues,
        envEditSession,
      );
      behaviourEvent?.fireSourceTriggerEvent(
        BehaviourEventAction.SourceAdd,
        BehaviourEventMedium.Button,
        getScreenNameFromPage(),
        MetricsEventSpace.Modal,
      );
      open = false;
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to save connector";
      // Reusing SourceCancel here because there is no dedicated SourceError event.
      // This fires on save failure, not user cancellation.
      behaviourEvent?.fireSourceTriggerEvent(
        BehaviourEventAction.SourceCancel,
        BehaviourEventMedium.Button,
        getScreenNameFromPage(),
        MetricsEventSpace.Modal,
      );
    } finally {
      saving = false;
    }
  }
</script>

<AlertDialog.Root bind:open>
  <AlertDialog.Content>
    <AlertDialog.Title>Add AI connector</AlertDialog.Title>

    <div class="flex flex-col gap-y-3">
      <div class="flex flex-col gap-y-2">
        <div class="flex items-center justify-between">
          <span class="text-sm font-medium">Provider</span>
          {#if docsUrl}
            <a
              href={docsUrl}
              rel="noreferrer noopener"
              target="_blank"
              class="inline-flex items-center gap-1 text-sm text-primary-500 hover:text-primary-600 hover:underline"
            >
              View documentation
              <ExternalLinkIcon size="14px" />
            </a>
          {/if}
        </div>
        <SelectPrimitive.Root
          type="single"
          value={schemaName}
          onValueChange={(val) => {
            if (val) handleProviderChange(val);
          }}
        >
          <SelectPrimitive.Trigger
            class="flex h-8 w-full items-center justify-between rounded-[2px] border bg-transparent px-3 text-sm ring-offset-background focus:outline-none focus:border-primary-400"
          >
            {#if selectedOption}
              <div class="flex items-center gap-2">
                <svelte:component this={selectedOption.icon} size="16px" />
                <span class="text-sm text-fg-primary">
                  {selectedOption.label}
                </span>
              </div>
            {:else}
              <span class="text-fg-muted">Select a provider</span>
            {/if}
            <div class="caret transition-transform ml-2">
              <CaretDownIcon size="12px" className="fill-fg-secondary" />
            </div>
          </SelectPrimitive.Trigger>

          <Select.Content sameWidth>
            {#each providerOptions as option (option.value)}
              <Select.Item value={option.value} class="py-1.5">
                <div class="flex items-center gap-2">
                  <svelte:component this={option.icon} size="16px" />
                  <span class="text-sm">{option.label}</span>
                </div>
              </Select.Item>
            {/each}
          </Select.Content>
        </SelectPrimitive.Root>
      </div>

      {#each fields as [key, prop] (`${schemaName}.${key}`)}
        {#if prop["x-display"] === "radio"}
          <div class="flex flex-col gap-y-2">
            <span class="text-sm font-medium">{prop.title ?? key}</span>
            <Radio
              name={`ai-connector-${key}`}
              options={radioOptions(prop)}
              bind:value={values[key]}
            />
          </div>
        {:else}
          <Input
            id={`ai-connector-${key}`}
            label={prop.title ?? key}
            placeholder={prop["x-placeholder"] ?? ""}
            hint={prop.description ?? ""}
            secret={!!prop["x-secret"]}
            optional={!required.has(key)}
            bind:value={values[key]}
          />
        {/if}
        {#if key === "api_key" && apiKeyWarning}
          <p class="text-sm text-red-500">{apiKeyWarning}</p>
        {/if}
      {/each}
      {#if existingAiConnector}
        <p class="text-sm text-red-500">
          This will replace your existing AI connector ({existingAiConnector}).
        </p>
      {/if}
      {#if error}
        <p class="text-sm text-red-500">{error}</p>
      {/if}
    </div>

    <AlertDialog.Footer>
      <AlertDialog.Cancel>
        {#snippet child({ props })}
          <Button {...props} large type="secondary" disabled={saving}
            >Cancel</Button
          >
        {/snippet}
      </AlertDialog.Cancel>

      <!-- Use a plain button instead of AlertDialog.Action to prevent
           the dialog from auto-closing before the async save completes. -->
      <Button
        disabled={!canSave || saving}
        large
        type="primary"
        onClick={handleSave}
      >
        {saving ? "Saving..." : "Save"}
      </Button>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  :global(button[aria-expanded="true"] > .caret) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
