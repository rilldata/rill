<script lang="ts">
  import FileInput from "@rilldata/web-common/components/forms/FileInput.svelte";

  let singleValue: string | undefined = undefined;
  let multipleValue: string[] | undefined = undefined;

  $: console.log(singleValue, multipleValue);
</script>

<FileInput
  value={singleValue}
  onInput={(files) => {
    singleValue = files[0]?.name;
  }}
/>

<FileInput
  value={multipleValue}
  multiple
  onInput={(files) => {
    multipleValue = files.map((f) => f.name);
  }}
/>
