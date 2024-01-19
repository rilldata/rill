<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { createForm } from "svelte-forms-lib";
  import { Button } from "../../components/button";
  import Dialog from "../../components/dialog-v2/Dialog.svelte";

  export let open: boolean;

  const dispatch = createEventDispatcher();

  createForm({
    initialValues: {},
    onSubmit: async (values) => {
      console.log("submitting alerts form with these values: ", values);
      dispatch("close");
    },
  });

  // const { isSubmitting, form } = formState;
</script>

<Dialog {open}>
  <svelte:fragment slot="title">Create alert</svelte:fragment>
  <svelte:fragment slot="body"></svelte:fragment>
  <svelte:fragment slot="footer">
    <div class="flex items-center gap-x-2 mt-5">
      <div class="grow" />
      <Button on:click={() => dispatch("close")} type="secondary">
        Cancel
      </Button>
      <Button form="create-alert-form" submitForm type="primary">Create</Button>
    </div>
  </svelte:fragment>
</Dialog>
