# `provisioner/kubernetes`

## Configuration

Example provisioner spec for the Kubernetes provisioner:

```json
{
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

## Development

Be aware that the runtimes provisioned in Kubernetes will need to be able to communicate with the admin server to function correctly, so if the admin server is running locally and you setup provisioning to an external cluster, you'll need to make sure there's an available network path from the runtimes to your local admin server.

## Templates

The Kubernetes resource templates provides a high level of flexibility, but they will need to be adapted to the specific Kubernetes environment. The simplified examples below will provide a good starting point.

Note: For internal Rill users refer to our private infra repos containing environment specific configurations and templates.

### deployment.yaml
```
apiVersion: apps/v1
kind: Deployment
spec:
  replicas: 1
  strategy:
    type: Recreate
  selector:
    matchLabels:
      app.kubernetes.io/name: {{ .Names.Deployment }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ .Names.Deployment }}
    spec:
      securityContext:
        fsGroup: 1000
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: {{ .Names.PVC }}
      containers:
      - args:
        - runtime
        - start
        command:
        - rill
        env:
        - name: RILL_RUNTIME_GRPC_PORT
          value: "8080"
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
        - containerPort: 8080
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
    port: 8080
    targetPort: 8080
  selector:
    app.kubernetes.io/name: {{ .Names.Deployment }}
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
              number: 8080
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

### pvc.yaml
```
apiVersion: v1
kind: PersistentVolumeClaim
spec:
  accessModes:
  - ReadWriteOnce
  resources:
    requests:
      storage: {{ .StorageBytes }}
  storageClassName: storageclass-example
```
