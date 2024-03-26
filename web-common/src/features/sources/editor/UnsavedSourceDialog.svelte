<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "../../../components/button";
  import { Dialog } from "../../../components/modal";
  import DialogFooter from "../../../components/modal/dialog/DialogFooter.svelte";

  export let context: string = "source";

  const dispatch = createEventDispatcher();
</script>

<Dialog
  on:cancel={() => dispatch("cancel")}
  size="sm"
  useContentForMinSize
  focusTriggerOnClose={false}
>
  <svelte:fragment slot="title">Leave {context} without saving?</svelte:fragment
  >
  <div class="text-sm text-slate-500" slot="body">
    Navigating away will lose your changes.
  </div>
  <DialogFooter slot="footer">
    <Button type="secondary" on:click={() => dispatch("cancel")}
      >Keep editing</Button
    >
    <Button type="primary" on:click={() => dispatch("confirm")}
      >Yes, leave {context}</Button
    >
  </DialogFooter>
</Dialog>
