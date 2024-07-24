<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
    AlertDialogTrigger,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";

  export let open = false;
  export let loading = false;
  export let onConnect: () => void;
</script>

<AlertDialog bind:open>
  <AlertDialogTrigger asChild let:builder>
    <Button on:click={onConnect} builders={[builder]} type="primary" {loading}>
      Connect
    </Button>
  </AlertDialogTrigger>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>
        Are you sure you want to connect to this repo?
      </AlertDialogTitle>
      <AlertDialogDescription>
        We will push this project’s data to this repo, replacing any current
        contents.
      </AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter class="mt-5">
      <Button type="secondary" on:click={() => (open = false)}>Back</Button>
      <Button
        type="primary"
        on:click={() => {
          open = false;
          onConnect();
        }}
      >
        Yes, connect
      </Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
