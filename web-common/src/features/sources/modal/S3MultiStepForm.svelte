<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";
  import MultiStepAuthForm from "./MultiStepAuthForm.svelte";
  import { normalizeErrors } from "./utils";
  import { S3_AUTH_OPTIONS, type S3AuthMethod } from "./constants";

  export let properties: any[] = [];
  export let paramsForm: any;
  export let paramsErrors: Record<string, any>;
  export let onStringInputChange: (e: Event) => void;
  export let handleFileUpload: (file: File) => Promise<string>;

  const filteredParamsProperties = properties;

  const S3_CLEAR_FIELDS: Record<S3AuthMethod, string[]> = {
    access_keys: ["aws_role_arn", "aws_role_session_name", "aws_external_id"],
    role: ["aws_access_key_id", "aws_secret_access_key"],
  };

  const S3_EXCLUDED_KEYS = [
    "aws_access_key_id",
    "aws_secret_access_key",
    "aws_role_arn",
    "aws_role_session_name",
    "aws_external_id",
  ];
</script>

<MultiStepAuthForm
  properties={filteredParamsProperties}
  {paramsForm}
  {paramsErrors}
  {onStringInputChange}
  {handleFileUpload}
  authOptions={S3_AUTH_OPTIONS}
  defaultAuthMethod="access_keys"
  clearFieldsByMethod={S3_CLEAR_FIELDS}
  excludedKeys={S3_EXCLUDED_KEYS}
>
  <svelte:fragment slot="auth-fields" let:option let:paramsErrors>
    {#if option.value === "access_keys"}
      <div class="space-y-3">
        <Input
          id="aws_access_key_id"
          label="Access Key ID"
          placeholder="Enter AWS access key ID"
          optional={false}
          secret={true}
          hint="AWS access key ID for the bucket"
          errors={normalizeErrors(paramsErrors.aws_access_key_id)}
          bind:value={$paramsForm.aws_access_key_id}
          alwaysShowError
        />
        <Input
          id="aws_secret_access_key"
          label="Secret Access Key"
          placeholder="Enter AWS secret access key"
          optional={false}
          secret={true}
          hint="AWS secret access key for the bucket"
          errors={normalizeErrors(paramsErrors.aws_secret_access_key)}
          bind:value={$paramsForm.aws_secret_access_key}
          alwaysShowError
        />
      </div>
    {:else if option.value === "role"}
      <div class="space-y-3">
        <Input
          id="aws_role_arn"
          label="Role ARN"
          placeholder="Enter AWS IAM role ARN"
          optional={false}
          secret={true}
          hint="Role ARN to assume for accessing the bucket"
          errors={normalizeErrors(paramsErrors.aws_role_arn)}
          bind:value={$paramsForm.aws_role_arn}
          alwaysShowError
        />
        <Input
          id="aws_role_session_name"
          label="Role session name"
          placeholder="Optional session name (defaults to rill-session)"
          optional={true}
          secret={false}
          errors={normalizeErrors(paramsErrors.aws_role_session_name)}
          bind:value={$paramsForm.aws_role_session_name}
          alwaysShowError
        />
        <Input
          id="aws_external_id"
          label="External ID"
          placeholder="Optional external ID for cross-account access"
          optional={true}
          secret={true}
          errors={normalizeErrors(paramsErrors.aws_external_id)}
          bind:value={$paramsForm.aws_external_id}
          alwaysShowError
        />
      </div>
    {/if}
  </svelte:fragment>
</MultiStepAuthForm>
