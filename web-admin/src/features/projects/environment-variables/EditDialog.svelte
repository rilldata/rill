<script lang="ts">
  import {
    Dialog,
    DialogContent,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
  } from "@rilldata/web-common/components/dialog-v2";
  import Button from "@rilldata/web-common/components/button/Button.svelte";

  export let open = false;

  // let searchText = "";

  // $: organization = $page.params.organization;
  // $: listUsergroupMemberUsers = createAdminServiceListUsergroupMemberUsers(
  //   organization,
  //   groupName,
  // );

  // const queryClient = useQueryClient();
  // const removeUserGroupMember = createAdminServiceRemoveUsergroupMemberUser();
  // const addUsergroupMemberUser = createAdminServiceAddUsergroupMemberUser();
  // const renameUserGroup = createAdminServiceRenameUsergroup();

  // async function handleAddUsergroupMemberUser(
  //   email: string,
  //   usergroup: string,
  // ) {
  //   try {
  //     await $addUsergroupMemberUser.mutateAsync({
  //       organization: organization,
  //       usergroup: usergroup,
  //       email: email,
  //       data: {},
  //     });

  //     await queryClient.invalidateQueries(
  //       getAdminServiceListOrganizationMemberUsersQueryKey(organization),
  //     );

  //     await queryClient.invalidateQueries(
  //       getAdminServiceListUsergroupMemberUsersQueryKey(
  //         organization,
  //         usergroup,
  //       ),
  //     );

  //     eventBus.emit("notification", {
  //       message: "User added to user group",
  //     });
  //   } catch (error) {
  //     console.error("Error adding user to user group", error);
  //     eventBus.emit("notification", {
  //       message: "Error adding user to user group",
  //       type: "error",
  //     });
  //   }
  // }

  // async function handleRename(groupName: string, newName: string) {
  //   try {
  //     await $renameUserGroup.mutateAsync({
  //       organization: organization,
  //       usergroup: groupName,
  //       data: {
  //         name: newName,
  //       },
  //     });

  //     await queryClient.invalidateQueries(
  //       getAdminServiceListOrganizationMemberUsergroupsQueryKey(organization),
  //     );

  //     eventBus.emit("notification", { message: "User group renamed" });
  //   } catch (error) {
  //     console.error("Error renaming user group", error);
  //     eventBus.emit("notification", {
  //       message: "Error renaming user group",
  //       type: "error",
  //     });
  //   }
  // }

  // async function handleRemoveUser(groupName: string, email: string) {
  //   try {
  //     await $removeUserGroupMember.mutateAsync({
  //       organization: organization,
  //       usergroup: groupName,
  //       email: email,
  //     });

  //     await queryClient.invalidateQueries(
  //       getAdminServiceListUsergroupMemberUsersQueryKey(
  //         organization,
  //         groupName,
  //       ),
  //     );

  //     eventBus.emit("notification", {
  //       message: "User removed from user group",
  //     });
  //   } catch (error) {
  //     console.error("Error removing user from user group", error);
  //     eventBus.emit("notification", {
  //       message: "Error removing user from user group",
  //       type: "error",
  //     });
  //   }
  // }

  // // We don't want to show users that are already in the group
  // $: availableSearchUsersList = searchUsersList.filter(
  //   (user) =>
  //     !$listUsergroupMemberUsers.data?.members.some(
  //       (member) => member.userEmail === user.userEmail,
  //     ),
  // );

  const formId = "edit-environment-variable-form";

  // const initialValues = {
  //   newName: groupName,
  // };

  // const schema = yup(
  //   object({
  //     newName: string()
  //       .required("New user group name is required")
  //       .min(3, "New user group name must be at least 3 characters")
  //       .matches(
  //         /^[a-z0-9]+(-[a-z0-9]+)*$/,
  //         "New user group name must be lowercase and can contain letters, numbers, and hyphens (slug)",
  //       ),
  //   }),
  // );

  // const { form, enhance, submit, errors, submitting } = superForm(
  //   defaults(initialValues, schema),
  //   {
  //     SPA: true,
  //     validators: schema,
  //     async onUpdate({ form }) {
  //       if (!form.valid) return;
  //       const values = form.data;

  //       try {
  //         await handleRename(groupName, values.newName);
  //         open = false;
  //       } catch (error) {
  //         console.error(error);
  //       }
  //     },
  //   },
  // );

  // $: coercedUsersToOptions = availableSearchUsersList.map((user) => ({
  //   value: user.userEmail,
  //   label: user.userEmail,
  //   name: user.userName,
  // }));
</script>

<Dialog
  bind:open
  onOutsideClick={(e) => {
    e.preventDefault();
    open = false;
  }}
>
  <DialogTrigger asChild>
    <div class="hidden"></div>
  </DialogTrigger>
  <DialogContent class="translate-y-[-200px]">
    <DialogHeader>
      <DialogTitle>Edit environment variable</DialogTitle>
    </DialogHeader>

    <DialogFooter>
      <Button
        type="plain"
        on:click={() => {
          open = false;
        }}>Cancel</Button
      >
      <Button type="primary" form={formId} submitForm>Save</Button>
    </DialogFooter>
  </DialogContent>
</Dialog>
