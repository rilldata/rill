<script>
  import { Button } from "@rilldata/web-local/lib/components/button";
  import { uploadFilesWithDialog } from "@rilldata/web-local/lib/util/file-upload";
  import { createEventDispatcher } from "svelte";

  const dispatch = createEventDispatcher();

  async function handleUpload() {
    await uploadFilesWithDialog();
    dispatch("close");
  }
</script>

<div class="grid place-items-center h-full">
  <Button on:click={handleUpload} type="primary"
    >Upload a CSV or Parquet file</Button
  >
</div>
