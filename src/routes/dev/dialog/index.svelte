<script lang="ts">
  import { Button } from "$lib/components/button";
  import { Dialog } from "$lib/components/modal";
  let replaceSource = false;
  let changeThing = false;
  let changeThingValue = "original change thing";
</script>

<div style:height="200vh">
  <button
    on:click={() => {
      replaceSource = !replaceSource;
    }}>Replace Source</button
  >
  <button
    on:click={() => {
      changeThing = !changeThing;
    }}>Change thing</button
  >

  {#if replaceSource}
    <Dialog
      showCancel
      on:cancel={() => (replaceSource = false)}
      on:submit={() => (replaceSource = !replaceSource)}
    >
      <svelte:fragment slot="title">Replace existing source?</svelte:fragment>
      <svelte:fragment slot="body"
        >This action will replace all existing measures and dimensions.</svelte:fragment
      >
      <svelte:fragment slot="footer">
        <div class="flex flex-row gap-x-3 justify-items-end justify-end">
          <Button on:click={() => (replaceSource = false)} type="text"
            >cancel</Button
          >
          <Button type="primary">Update source</Button>
        </div>
      </svelte:fragment>
    </Dialog>
  {/if}

  {#if changeThing}
    <Dialog
      showCancel
      on:cancel={() => {
        changeThing = false;
      }}
    >
      <svelte:fragment slot="title"
        >Change the name of this thing</svelte:fragment
      >
      <svelte:fragment slot="body">
        <input
          value={changeThingValue}
          on:change={(event) => {
            changeThingValue = event.target.value;
          }}
        />
        {changeThingValue}
      </svelte:fragment>
      <svelte:fragment slot="footer">
        <div class="flex flex-row gap-x-3 justify-items-end justify-end">
          <Button on:click={() => (changeThing = false)} type="text"
            >cancel</Button
          >
          <Button type="primary">Update name</Button>
        </div>
      </svelte:fragment>
    </Dialog>
  {/if}
</div>

<div class="fixed left-8 top-16" style:width="300px" style:height="300px">
  buncha other stuff!!!
</div>
