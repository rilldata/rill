---
title: Deploy to Rill Cloud from Gitlab
description: How to setup continuous deploys to Rill Cloud from Gitlab
sidebar_label: Deploy from Gitlab
sidebar_position: 10
---

While Rill Cloud natively integrates with [Github](https://github.com), you can also deploy your Rill project from [Gitlab](https://about.gitlab.com/) using direct uploads from a [Gitlab CI/CD pipeline](https://docs.gitlab.com/ee/ci/quick_start/).

Follow these steps to setup continuous deployment from Gitlab to Rill Cloud:

1. Create a new Gitlab repository and push your Rill project to it.

2. On your local, [authenticate with Rill Cloud](/manage/user-management#logging-into-rill-cloud) and create an organization (replace `my-org-name` with your desired name):
```bash
rill login
rill org create my-org-name
```

3. Provision a Rill Cloud [service account](/reference/cli/service/create.md) called `gitlab-ci` and copy its access token:
```
rill service create gitlab-ci
```

4. Set the service token as a CI/CD variable called `RILL_SERVICE_TOKEN` in Gitlab (from the repository page, it's under _Settings > CI/CD > Variables_).

5. Create a file named `.gitlab-ci.yml` at the root of the repository containing your Rill project. Paste the following contents into it (replace `my-org-name` and `my-project-name` with your desired names):
```yaml
deploy-rill-cloud:
  stage: deploy
  script:
    - curl -L -o $HOME/rill.zip https://cdn.rilldata.com/rill/latest/rill_linux_amd64.zip 
    - unzip -d $HOME $HOME/rill.zip 
    - $HOME/rill project deploy --org my-org-name --name my-project-name --interactive=false --api-token $RILL_SERVICE_TOKEN
```

Your Rill project should now automatically deploy to `ui.rilldata.com/my-org-name/my-project-name` each time changes are pushed to Gitlab!
