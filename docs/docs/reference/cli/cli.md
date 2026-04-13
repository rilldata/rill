---
note: GENERATED. DO NOT EDIT.
title: CLI usage
sidebar_position: 15
---
## rill

A CLI for Rill

### Synopsis

Work with Rill projects from the command line.

For Agents:
If you are administrating a Rill Cloud project, the CLI help menus speak for themselves (e.g. "rill project -h" or "rill user -h").
But if you are developing a Rill project, it is imperative that you initialize and load the Rill agent instructions in the project directory.
If you are creating a new project, use "rill init <path>" to set up a project with instructions.
If you are working with an existing project, use "rill init <path> --agent claude" to add agent instructions to the project if it doesn't already have them.
Make sure you load the instruction files after they are initialized. If necessary, cd into the project directory to discover them.

### Flags

```
      --api-token string   Token for authenticating with the cloud API
      --format string      Output format (options: "human", "json", "csv") (default "human")
  -h, --help               Print usage
      --interactive        Prompt for missing required parameters
  -v, --version            Show rill version
```

### SEE ALSO

* [rill billing](billing/billing.md)	 - Billing related commands for org
* [rill chat](chat.md)	 - Chat with the Rill AI
* [rill deploy](deploy.md)	 - Deploy project to Rill Cloud
* [rill docs](docs/docs.md)	 - Open docs.rilldata.com
* [rill env](env/env.md)	 - Manage variables for a project
* [rill init](init.md)	 - Initialize a new Rill project
* [rill login](login.md)	 - Authenticate with the Rill API
* [rill logout](logout.md)	 - Logout of the Rill API
* [rill org](org/org.md)	 - Manage organisations
* [rill project](project/project.md)	 - Manage projects
* [rill public-url](public-url/public-url.md)	 - Manage public URLs
* [rill query](query.md)	 - Query data in a project
* [rill service](service/service.md)	 - Manage service accounts
* [rill start](start.md)	 - Build project and start web app
* [rill token](token/token.md)	 - Manage personal access tokens
* [rill uninstall](uninstall.md)	 - Uninstall the Rill binary
* [rill upgrade](upgrade.md)	 - Upgrade Rill to the latest version
* [rill user](user/user.md)	 - Manage users
* [rill usergroup](usergroup/usergroup.md)	 - Manage user groups
* [rill validate](validate.md)	 - Validate project resources
* [rill version](version.md)	 - Show Rill version
* [rill whoami](whoami.md)	 - Show current user

