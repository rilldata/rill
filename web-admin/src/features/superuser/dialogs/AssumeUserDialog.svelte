<!-- Confirms before starting a session as another user. Applies to the three
     call sites in the superuser console: Users page, Org Members table, and
     Org Projects list. -->
<script lang="ts">
  import {
    AlertDialog,
    AlertDialogContent,
    AlertDialogDescription,
    AlertDialogFooter,
    AlertDialogHeader,
    AlertDialogTitle,
  } from "@rilldata/web-common/components/alert-dialog";
  import { Button } from "@rilldata/web-common/components/button";
  import { assumedUser } from "@rilldata/web-admin/features/superuser/users/assume-state";

  export let open = false;
  export let email: string;
  export let redirect: string | undefined = undefined;
  export let contextLabel: string | undefined = undefined;

  $: description = contextLabel
    ? `You will start browsing Rill Cloud as ${email}, landing on "${contextLabel}". The session will expire after 60 minutes. Use the banner to unassume when done.`
    : `You will start browsing Rill Cloud as ${email}. The session will expire after 60 minutes. Use the banner to unassume when done.`;

  function handleConfirm() {
    assumedUser.assume(email, redirect ? { redirect } : {});
    open = false;
  }
</script>

<AlertDialog bind:open>
  <AlertDialogContent>
    <AlertDialogHeader>
      <AlertDialogTitle>Open as User</AlertDialogTitle>
      <AlertDialogDescription>{description}</AlertDialogDescription>
    </AlertDialogHeader>
    <AlertDialogFooter>
      <Button type="tertiary" onClick={() => (open = false)}>Cancel</Button>
      <Button type="primary" onClick={handleConfirm}>Open as User</Button>
    </AlertDialogFooter>
  </AlertDialogContent>
</AlertDialog>
