<script lang="ts">
  import type { PageData } from "./$types";
  import ImportTableStatus from "@rilldata/web-common/features/add-data/import/ImportTableStatus.svelte";
  import { Button } from "@rilldata/web-common/components/button/index.ts";

  export let data: PageData;
  $: ({ runner } = data);
</script>

<ImportTableStatus {runner} />
<div class="w-fit mx-auto mb-4">
  <Button type="secondary" href="/" fit>Skip and view project</Button>
</div>
