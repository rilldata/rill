<script lang="ts">
  import { Select as SelectPrimitive } from "bits-ui";
  import * as AlertDialog from "@rilldata/web-common/components/alert-dialog";
  import * as Select from "@rilldata/web-common/components/select";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import CaretDownIcon from "../../components/icons/CaretDownIcon.svelte";
  import ClaudeIcon from "../../components/icons/connectors/ClaudeIcon.svelte";
  import GeminiIcon from "../../components/icons/connectors/GeminiIcon.svelte";
  import OpenAIIcon from "../../components/icons/connectors/OpenAIIcon.svelte";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { getConnectorSchema } from "../sources/modal/connector-schemas";
  import { saveAiConnector } from "../sources/modal/submitAddDataForm";
  import type { ComponentType, SvelteComponent } from "svelte";

  export let open = false;

  const queryClient = useQueryClient();

  const providerOptions: Array<{
    value: string;
    label: string;
    icon: ComponentType<SvelteComponent>;
  }> = [
    { value: "claude", label: "Claude", icon: ClaudeIcon },
    { value: "openai", label: "OpenAI", icon: OpenAIIcon },
    { value: "gemini", label: "Gemini", icon: GeminiIcon },
  ];

  let schemaName = "claude";
  let apiKey = "";
  let model = "";
  let saving = false;
  let error = "";

  $: schema = schemaName ? getConnectorSchema(schemaName) : null;
  $: apiKeyProp = schema?.properties?.api_key;
  $: modelProp = schema?.properties?.model;
  $: selectedOption = providerOptions.find((o) => o.value === schemaName);

  // Reset form fields when dialog opens
  $: if (open) {
    schemaName = "claude";
    apiKey = "";
    model = "";
    saving = false;
    error = "";
  }

  async function handleSave() {
    if (!schemaName || !apiKey) return;
    saving = true;
    error = "";
    try {
      const formValues: Record<string, string> = { api_key: apiKey };
      if (model) formValues.model = model;
      await saveAiConnector(queryClient, schemaName, formValues);
      open = false;
    } catch (e) {
      error = e instanceof Error ? e.message : "Failed to save connector";
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
        <span class="text-sm font-medium">Provider</span>
        <SelectPrimitive.Root
          selected={{ value: schemaName }}
          onSelectedChange={(s) => {
            if (s?.value) schemaName = s.value;
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

      <Input
        id="ai-connector-model"
        label={modelProp?.title ?? "Model"}
        placeholder={modelProp?.["x-placeholder"] ?? ""}
        hint={modelProp?.description ?? ""}
        optional
        bind:value={model}
      />

      {#if error}
        <p class="text-sm text-red-500">{error}</p>
      {/if}
    </div>

    <AlertDialog.Footer>
      <AlertDialog.Cancel asChild let:builder>
        <Button large builders={[builder]} type="secondary">Cancel</Button>
      </AlertDialog.Cancel>

      <AlertDialog.Action asChild let:builder>
        <Button
          disabled={!apiKey || saving}
          large
          builders={[builder]}
          type="primary"
          onClick={handleSave}
        >
          {saving ? "Saving..." : "Save"}
        </Button>
      </AlertDialog.Action>
    </AlertDialog.Footer>
  </AlertDialog.Content>
</AlertDialog.Root>

<style lang="postcss">
  :global(button[aria-expanded="true"] > .caret) {
    @apply transform -rotate-180 transition-transform;
  }
</style>
