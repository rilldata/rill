---
title: User Management
sidebar_label: User Management 
sidebar_position: 21
---

In Rill Cloud, there are several levels of user management:
1. **Organization members and guests**: a user must have an organization-level role to access anything in an organization.
2. **Project roles**: organization members can have direct roles on a project. For example, the creator of a project automatically becomes an *admin* on the project.
3. **User group members and roles**: organization members can belong to user groups. When a user group has a role on a project, the role automatically propagates to all members of the user group. By default, a system-managed group consisting of all organization members (but not guests) is added to projects with the *viewer* role.

:::info

For a detailed breakdown of access permissions at different levels, see the [Roles and Permissions](roles-permissions) page!

:::

:::tip More Rill Cloud workflows coming

We have begun releasing new features around user management via Rill Cloud. If you'd like to learn more, please feel free to [reach out](/contact)!

:::

## Organization Users

At the organization-level, users can have one of the following roles:

- **Admins** can deploy and manage projects and manage billing.
- **Editors** can manage non-admin members of the organization by granting and revoking access.
- **Viewers** can browse the projects they have been given direct access to or indirect access to through user group memberships.
- **Guests** are similar to viewers, but not part of the "all members" usergroup which is added to projects by default.

For a detailed list of permissions, please refer to the [Roles and Permissions](roles-permissions).

### How to Add an Organization User
Admins can be invited to an organization from the *users* page, or via the CLI.

#### From Rill Cloud User page
From the organization page, you can manage users under the *Users* tab. Adding users from this page will add the user to the organization.

<img src = '/img/manage/user-management/add-user-cloud.png' class='rounded-gif' />
<br />


#### Via the CLI
```
rill user add
? Select role  [Use arrows to move, type to filter]
> admin
  editor
  viewer
  guest
```
You will then be prompted for details about the user.

:::tip Check your inbox (or spam)
If you add a user who has not yet signed up for Rill, they will receive an email inviting them to sign up and join.
:::

### How to add a user to a user group
If you have already set up a user group, instead of setting up users individually, you can add them to a user group.
```
rill user add --group <group name>
? Enter email <email here>
User "<email here>" added to the user group "<group name>"
```

### Automatically add members by email domain

You can automatically add users to your organization by their email domain. During the deployment process and in the Organization settings page. This is limited to the same domain of your user email. If you want to whitelist other domains, contact us! 

<img src = '/img/manage/user-management/rill-org-settings.png' class='rounded-gif' />
<br />


 For example, if you whitelist `yourdomain.com`, new and existing users with an email address ending on `@yourdomain.com` will automatically be added to your organization.

:::info Interested in whitelisting a different domain?

