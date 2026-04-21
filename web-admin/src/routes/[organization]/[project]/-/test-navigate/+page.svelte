<script lang="ts">
  import { page } from "$app/state";
  import { goto } from "$app/navigation";
  import { Button } from "@rilldata/web-common/components/button/index.ts";

  let organization = $derived(page.params.organization);
  let project = $derived(page.params.project);

  async function navigate() {
    return goto(`/${organization}/${project}/@dev/-/deploying`);
  }
</script>

<Button onClick={navigate}>Test navigate</Button>
