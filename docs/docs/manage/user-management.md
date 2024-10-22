---
title: User Management
sidebar_label: Users
sidebar_position: 22
---

In Rill Cloud, access can be granted at the organization, project, or group level using the Rill CLI. 

:::info

Note that the permissions may vary from each level, please review the [Roles and Permissions](roles-permissions.md) page for more information!

:::

:::tip More UI-based workflows coming

We have begun releasing new features around user managment via the UI. If you'd like to learn more, please feel free to [reach out](contact.md)!

:::

## Install and authenticate the Rill CLI

To manage cloud permissions with the Rill CLI, you must first authenticate it. If you have not already done so, run:
```
rill login
```

  
## Adding a member to the organization

When you invite a user to an organization on Rill Cloud, they automatically get access to *all projects* in the organization. Users can have one of two roles:

- **Viewers** can browse projects and view dashboards
- **Admins** can manage projects by deploying new projects, making changes to existing projects, or deleting deployed projects. They can also manage members of an organization by granting or revoking access to other users.  


### Add a member
To add a member to an organization, run the following command:
```
rill user add
```
You will then be prompted for details about the user.

:::tip Check your inbox (or spam)

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to sign up and join.

:::

### Automatically add members by email domain

You can automatically add users to your organization by their email domain. For example, if you whitelist `yourdomain.com`, new and existing users with an email address ending on `@yourdomain.com` will automatically be added to your organization.

:::info Interested in whitelisting a domain?

The feature currently requires manual action by a support representative at Rill. Just [reach out here](https://www.rilldata.com/contact) and ask us to whitelist your domain.

:::

### Other actions

Run `rill user --help` to show commands for listing members or changing access.

## Adding a member to a specific project

:::tip Did you know?

Starting from version 0.48, you can add a user to a specific project via the UI!

:::

### Via the UI

#### Option 1 - Admin invites user

From the project's splash screen, please select share and type the email[s] along with the type of permissions.

![img](/img/manage/user-management/share-project.png)

Once sent, your invited users will receive this email and will need to accept it to view the project.

![img](/img/manage/user-management/email-invite.png)

#### Option 2 - User requests access

Alternatively, if you provide the project URL to your users, they can request access to the group admin. Users can request access via the page below:

![img](/img/manage/user-management/request-access.png)

The admin would receive an email to allow access, and can set the permission after accepting the request via the UI.

![img](/img/manage/user-management/admin-reply.png)

---

### Via the CLI
By default, adding a user to an organization grants them access to all its projects. You can alternatively add a user only to a specific project. Users can have one of two roles on a project:

- **Viewers** can view the project's dashboards
- **Admins** can additionally edit the project, and view and edit project members

#### Add a member

To add a member to a project, run the following command:
```
rill user add --project [PROJECT NAME]
```
You will then be prompted for details about the user. HINT: Run `rill project list` to show available projects.

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to join.

#### Other actions

Run `rill user --help` to show commands for listing members or changing access.


## Adding a member to a specific user group

Another way to manage user is via User groups. Users can be added to a group using the following from the Rill CLI. You can define the usergroup to have certain permissions on specific groups.

- **Viewers** can view the project's dashboards
- **Admins** can additionally edit the project, and view and edit project members


### Add a member 
```
rill user add --group <group-name>
```
You will then be prompted for details on the user. 

:::note

Currently, only users part of the organization can be added to user groups. You will need to wait till after they accept the invitation to add them to a group.
:::

To see the current members of a group:

```
rill user list --group <group-name>
```

To find the current user group roles, with project flag if looking for specific project's role:

```
rill usergroup list <--project my_project_name>
```
### Other actions
Run `rill usergroup --help` to show commands for listing usergroups or changing access.


## Which privilege wins?

Rill uses a logical **OR** operand to define the winning privilege. In other words, whichever has the higher privilege will be applied. See below for some example situations that may arise.
<div>
| # | Organization | Project (*not required*)   | Group  (*not required*)| Resulting privilege |
|---|:--------------|:--------------|--------|---------------------:|
| 1 | admin        | viewer        | ---    | admin               |
| 2 | viewer       | admin         | viewer | admin               |
| 3 | viewer       | ---           | viewer | viewer              |
| 4 | viewer       | viewer           | --- | viewer              |

</div>
## Logging into Rill Cloud

In order to access a deployed project and/or view a shared dashboard, users will need to first login to [Rill Cloud](https://ui.rilldata.com/). When you first navigate to https://ui.rilldata.com/, you will see a few different options to login, including:
- Google SSO
- Microsoft SSO
- Email _(basic auth)_

:::info SAML Authentication

Rill Cloud **does** support SAML authentication for our enterprise customers. If this is a requirement, [please get in contact](contact.md) with us and we can discuss appropriate next steps to help you with your setup.

:::

If this is the first time you are accessing Rill Cloud, you will want to sign up instead.

![Signing Up](/img/manage/user-management/sign-up.png)

:::tip Signing up with basic auth

If you are unsure which option to select, select `Continue with Email` and set up basic authentication (email address / password).

:::

Afterwards, you should receive an email verification to complete the sign up process. 

![Verification Email](/img/manage/user-management/verification-email.png)

You should now be authenticated with Rill Cloud and be able to sign-in directly going forward!