The feature currently requires manual action by a support representative at Rill. Just [reach out here](https://www.rilldata.com/contact) and ask us to whitelist your domain.

:::


## Project Users

Access to projects are managed at the individual project level subject to some notable rules:

1. By default, all organization members (but not guests) are added to new projects through a user group membership with the **viewer** role. You can manually remove this relationship in the project's member settings.
2. If you grant a project level role to someone who is not a member of the parent organization, they will automatically be added to the organization with the **guest** role.
3. Removing an organization member or guest automatically also removes them from all projects in the organization.
4. Organization admins implicitly have admin privileges on all projects in the organization.

### Project roles

Project users can have one of three roles:

- **Viewers** can browse the specific project, view dashboards, and setup alerts and reports
- **Editors** can add and remove non-admin project members
- **Admins** can manage the project by updating the project files, configuring environment variables, and accessing the status page

For a detailed list of permissions, please refer to the [Roles and Permissions](roles-permissions).

### How to add a Project User
There are a few ways to add a project user to Rill Cloud.
1. Admin invites user to the project using `Share`.
2. User requests access via the project URL.`https://ui.rilldata.com/<project_name>`
3. Administrator invites user via the CLI with `--project <project_name>` flag.

### Admin invites user from Rill Cloud

From the project's splash screen, please select share and type the email[s] along with the type of permissions.

<img src = '/img/manage/user-management/share-project.png' class='rounded-gif' />
<br />

Once sent, your invited users will receive this email and will need to accept it to view the project.

<img src = '/img/manage/user-management/email-invite.png' class='rounded-gif' />
<br />

### User requests access via URL

Alternatively, if you provide the project URL to your users, they can request access to the group admin. Users can request access via the page below:

<img src = '/img/manage/user-management/request-access.png' class='rounded-gif' />
<br />


The admin would receive an email to allow access, and can set the permission after accepting the request via the UI.

<img src = '/img/manage/user-management/admin-reply.png' class='rounded-gif' />
<br />


---

### Admin invites user via the CLI
To add a member to a project, run the following command:
```
rill user add --project [PROJECT NAME]
```
You will then be prompted for details about the user. HINT: Run `rill project list` to show available projects.

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to join.

#### Other actions

Run `rill user --help` to show commands for listing members or changing access.

## User Groups 

Another way to manage access is via user groups. You use the Rill CLI to create user groups and add members to them. Once you have created a user group, you can assign roles to it at the organization or project level, similar to how you assign roles to individual users.

User groups are scoped to an organization. They cannot be created only for a single project. Only organization members can be added to user groups, and user groups can only be added to projects within the organization they were created.

For more information on setting up user groups, see [user group permissions](usergroup-management).

### How to add a user to a Usergroup
There are two ways to add a user to a user group.
1. Admin adds them via Rill Cloud (Coming soon!)
2. Admin adds them via the CLI

#### Adds a user to a user group in Rill Cloud

### Managing Users via Rill Cloud
There are two ways that a user can get access to Rill Cloud. 

**Organization invites from Admin**
From the Users page on the Organization page, you can inivte a user to the organization. Please note that organization viewers have access to view all projects. 

<img src = '/img/tutorials/admin/org-user-management.png' class='rounded-gif' />
<br />
**Project level access requests**

  Please refer to the <a href='https://docs.rilldata.com/manage/user-management#admin-invites-user' target = "blank">documentation how a user can request access to project, or how an admin can invite a user to the project. </a>



#### Adds a user to a user group with the Rill CLI

```
rill user add --group <group-name>
```
You will then be prompted for details on the user. 

:::note
If the user you are trying to add is not part of the organization yet, the CLI will prompt you to add them to the organization first then proceed to adding them to a group.
:::

To see the current members of a group:

```
rill user list --group <group-name>
```

To find the current user group roles, with project flag if looking for specific project's role:

```
rill usergroup list [--project my_project_name]
```

## Which privilege wins?

Rill uses a logical **OR** operand to define the winning privilege. In other words, if any direct role or indirect role through a usergroup allows a user to take an action, the action will succeed.

## Logging into Rill Cloud

In order to access a deployed project and/or view a shared dashboard, users will need to first login to [Rill Cloud](https://ui.rilldata.com/). When you first navigate to https://ui.rilldata.com/, you will see a few different options to login, including:
- Google SSO
- Microsoft SSO
- Email _(basic auth)_

:::info SAML Authentication

Rill Cloud **does** support SAML authentication for our enterprise customers. If this is a requirement, [please get in contact](/contact) with us and we can discuss appropriate next steps to help you with your setup.

:::

If this is the first time you are accessing Rill Cloud, you will want to sign up instead.

<img src = '/img/manage/user-management/sign-up.png' class='rounded-gif' />
<br />


:::tip Signing up with basic auth

If you are unsure which option to select, select `Continue with Email` and set up basic authentication (email address / password).

:::

Afterwards, you should receive an email verification to complete the sign up process. 

<img src = '/img/manage/user-management/verification-email.png' class='rounded-gif' />
<br />

You should now be authenticated with Rill Cloud and be able to sign-in directly going forward!

