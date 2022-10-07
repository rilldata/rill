<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import * as yup from "yup";
  import {
    GCS,
    GCSYupSchema,
    HTTP,
    HTTPYupSchema,
    S3,
    S3YupSchema,
  } from "../../../connectors/schemas";
  import { Button } from "../../button";
  import Input from "../../Input.svelte";
  import DialogFooter from "../../modal/dialog/DialogFooter.svelte";
  import Tab from "../../tab/Tab.svelte";
  import TabGroup from "../../tab/TabGroup.svelte";

  let selectedConnector = "S3";
  let connectorSpec = S3;
  let yupSchema = S3YupSchema;

  const dispatch = createEventDispatcher();

  function extendYupSchemaWithSourceName(yupSchema: yup.AnyObjectSchema) {
    return yupSchema.concat(
      yup.object().shape({
        sourceName: yup.string().required(),
      })
    );
  }

  yupSchema = extendYupSchemaWithSourceName(yupSchema);

  function onConnectorChange(selectedConnector: string) {
    if (selectedConnector === "S3") {
      connectorSpec = S3;
      yupSchema = S3YupSchema;
    } else if (selectedConnector === "GCS") {
      connectorSpec = GCS;
      yupSchema = GCSYupSchema;
    } else if (selectedConnector === "HTTP") {
      connectorSpec = HTTP;
      yupSchema = HTTPYupSchema;
    }

    yupSchema = extendYupSchemaWithSourceName(yupSchema);
  }

  $: onConnectorChange(selectedConnector);

  function compileCreateSourceSql(values) {
    const compiledKeyValues = Object.entries(values)
      .map(([key, value]) => `${key}='${value}'`)
      .join(", ");

    return (
      `CREATE SOURCE ${values.sourceName} WITH (connector = '${connectorSpec.name}', ` +
      compiledKeyValues +
      `)`
    );
  }

  const { form, errors, handleSubmit } = createForm({
    // TODO: initialValues should come from SQL asset and be reactive to asset modifications
    initialValues: {},
    validationSchema: yupSchema,
    onSubmit: (values) => {
      const sql = compileCreateSourceSql(values);
      // TODO: dispatch sql to SQL editor
      // TODO: submit sql to backend
      alert(sql);
    },
  });
</script>

<div class="h-full flex flex-col">
  <div class="pt-4 px-4">
    <TabGroup
      variant="secondary"
      on:select={(event) => {
        selectedConnector = event.detail;
      }}
    >
      <Tab value={"S3"}>S3</Tab>
      <Tab value={"GCS"}>GCS</Tab>
      <Tab value={"HTTP"}>https</Tab>
    </TabGroup>
  </div>
  <div class="p-4">
    {@html connectorSpec.description}
  </div>
  {#key selectedConnector}
    <div class="px-4 flex-grow overflow-y-auto">
      <form
        on:submit={handleSubmit}
        id="remote-source-{selectedConnector}-form"
      >
        <div class="pb-2">
          <div class="py-2">
            <Input
              label="Source name"
              bind:value={$form["sourceName"]}
              error={$errors["sourceName"]}
              placeholder="my_new_source"
            />
          </div>
          {#each Object.entries(connectorSpec.fields) as [name, attributes]}
            {@const label =
              attributes.label + (attributes.required ? "" : " (optional)")}
            <div class="py-2">
              {#if attributes.type === "text"}
                <Input
                  id={name}
                  {label}
                  error={$errors[name]}
                  bind:value={$form[name]}
                  placeholder={attributes.placeholder}
                />
              {/if}
              {#if attributes.type === "checkbox"}
                <label for={name} class="flex items-center">
                  <input
                    id={name}
                    type="checkbox"
                    bind:checked={$form[name]}
                    class="h-5 w-5"
                  />
                  <span class="ml-2 text-sm">{label}</span>
                </label>
              {/if}
            </div>
          {/each}
        </div>
      </form>
    </div>
    <div class="bg-gray-100 border-t border-gray-300">
      <DialogFooter>
        <Button
          on:click={() => {
            dispatch("cancel");
          }}
          type="text"
        >
          Cancel
        </Button>
        <Button
          type="primary"
          submitForm
          form="remote-source-{selectedConnector}-form"
        >
          Add source
        </Button>
      </DialogFooter>
    </div>
  {/key}
</div>
