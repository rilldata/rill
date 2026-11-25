---
title: "Project Tags"
description: "Label Rill Cloud projects with tags such as prod, stage or sandbox, and filter your organization's projects by tag"
sidebar_label: "Project Tags"
sidebar_position: 23
---

# Project Tags

Tags are short labels on a Rill Cloud project, such as `prod`, `stage` or `sandbox`. Use them to tell production projects apart from projects that are still being reviewed or tested, or to group projects by team or region. Tags appear on your organization's projects page, where anyone can filter the list by tag.

Tags only organize projects. They don't change who can access a project, how it's deployed, or what it costs.

## Add tags in Rill Cloud

You need permission to manage the project to change its tags.

1. Open your project and select **Settings**.
2. In the **Tags** section of the **General** page, type a tag and press Enter. As you type, Rill suggests the tags that other projects in your organization already use.
3. To remove a tag, click the **x** next to it.
4. Click **Save**.

## Filter projects by tag

Each card on your organization's projects page shows the project's tags. When any project has tags, a tag filter appears next to **+ New project**:

1. Click the filter, which shows **All tags** when no tag is selected.
2. Select one or more tags. The page shows the projects that have at least one of the selected tags.
3. To show all projects again, select **Clear filter**.

The selected tags are part of the page URL, for example `https://ui.rilldata.com/my-org?tags=sandbox`, so you can bookmark or share a filtered list.

## Manage tags from the CLI

Set tags when you create a project:

```bash
rill project deploy --project revenue-review --tag sandbox
```

[`rill project connect-github`](/reference/cli/project/connect-github) and [`rill deploy`](/reference/cli/deploy) take the same `--tag` flag. Tags only apply when the command creates the project. To change the tags of an existing project, use [`rill project edit`](/reference/cli/project/edit):

```bash
rill project edit revenue-review --add-tag stage --remove-tag sandbox
```

To list the projects that have any of the given tags, use [`rill project list`](/reference/cli/project/list):

```bash
rill project list --tag prod
rill project list --tag stage,sandbox
```

You can repeat `--tag`, `--add-tag` and `--remove-tag`, or pass a comma-separated list. `rill project list` and `rill project show` print the tags of each project.

## Tag format

Rill normalizes tags so that each tag has only one spelling:

- Tags are case-insensitive and stored in lowercase, so `Prod` becomes `prod`.
- A tag has 1 to 40 characters: letters and digits in any language, hyphens and underscores. It starts with a letter or a digit. In Rill Cloud, spaces and other characters you type become hyphens, so `Data Science` becomes `data-science`.
- A project can have up to 10 tags, shown in alphabetical order.

## Who can see and change tags

| Action | Who |
| :----- | :-- |
| See a project's tags | Anyone who can see the project, including viewers. For a public project, anyone who can open your organization's projects page. |
| Set tags when creating a project | Anyone who can create projects in the organization: organization admins and editors |
| Change a project's tags | Project admins and organization admins (the `manage_project` permission) |

:::caution Don't put sensitive information in tags
Everyone who can see a project can see its tags.
:::

See [Roles and Permissions](/guide/administration/users-and-access/roles-permissions) for details on each role.

## Tags or branches?

A common pattern is to give each person a project tagged `sandbox` to build in, and to review that project before copying the changes into a project tagged `prod`. When the changes are meant for an existing production project, working on a branch of that project is often simpler:

| | Branch of the production project | Separate project tagged `sandbox` or `stage` |
| :-- | :-- | :-- |
| Where you work | A Git branch of the production project, edited with [cloud editing](/developers/deploy/cloud-editing) or in Rill Developer | A separate project, deployed from a local folder |
| How reviewers see it | The branch's preview at `https://ui.rilldata.com/<org>/<project>/@<branch>`, which anyone with access to the project can open, viewers included | The separate project, which reviewers need access to |
| How changes reach production | Merge the branch: **Publish** or **Merge to production** in cloud editing, or a pull request on GitHub | Copy the changes into the production project by hand |
| Credentials and variables | The production project's, with [`dev` overrides](/developers/build/connectors/templating) | Set up again for each project |
| Cost while idle | Branch deployments hibernate after an hour of inactivity | Each project runs its own production deployment |

Use a branch when the work is a change to an existing production project that should be reviewed and merged. Use a separate tagged project for experiments or sandboxes that may never be merged, that need different data or credentials, or that a different group of people should see. Many teams use both: branches for changes, and tags to label each project `prod`, `stage` or `sandbox`.
