<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
  import {
    ConnectorDriverPropertyType,
    type RpcStatus,
    type V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { defaults, superForm } from "sveltekit-superforms";
  import { yup } from "sveltekit-superforms/adapters";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "./errors";
  import { submitAddDataForm } from "./submitAddDataForm";
  import type { AddDataFormType } from "./types";
  import { getYupSchema, toYupFriendlyKey } from "./yupSchemas";

  export let connector: V1ConnectorDriver;
  export let formType: AddDataFormType;
  export let onBack: () => void;
  export let onClose: () => void;

  $: formId = `add-data-${connector.name}-form`;

  $: isSourceForm = formType === "source";
  $: isConnectorForm = formType === "connector";
  $: properties = isConnectorForm
    ? (connector.configProperties ?? [])
    : (connector.sourceProperties ?? []);

  let rpcError: RpcStatus | null = null;

  const schema = yup(getYupSchema[connector.name as keyof typeof getYupSchema]);

  const { form, errors, enhance, tainted, submit, submitting } = superForm(
    defaults(schema),
    {
      SPA: true,
      validators: schema,
      async onUpdate({ form }) {
        if (!form.valid) return;
        const values = form.data;
        if (isSourceForm) {
          try {
            await submitAddDataForm(queryClient, formType, connector, values);
            onClose();
          } catch (e) {
            rpcError = e?.response?.data;
          }
          return;
        }

        // Connectors
        try {
          await submitAddDataForm(queryClient, formType, connector, values);
          onClose();
        } catch (e) {
          rpcError = e?.response?.data;
        }
      },
    },
  );

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($tainted?.name) return;
      const name = inferSourceName(connector, value);
      if (name)
        form.update(
          ($form) => {
            $form.name = name;
            return $form;
          },
          { taint: false },
        );
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  <form
    class="pb-5 flex-grow overflow-y-auto"
    id={formId}
    use:enhance
    on:submit|preventDefault={submit}
  >
    <div class="pb-2 text-slate-500">
      Need help? Refer to our
      <a
        href="https://docs.rilldata.com/build/connect"
        rel="noreferrer noopener"
        target="_blank">docs</a
      > for more information.
    </div>
    {#if rpcError}
      <SubmissionError
        message={humanReadableErrorMessage(
          connector.name,
          rpcError.code,
          rpcError.message,
        )}
      />
    {/if}

    {#each properties as property (property.key)}
      {#if property.key !== undefined && !property.noPrompt}
        {@const label =
          property.displayName + (property.required ? "" : " (optional)")}
        <div class="py-1.5">
          {#if property.type === ConnectorDriverPropertyType.TYPE_STRING || property.type === ConnectorDriverPropertyType.TYPE_NUMBER}
            <Input
              id={toYupFriendlyKey(property.key)}
              label={property.displayName}
              placeholder={property.placeholder}
              optional={!property.required}
              secret={property.secret}
              hint={property.hint}
              errors={$errors[toYupFriendlyKey(property.key)]}
              bind:value={$form[toYupFriendlyKey(property.key)]}
              onInput={(_, e) => onStringInputChange(e)}
              alwaysShowError
            />
          {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN}
            <label for={property.key} class="flex items-center">
              <input
                id={property.key}
                type="checkbox"
                bind:checked={$form[property.key]}
                class="h-5 w-5"
              />
              <span class="ml-2 text-sm">{label}</span>
            </label>
          {:else if property.type === ConnectorDriverPropertyType.TYPE_INFORMATIONAL}
            <InformationalField
              description={property.description}
              hint={property.hint}
              href={property.docsUrl}
            />
          {/if}
        </div>
      {/if}
    {/each}
  </form>
  <div class="flex items-center space-x-2 ml-auto">
    <Button on:click={onBack} type="secondary">Back</Button>
    <Button disabled={$submitting} form={formId} submitForm type="primary">
      Add data
    </Button>
  </div>
</div>
