<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";

  // TODO: implement
  function handleCreateOrganization() {
    console.log("Create organization");
  }
</script>

<div class="flex flex-col justify-center items-center h-3/5">
  <h1 class="text-2xl font-bold">You're lonely!</h1>
  <p class="mt-4 text-lg">You are not yet part of a Rill organization.</p>
  <div class="mt-4">
    <Button type="primary" on:click={handleCreateOrganization}
      >Create organization</Button
    >
  </div>
</div>
