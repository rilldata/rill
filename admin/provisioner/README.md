# provisioner
This directory contains the provisioner package. On a high level the provisioner's main responsibility is to allocate resources on a runtime to a deployment. Currently there is two different types of provisioners, `static` and `kubernetes`. The `static` type will allocate a runtime from a statically pre-defined pool of runtimes and the `kubernetes` type will dynamically provision a dedicated runtime in Kubernetes and allocate it to the deployment.

## Configuration
The provisioner is configured using `RILL_ADMIN_PROVISIONER_SET_JSON` with a named set of provisioners using a format like the following example. More provisioners of the same type can be configured, this is a useful for example to support deployments to different Kubernetes clusters. Furthermore the name of the default provisioner needs to be specified with `RILL_ADMIN_DEFAULT_PROVISIONER`, this provisioner will be used for all deployed projects where a provisioner is not explicitly chosen.
```
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
              "statefulset": "templates/statefulset.yaml"     // Statefulset resource template
            }
        }
    }
}
```

## Development

Be aware that the runtimes provisioned in Kubernetes will need to be able to communicate with the admin server to function correctly, so if the admin server is running locally and you setup provisioning to an external cluster, you'll need to make sure there's an available network path from the runtimes to your local admin server.

## Templates

The Kubernetes resource templates provides a high level of flexibility, but they will need to be adapted to the specific Kubernetes environment. The simplified examples below will provide a good starting point.

Note: For internal Rill users refer to our private infra repos containing environment specific configurations and templates.

### statefulset.yaml
```
apiVersion: apps/v1
kind: StatefulSet
spec:
  persistentVolumeClaimRetentionPolicy:
    whenDeleted: Delete
    whenScaled: Retain
  volumeClaimTemplates:
  - metadata:
      name: data
    spec:
      accessModes:
        - ReadWriteOnce
      resources:
        requests:
          storage: {{ .StorageBytes }}
  replicas: 1
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ .Names.StatefulSet }}
  serviceName: cloud-runtime
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Names.StatefulSet }}
    spec:
      securityContext:
        fsGroup: 1000
      containers:
      - args:
        - runtime
        - start
        command:
        - rill
        env:
        - name: RILL_RUNTIME_GRPC_PORT
          value: "9090"
        - name: RILL_RUNTIME_HTTP_PORT
          value: "8080"

        ########################################################################
        # Add all the relevant runtime configuration environment variables here
        ########################################################################

        image: "{{ .Image }}:{{ .ImageTag }}"
        imagePullPolicy: Always
        livenessProbe:
          failureThreshold: 3
          httpGet:
            path: /v1/ping
            port: 8080
            scheme: HTTP
          periodSeconds: 10
          successThreshold: 1
          timeoutSeconds: 1
        name: cloud-runtime
        ports:
        - containerPort: 8080
          protocol: TCP
        - containerPort: 9090
          protocol: TCP
        resources:
          limits:
            cpu: {{ .CPU }}
            memory: {{ .MemoryGB }}Gi
          requests:
            cpu: {{ .CPU }}
            memory: {{ .MemoryGB }}Gi
        securityContext:
          allowPrivilegeEscalation: false
          capabilities:
            add:
            - NET_BIND_SERVICE
            drop:
            - all
          runAsNonRoot: true
          runAsUser: 1000
        terminationMessagePolicy: FallbackToLogsOnError
        volumeMounts:
        - mountPath: {{ .DataDir }}
          name: data
```

### service.yaml
```
apiVersion: v1
kind: Service
spec:
  type: ClusterIP
  ports:
  - name: http
    port: 8080
    targetPort: 8080
  - name: grpc
    port: 9090
    targetPort: 9090
  selector:
    app.kubernetes.io/name: {{ .Names.StatefulSet }}
```

### grpc_ingress.yaml
```
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  annotations:
    nginx.ingress.kubernetes.io/backend-protocol: GRPC
spec:
  ingressClassName: nginx
  rules:
  - host: {{ .Host }}
    http:
      paths:
      - backend:
          service:
            name: {{ .Names.Service }}
            port:
              number: 9090
        path: /
        pathType: Prefix
```


### http_ingress.yaml
```
apiVersion: networking.k8s.io/v1
kind: Ingress
spec:
  ingressClassName: nginx
  rules:
  - host: {{ .Host }}
    http:
      paths:
      - backend:
          service:
            name: {{ .Names.Service }}
            port:
              number: 8080
        path: /v1
        pathType: Prefix
```
