---
title: User Management
sidebar_label: User Management 
sidebar_position: 21
---

In Rill Cloud, there are three types of users:
1. [**Organizational Users**](#organization-users): these users have at least viewer access to all projects within an organization.
2. [**Project Users**](#project-users): these users have at least viewer access to a single Rill project. These users **do not** need to be part of the organization.
3. [**User group Users**](#user-groups): These users are added to your organization and are given permissions via user groups. 

:::info

Note that the permissions may vary from each level, please review the [Roles and Permissions](roles-permissions.md) page for more information!

:::

:::tip More Rill Cloud workflows coming

We have begun releasing new features around user managment via Rill Cloud. If you'd like to learn more, please feel free to [reach out](contact.md)!

:::


  
## Organization Users

When you invite a user to an organization on Rill Cloud, they automatically get access to *all projects* in the organization. Users can have one of two roles:

- **Viewers** can browse projects and view dashboards
- **Admins** can manage projects by deploying new projects, making changes to existing projects, or deleting deployed projects. They can also manage members of an organization by granting or revoking access to other users.  

For a detailed list of permissions, please refer to the [Roles and Permissions](roles-permissions.md).

### How to add an Organization User
Administrators can be invited to an organization from the *users* page, or via the CLI.

### Administrator adds from Rill Cloud User page
From the organization page, you can manage users under the *Users* tab. Adding users from this page will add the user to the organization.

![img](/img/manage/user-management/add-user-cloud.png)


### Administrator invites user via the CLI
```
rill user add
? Select role  [Use arrows to move, type to filter]
> admin
  viewer
```
You will then be prompted for details about the user.

:::tip Check your inbox (or spam)
If you add a user who has not yet signed up for Rill, they will receive an email inviting them to sign up and join.
:::

### Administrator adds the user to a user group with organization access
If you have already set up a user group that has access on the organization level, instead of setting up users individually, you can [add them to the user group](#how-to-add-a-user-to-a-usergroup).
```
rill user add --group <group_with_org_permissions>
? Enter email <email here>
User "<email here>" added to the user group "<group_with_org_permissions>"
```

![] add image here

### Automatically add members by email domain

You can automatically add users to your organization by their email domain. During the deployment process and in the Organization settings page, you can input the domain to whitelist.

![img](/img/manage/user-management/domain.png)

 For example, if you whitelist `yourdomain.com`, new and existing users with an email address ending on `@yourdomain.com` will automatically be added to your organization.

:::info Interested in whitelisting a different domain?

The feature currently requires manual action by a support representative at Rill. Just [reach out here](https://www.rilldata.com/contact) and ask us to whitelist your domain.

:::


## Project Users
When a user gains access to Rill Cloud via the project invite, they will only be able to view that specific project. Project Users can have two roles:

- **Viewers** can browse the specific project and view dashboards
- **Admins** can manage the project by making changes to the project's files, gains access to the Status and Settings page on the project and can invite other admins to the project.

For a detailed list of permissions, please refer to the [Roles and Permissions](roles-permissions.md).

### How to add a Project User
There are a few ways to add a project user to Rill Cloud.
1. Administrator invites user to the project using `Share`.
2. User requests access via the project URL.`https://ui.rilldata.com/<project_name>`
3. Adminstrator invites user via the CLI with `--project <project_name>` flag.



### Admin invites user

From the project's splash screen, please select share and type the email[s] along with the type of permissions.

![img](/img/manage/user-management/share-project.png)

Once sent, your invited users will receive this email and will need to accept it to view the project.

![img](/img/manage/user-management/email-invite.png)

### User requests access via URL

Alternatively, if you provide the project URL to your users, they can request access to the group admin. Users can request access via the page below:

![img](/img/manage/user-management/request-access.png)

The admin would receive an email to allow access, and can set the permission after accepting the request via the UI.

![img](/img/manage/user-management/admin-reply.png)

---

### Administrator invites user via the CLI
To add a member to a project, run the following command:
```
rill user add --project [PROJECT NAME]
```
You will then be prompted for details about the user. HINT: Run `rill project list` to show available projects.

If you add a user who has not yet signed up for Rill, they will receive an email inviting them to join.

#### Other actions

Run `rill user --help` to show commands for listing members or changing access.


## User Groups 

Another way to manage user is via User groups. Users can be added to a group using the following from the Rill CLI. You can define the usergroup to have certain permissions on specific groups.

- **Viewers** can view the project's dashboards that the user group has gained permissions
- **Admins** can additionally edit the project, and view and edit project members

For more information on setting up [user group permissions](usergroup-management.md).

### How to add a user to a Usergroup
There are two ways to add a user to a user group.
1. Administrator adds them via Rill Cloud (Coming soon!)
2. Administrator adds them via the CLI

### Administrator adds from Rill Cloud
From the organization page, you can manage users under the *Users* tab. Adding users from this page will add the user to the organizat

import ComingSoon from '@site/src/components/ComingSoon';


![img] Update this image

<ComingSoon />

<div class='contents_to_overlay'>
a
</div>

### Administrator adds from the CLI

```
rill user add --group <group-name>
```
You will then be prompted for details on the user. 

:::note
You will not be asked the permissions when adding to a group as the group defines the user permissions. 

If the user you are trying to add is not part of the organization yet, the CLI will prompt you to add them to the organization first then proceed to adding them to a group.
:::

To see the current members of a group:

```
rill user list --group <group-name>
```

To find the current user group roles, with project flag if looking for specific project's role:

```
rill usergroup list <--project my_project_name>
```

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


## Reference: Walking through access levels


In the following example, you can see the different levels of access to Rill via the organization, project-specific access, user group and user privileges.


<img src = '/img/manage/project-management/project-access.png' class='rounded-gif' />


### Key things to note
1. There are **three** levels of access: organizations, projects, and groups.
2. User groups can _only exist_ within an organization.
    - In the case of adding a user who is not part of the organization to a user group, you will prompted to add them first.
3. User groups permissions can either be added for the organization as a whole, or specific projects.
    - `rill usergroup create [--project project_name]`    
4. All users added to an organization must have at least `viewer` privilege. 
    - In the above diagram, `User 5` is redundant as there's already `viewer` access.