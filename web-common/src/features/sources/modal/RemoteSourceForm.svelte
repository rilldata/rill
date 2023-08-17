<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import DialogFooter from "@rilldata/web-common/components/modal/dialog/DialogFooter.svelte";
  import {
    ConnectorSpecProperty,
    ConnectorSpecPropertyType,
    RpcStatus,
    V1ConnectorSpec,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import type { Writable } from "svelte/store";
  import type * as yup from "yup";
  import { overlay } from "../../../layout/overlay-store";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "./errors";
  import { submitRemoteSourceForm } from "./submitRemoteSourceForm";
  import { getYupSchema, toYupFriendlyKey } from "./yupSchemas";

  export let connector: V1ConnectorSpec;

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  let connectorProperties: ConnectorSpecProperty[];
  let yupSchema: yup.AnyObjectSchema;
  let rpcError: RpcStatus = null;

  // state from svelte-forms-lib
  let form: Writable<any>;
  let touched: Writable<Record<any, boolean>>;
  let errors: Writable<Record<never, string>>;
  let handleChange: (event: Event) => any;
  let handleSubmit: (event: Event) => any;
  let isSubmitting: Writable<boolean>;

  function onConnectorChange(connector: V1ConnectorSpec) {
    yupSchema = getYupSchema(connector);

    ({ form, touched, errors, handleChange, handleSubmit, isSubmitting } =
      createForm({
        initialValues: {
          sourceName: "", // avoids `values.sourceName` warning
        },
        validationSchema: yupSchema,
        onSubmit: async (values) => {
          overlay.set({ title: `Importing ${values.sourceName}` });
          try {
            await submitRemoteSourceForm(queryClient, connector.name, values);
            goto(`/source/${values.sourceName}`);
            dispatch("close");
          } catch (e) {
            rpcError = e?.response?.data;
          }
          overlay.set(null);
        },
      }));

    // Place the "Source name" field directly under the "Path" field, which is the first property for each connector (s3, gcs, https).
    connectorProperties = [
      ...connector.properties.slice(0, 1),
      {
        key: "sourceName",
        displayName: "Source name",
        description: "The name of the source",
        placeholder: "my_new_source",
        type: ConnectorSpecPropertyType.TYPE_STRING,
        nullable: false,
      },
      ...connector.properties.slice(1),
    ];
  }

  $: onConnectorChange(connector);

  function onStringInputChange(event: Event) {
    const target = event.target as HTMLInputElement;
    const { name, value } = target;

    if (name === "path") {
      if ($touched.sourceName) return;
      const sourceName = inferSourceName(connector, value);
      $form.sourceName = sourceName ? sourceName : $form.sourceName;
    }
  }
</script>

<div class="h-full w-full flex flex-col">
  <form
    on:submit|preventDefault={handleSubmit}
    id="remote-source-{connector.name}-form"
    class="px-4 pb-2 flex-grow overflow-y-auto"
  >
    <div class="pt-4 pb-2">
      Need help? Refer to our
      <a
        href="https://docs.rilldata.com/develop/import-data"
        target="_blank"
        rel="noreferrer">docs</a
      > for more information.
    </div>
    {#if rpcError}
      <SubmissionError
        message={humanReadableErrorMessage(
          connector.name,
          rpcError.code,
          rpcError.message
        )}
      />
    {/if}

    {#each connectorProperties as property}
      {@const label =
        property.displayName + (property.nullable ? " (optional)" : "")}
      <div class="py-2">
        {#if property.type === ConnectorSpecPropertyType.TYPE_STRING}
          <Input
            id={toYupFriendlyKey(property.key)}
            {label}
            placeholder={property.placeholder}
            hint={property.hint}
            error={$errors[toYupFriendlyKey(property.key)]}
            bind:value={$form[toYupFriendlyKey(property.key)]}
            on:input={onStringInputChange}
            on:change={handleChange}
          />
        {:else if property.type === ConnectorSpecPropertyType.TYPE_BOOLEAN}
          <label for={property.key} class="flex items-center">
            <input
              id={property.key}
              type="checkbox"
              bind:checked={$form[property.key]}
              class="h-5 w-5"
            />
            <span class="ml-2 text-sm">{label}</span>
          </label>
        {:else if property.type === ConnectorSpecPropertyType.TYPE_INFORMATIONAL}
          <InformationalField
            description={property.description}
            hint={property.hint}
            href={property.href}
          />
        {/if}
      </div>
    {/each}
  </form>
  <div class="bg-gray-100 border-t border-gray-300">
    <DialogFooter>
      <div class="flex items-center space-x-2">
        <Button
          type="primary"
          submitForm
          form="remote-source-{connector.name}-form"
          disabled={$isSubmitting}
        >
          Add source
        </Button>
      </div>
    </DialogFooter>
  </div>
</div>
