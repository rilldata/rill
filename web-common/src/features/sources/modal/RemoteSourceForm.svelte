<script lang="ts">
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import SubmissionError from "@rilldata/web-common/components/forms/SubmissionError.svelte";
  import {
    ConnectorDriverPropertyType,
    RpcStatus,
    V1ConnectorDriver,
  } from "@rilldata/web-common/runtime-client";
  import { useQueryClient } from "@tanstack/svelte-query";
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { overlay } from "../../../layout/overlay-store";
  import { directoryState } from "../../file-explorer/directory-store";
  import { inferSourceName } from "../sourceUtils";
  import { humanReadableErrorMessage } from "./errors";
  import { submitRemoteSourceForm } from "./submitRemoteSourceForm";
  import { getYupSchema, toYupFriendlyKey } from "./yupSchemas";

  export let connector: V1ConnectorDriver;

  const queryClient = useQueryClient();
  const dispatch = createEventDispatcher();

  let rpcError: RpcStatus | null = null;

  const { form, touched, errors, handleChange, handleSubmit, isSubmitting } =
    createForm({
      initialValues: {
        sourceName: "", // avoids `values.sourceName` warning
      },
      validationSchema: getYupSchema(connector),
      onSubmit: async (values) => {
        overlay.set({ title: `Importing ${values.sourceName}` });
        try {
          // the following error provides type narrowing for `connector.name`
          if (connector.name === undefined)
            throw new Error("connector name is undefined");
          await submitRemoteSourceForm(queryClient, connector.name, values);
          await goto(`/files/sources/${values.sourceName}`);
          directoryState.expand("sources");
          dispatch("close");
        } catch (e) {
          rpcError = e?.response?.data;
        }
        overlay.set(null);
      },
    });

  // Place the "Source name" field directly under the "Path" field, which is the first property for each connector (s3, gcs, https).
  const connectorProperties = [
    ...(connector.sourceProperties?.slice(0, 1) ?? []),
    {
      key: "sourceName",
      displayName: "Source name",
      description: "The name of the source",
      placeholder: "my_new_source",
      type: ConnectorDriverPropertyType.TYPE_STRING,
      required: true,
    },
    ...(connector.sourceProperties?.slice(1) ?? []),
  ];

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
    class="pb-5 flex-grow overflow-y-auto"
    id="remote-source-{connector.name}-form"
    on:submit|preventDefault={handleSubmit}
  >
    <div class="pb-2 text-slate-500">
      Need help? Refer to our
      <a
        href="https://docs.rilldata.com/build/connect"
        rel="noreferrer"
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

    {#each connectorProperties as property}
      {@const label =
        property.displayName + (property.required ? "" : " (optional)")}
      <div class="py-1.5">
        {#if property.type === ConnectorDriverPropertyType.TYPE_STRING && property.key !== undefined}
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
        {:else if property.type === ConnectorDriverPropertyType.TYPE_BOOLEAN && property.key !== undefined}
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
    {/each}
  </form>
  <div class="flex items-center space-x-2">
    <div class="grow" />
    <Button on:click={() => dispatch("back")} type="secondary">Back</Button>
    <Button
      disabled={$isSubmitting}
      form="remote-source-{connector.name}-form"
      submitForm
      type="primary"
    >
      Add source
    </Button>
  </div>
</div>
