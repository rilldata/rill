<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import CaretDownIcon from "../../../components/icons/CaretDownIcon.svelte";
  import ClaudeIcon from "../../../components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "../../../components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "../../../components/icons/connectors/OpenAIIcon.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import {
    getBackendConnectorName,
    getConnectorSchema,
  } from "../../sources/modal/connector-schemas";
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

  export let open = false;

  const queryClient = useQueryClient();
  const runtimeClient = useRuntimeClient();

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
  ];

  let schemaName = "claude";
  let apiKey = "";
  let model = "";
  let saving = false;
  let error = "";
  let existingAiConnector = "";

  $: schema = schemaName ? getConnectorSchema(schemaName) : null;
  $: apiKeyProp = schema?.properties?.api_key;
  $: modelProp = schema?.properties?.model;
  $: selectedOption = providerOptions.find((o) => o.value === schemaName);
  $: docsUrl = schemaName
    ? `https://docs.rilldata.com/developers/build/connectors/services/${getBackendConnectorName(schemaName)}`
    : "";

  // Soft validation: warn when the API key doesn't match the expected prefix
  $: apiKeyWarning = getApiKeyWarning(schemaName, apiKey);

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

  // Reset form fields when dialog opens; also fetch the current AI connector
  $: if (open) {
    schemaName = "claude";
    apiKey = "";
    model = "";
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
    apiKey = "";
    model = "";
    error = "";
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
    if (!schemaName || !apiKey) return;
    saving = true;
    error = "";
    try {
      const formValues: Record<string, string> = { api_key: apiKey };
      if (model) formValues.model = model;
      await saveAiConnector(runtimeClient, queryClient, schemaName, formValues);
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

      <Input
        id="ai-connector-api-key"
        label={apiKeyProp?.title ?? "API Key"}
        placeholder={apiKeyProp?.["x-placeholder"] ?? ""}
        hint={apiKeyProp?.description ?? ""}
        secret
        bind:value={apiKey}
      />
      {#if apiKeyWarning}
        <p class="text-sm text-red-500">{apiKeyWarning}</p>
      {/if}

      <Input
        id="ai-connector-model"
        label={modelProp?.title ?? "Model"}
        placeholder={modelProp?.["x-placeholder"] ?? ""}
        hint={modelProp?.description ?? ""}
        optional
        bind:value={model}
      />
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
      <AlertDialog.Cancel asChild>
        <Button large type="secondary" disabled={saving}
          >Cancel</Button
        >
      </AlertDialog.Cancel>

      <!-- Use a plain button instead of AlertDialog.Action to prevent
           the dialog from auto-closing before the async save completes. -->
      <Button
        disabled={!apiKey || saving}
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
