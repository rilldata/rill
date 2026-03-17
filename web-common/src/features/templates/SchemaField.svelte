<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import InformationalField from "@rilldata/web-common/components/forms/InformationalField.svelte";
  import Checkbox from "@rilldata/web-common/components/forms/Checkbox.svelte";
  import Radio from "@rilldata/web-common/components/forms/Radio.svelte";
  import Select from "@rilldata/web-common/components/forms/Select.svelte";
  import CredentialsInput from "@rilldata/web-common/components/forms/CredentialsInput.svelte";
  import KeyValueInput from "@rilldata/web-common/components/forms/KeyValueInput.svelte";
  import { normalizeErrors } from "./error-utils";
  import { getFileAccept } from "./file-encoding";
  import type { JSONSchemaField } from "./schemas/types";

  export let id: string;
  export let prop: JSONSchemaField;
  export let optional: boolean;
  export let errors: any;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (
    file: File,
    fieldKey: string,
  ) => Promise<string>;
  export let value: any;
  export let checked: boolean | undefined;
  export let options:
    | Array<{ value: string; label: string; description?: string }>
    | undefined;
  export let name: string | undefined;
  export let disabled: boolean = false;
</script>

{#if prop["x-informational"]}
  <InformationalField
    description={prop.description}
    hint={prop["x-hint"]}
    href={prop["x-docs-url"]}
  />
{:else if prop["x-display"] === "file" || prop.format === "file"}
  <CredentialsInput
    {id}
    label={prop.title ?? id}
    hint={prop.description ?? prop["x-hint"]}
    {optional}
    bind:value
    uploadFile={(file) => handleFileUpload(file, id)}
    accept={getFileAccept(prop)}
  />
{:else if prop.type === "boolean"}
  <Checkbox
    {id}
    bind:checked
    label={prop.title ?? id}
    hint={prop.description ?? prop["x-hint"]}
    {optional}
    {disabled}
  />
{:else if prop["x-display"] === "key-value"}
  <KeyValueInput
    {id}
    label={prop.title ?? id}
    hint={prop.description ?? prop["x-hint"]}
    {optional}
    bind:value
    keyPlaceholder={prop["x-placeholder"]}
  />
{:else if options?.length && prop["x-display"] === "radio"}
  <Radio bind:value {options} {name} {disabled} />
{:else if options?.length}
  <Select
    {id}
    label={prop.title ?? id}
    tooltip={prop.description ?? prop["x-hint"] ?? ""}
    {options}
    {optional}
    placeholder={prop["x-placeholder"] ?? ""}
    bind:value
    full
    sameWidth
    clearable
    {disabled}
  />
{:else}
  <Input
    {id}
    label={prop.title ?? id}
    placeholder={prop["x-placeholder"]}
    {optional}
    secret={prop["x-secret"]}
    hint={prop.description ?? prop["x-hint"]}
    errors={normalizeErrors(errors)}
    bind:value
    multiline={prop["x-display"] === "textarea"}
    fontFamily={prop["x-monospace"] ? "monospace" : "inherit"}
    onInput={(_, e) => onStringInputChange(e)}
    alwaysShowError
    {disabled}
  />
{/if}
