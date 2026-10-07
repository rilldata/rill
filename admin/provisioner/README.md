# `provisioner/`

This directory contains provisioners capapble of spinning up resources of a particular type. It has a generic design that supports multiple provisioner implementations and multiple resource types.

There is currently one supported resource type:
- `runtime`: an instance on a Rill runtime (see `runtime/` at the root of our monorepo)

There are currently two supported provisioner implementations:
- `static`: creates runtime instances using a pool of statically configured runtimes
- `kubernetes`: creates runtime instances by dynamically provisioning a dedicated runtime in Kubernetes

## Configuration

The provisioners are configured using the environment variable `RILL_ADMIN_PROVISIONER_SET_JSON` with a named set of provisioners using a format like the following example. More provisioners of the same type can be configured, this is a useful for example to support deployments to different Kubernetes clusters. Furthermore the name of the default provisioner needs to be specified with `RILL_ADMIN_DEFAULT_PROVISIONER`, this provisioner will be used for all deployed projects where a provisioner is not explicitly chosen.

### Choosing a provisioner for runtimes

When provisioning a runtime, the provisioner is resolved in this order:
1. The project's provisioner, if set (superuser-only, via `rill sudo project edit --provisioner`).
2. The org's default provisioner, if set (superuser-only, via `rill sudo org set-default-provisioner`).
3. The global default provisioner (`RILL_ADMIN_DEFAULT_PROVISIONER`).

```json
{
  "static-example":
    {
      "type": "static",
      "spec":
        {
          "runtimes":
            [
              {
                "host": "http://localhost:8081",          // Runtime host
                "slots": 50,                              // Amount of slots in the pre-provisioned runtime
                "data_dir": "/mnt/data",                  // Directory to use for data storage like DB files etc.
                "audience_url": "http://localhost:8081"   // Audience URL (JWT)
              }
            ]
        }
    },

  "kubernetes-example":
    {
      "type": "kubernetes",
      "spec":
        {
          "timeout_seconds": 30,                              // Maximum time to wait for the runtime to become ready
          "data_dir": "/mnt/data",                            // Directory to use for data storage like DB files etc.
          "host": "http://node-*.localhost",                  // The wildcard '*' will be replaced with the deployment's 'provision_id'
          "namespace": "cloud-runtime",                       // Namespace to use in the K8s cluster
          "image": "rilldata/rill",                           // Rill Docker image
          "kubeconfig_path": "kubeconfig.yaml",               // K8s config file to authenticate against the cluster
          "template_paths":
            {
              "http_ingress": "templates/http_ingress.yaml",  // Ingress resource template for HTTP
              "grpc_ingress": "templates/grpc_ingress.yaml",  // Ingress resource template for GRCP
              "service": "templates/service.yaml",            // Service resource template
              "deployment": "templates/deployment.yaml",      // Deployment resource template
              "pvc": "templates/pvc.yaml"                     // PVC resource template
            }
        }
    }
}
```

### Changing the provisioner of existing deployments

Running deployments keep their current provisioner. A change to the project-level or org-level provisioner is applied to a deployment the next time it goes from stopped to running (see `StartDeploymentInner` in `admin/deployments.go`). Use `rill sudo project restart` to stop and start a project's running deployments.

When a deployment starts with a different provisioner than its runtime was provisioned with, the admin service:
1. Deprovisions the runtime resource with the old provisioner (for example deleting its Kubernetes PVC, or releasing its static slots).
2. Resets the resource for the new provisioner, keeping its ID. The runtime instance ID is derived from the resource ID, so it stays the same.
3. Provisions the resource with the new provisioner and updates the deployment's runtime host.

The old runtime's disk is not migrated. The new runtime starts with an empty disk and relies on the runtime's metastore backups in object storage to recover its state. That requires both runtimes to share the same data bucket and metastore ID.
The admin service deliberately doesn't call `DeleteInstance` on the old runtime, since that would also delete the instance's data in object storage.
