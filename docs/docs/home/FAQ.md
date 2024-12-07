---
title: FAQ
sidebar_label: FAQ
sidebar_position: 20
---

## Getting Started

### What is Rill?
Rill is an operational BI tool that provides fast dashboards that your team will actually use. Rill’s unique architecture combines a last-mile ETL service, an in-memory database, and operational dashboards - all in a single solution

## Install Rill 

### How do I install Rill? 
You can install rill using our installation script:
```bash
curl https://rill.sh | sh
```
### How do I upgrade Rill to the latest version?
If you installed Rill using the installation script described above, you can upgrade by running:
```
rill upgrade
```


### Rill cannot be opened because it is from an unidentified developer.
This occurs when Rill binary is downloaded via the browser. You need to change the permissions to make it executable and remove it from Apple Developer identification quarantine. 
Below CLI commands will help you to do that: 
```bash
chmod a+x rill
xattr -d com.apple.quarantine ./rill
```


### Error - This macOS version is not supported. Please upgrade.
Rill uses duckDB internally which requires a newer [macOS version](https://github.com/duckdb/duckdb/issues/3824). 
Please upgrade your macOS version to 10.14 or higher.


### How do I uninstall Rill?

You can uninstall Rill using the following command:
```bash
rill uninstall
```


## Rill Developer

![dev](/img/concepts/rcvsrd/empty-project.png)

### What is Rill Developer?
Rill Developer is a local application used to preview your project and make any necessary changes before deploying to Rill Cloud. For more information, please review [our documentation](https://docs.rilldata.com/concepts/developerVsCloud#rill-developer).

### I'm having issues with Rill Developer...

Please refer to [our tutorials](/tutorials) to get started using Rill. If you still have any questions, please [contact us!](/contact)


### How do I start more than one instance of Rill Developer?

If you try to start two instances of Rill Developer, you will hit the following error:
```bash
Error: serve: server crashed: grpc port 49009 is in use by another process. Either kill that process or pass `--port-grpc PORT` to run Rill on another port
```

In other to run two instances, please use the following flags with a unique port number.
```bash
rill start --port 10010 --port-grpc 10011
```

### How do I share my dashboard with my colleagues?

You need to [deploy your dashboard to Rill Cloud](https://docs.rilldata.com/deploy/existing-project/) to share your dashboard.

## Rill Cloud

![dev](/img/concepts/rcvsrd/Rill-Cloud.png)


### What is Rill Cloud?
Rill Cloud is where your deployed Rill project exists and can be shared to your colleagues, or end-users. For more information, please review [our documentation](https://docs.rilldata.com/concepts/developerVsCloud#rill-cloud).

### How do I deploy to Rill Cloud?
You can deploy your project directly from the UI by selecting [the Deploy button](https://docs.rilldata.com/deploy/existing-project/#deploying-a-project-via-the-ui).

<img src = '/img/deploy/existing-project/deploy-ui.gif' class='rounded-gif' />
<br />


### How do I make changes to my dashboard in Rill Cloud?

You can follow the same steps as above. The button will have changed from `deploy` to `update`. After selecting this, the objects in your Rill project will be updated. Or, after syncing your Rill project to Github, simply push changes directly to the repository and this will automatically update your project on Rill Cloud.

### How do I share my dashboard to other users?

You will need to [invite users to your organization/project](https://docs.rilldata.com/manage/user-management#option-1---admin-invites-user) or send them a URL for them to [request access to your dashboard](https://docs.rilldata.com/manage/user-management#option-2---user-requests-access). If you just want them to see the contents of your dashboard, you can look into using [public URLs](https://docs.rilldata.com/explore/share-url).
