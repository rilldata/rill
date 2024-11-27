# `provisioner/`

This directory contains provisioners capapble of spinning up resources of a particular type. It has a generic design that supports multiple provisioner implementations and multiple resource types.

There is currently one supported resource type:
- `runtime`: an instance on a Rill runtime (see `runtime/` at the root of our monorepo)

There are currently two supported provisioner implementations:
- `static`: creates runtime instances using a pool of statically configured runtimes
- `kubernetes`: creates runtime instances by dynamically provisioning a dedicated runtime in Kubernetes

## Configuration

The provisioners are configured using the environment variable `RILL_ADMIN_PROVISIONER_SET_JSON` with a named set of provisioners using a format like the following example. More provisioners of the same type can be configured, this is a useful for example to support deployments to different Kubernetes clusters. Furthermore the name of the default provisioner needs to be specified with `RILL_ADMIN_DEFAULT_PROVISIONER`, this provisioner will be used for all deployed projects where a provisioner is not explicitly chosen.
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
                "host": "http://localhost:9091",          // Runtime host
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
