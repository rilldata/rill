<script lang="ts">
  import Input from "@rilldata/web-common/components/forms/Input.svelte";

  export let properties: any[];
  export let formId: string;
  export let form: any;
  export let errors: any;
  export let enhance: any;
  export let submit: any;
</script>

<form
  id={formId}
  class="pb-5 flex-grow overflow-y-auto"
  use:enhance
  on:submit|preventDefault={submit}
>
  {#each properties as property (property.key)}
    {@const propertyKey = property.key ?? ""}
    <div class="py-1.5 first:pt-0 last:pb-0">
      <Input
        id={propertyKey}
        label={property.displayName}
        placeholder={property.placeholder}
        secret={property.secret}
        hint={property.hint}
        errors={errors[propertyKey]}
        bind:value={$form[propertyKey]}
        alwaysShowError
      />
    </div>
  {/each}
</form>
