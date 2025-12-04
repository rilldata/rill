---
title: Deploy to Rill Cloud from Gitlab
description: How to setup continuous deploys to Rill Cloud from Gitlab
sidebar_label: Deploy from Gitlab
sidebar_position: 10
---

While Rill Cloud natively integrates with [GitHub](https://github.com), you can also deploy your Rill project from [Gitlab](https://about.gitlab.com/) using direct uploads from a [Gitlab CI/CD pipeline](https://docs.gitlab.com/ee/ci/quick_start/).

Follow these steps to set up continuous deployment from Gitlab to Rill Cloud:

1. Create a new Gitlab repository and push your Rill project to it.

2. On your local, [authenticate with Rill Cloud](/manage/user-management#logging-into-rill-cloud) and create an organization (replace `my-org-name` with your desired name):
```bash
rill login
rill org create my-org-name
```

3. Create the project in Rill Cloud
```bash
rill project deploy
```

:::note Multiple branches
If your repo contains multiple branches ensure the branch you want to deploy from via
```bash
rill project edit --project my-project-name --prod-branch my-branch-name
```
:::

4. Provision a Rill Cloud [service account](/reference/cli/service/create) called `gitlab-ci` and copy its access token:
```
rill service create gitlab-ci
```

5. Set the service token as a CI/CD variable called `RILL_SERVICE_TOKEN` in Gitlab (from the repository page, it's under _Settings > CI/CD > Variables_).

6. Create a file named `.gitlab-ci.yml` at the root of the repository containing your Rill project. Paste the following contents into it (replace `my-org-name` and `my-project-name` with your desired names):
```yaml
deploy-rill-cloud:
  stage: deploy
  script: 
    - curl -L -o $HOME/rill.zip https://cdn.rilldata.com/rill/latest/rill_linux_amd64.zip 
    - unzip -d $HOME $HOME/rill.zip 
    - git checkout -B "$CI_COMMIT_REF_NAME" "$CI_COMMIT_SHA"
    - $HOME/rill project deploy --org my-org-name --project my-project-name --interactive=false --api-token $RILL_SERVICE_TOKEN
```

Your Rill project should now automatically deploy to `ui.rilldata.com/my-org-name/my-project-name` each time changes are pushed to Gitlab!

:::note File size limits
We enforce a file size limit of 100mb so ensure you do not unpack the rill binary in the repo root or add it to your .gitignore
:::