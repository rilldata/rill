<script lang="ts">
  import FileInput from "@rilldata/web-common/components/forms/FileInput.svelte";

  let singleValue: string | undefined = undefined;
  let multipleValue: string[] | undefined = undefined;

  $: console.log(singleValue, multipleValue);
</script>

<div class="flex flex-col gap-y-4 m-4">
  <FileInput
    value={singleValue}
    error="ERROR!"
    onInput={(files) => {
      singleValue = files[0]?.name;
    }}
  />

  <FileInput
    value={multipleValue}
    error={{ 1: "Error!" }}
    multiple
    onInput={(files) => {
      multipleValue = files.map((f) => f.name);
    }}
  />
</div>
