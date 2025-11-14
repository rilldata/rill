package kubernetes

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/c2h5oh/datasize"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
	"github.com/rilldata/rill/admin/database"
	"github.com/rilldata/rill/admin/provisioner"
	"go.uber.org/multierr"
	"go.uber.org/zap"
	k8serrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/util/yaml"
	appsv1ac "k8s.io/client-go/applyconfigurations/apps/v1"
	corev1ac "k8s.io/client-go/applyconfigurations/core/v1"
	netv1ac "k8s.io/client-go/applyconfigurations/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func init() {
	provisioner.Register("kubernetes", NewKubernetes)
}

type KubernetesSpec struct {
	Host           string                   `json:"host"`
	Image          string                   `json:"image"`
	Namespace      string                   `json:"namespace"`
	TimeoutSeconds int                      `json:"timeout_seconds"`
	KubeconfigPath string                   `json:"kubeconfig_path"`
	TemplatePaths  *KubernetesTemplatePaths `json:"template_paths"`
}

type KubernetesTemplatePaths struct {
	HTTPIngress string `json:"http_ingress"`
	GRPCIngress string `json:"grpc_ingress"`
	Service     string `json:"service"`
	Deployment  string `json:"deployment"`
	PVC         string `json:"pvc"`
}

type KubernetesProvisioner struct {
	Spec              *KubernetesSpec
	clientset         *kubernetes.Clientset
	templates         *template.Template
	templatesChecksum string
}

var _ provisioner.Provisioner = (*KubernetesProvisioner)(nil)

type TemplateData struct {
	Image        string
	ImageTag     string
	ProvisionID  string
	Host         string
	CPU          int
	MemoryGB     int
	StorageBytes int64
	Slots        int
	Names        ResourceNames
	Annotations  map[string]string
	Environment  string
}

type ResourceNames struct {
	HTTPIngress string
	GRPCIngress string
	Service     string
	Deployment  string
	PVC         string
}

func NewKubernetes(specJSON []byte, db database.DB, logger *zap.Logger) (provisioner.Provisioner, error) {
	// Parse the Kubernetes provisioner spec
	ksp := &KubernetesSpec{}
	err := json.Unmarshal(specJSON, ksp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubernetes provisioner spec: %w", err)
	}

	// Build config from kubeconfig file, this will fallback to in-cluster config if no kubeconfig is specified
	config, err := clientcmd.BuildConfigFromFlags("", ksp.KubeconfigPath)
	if err != nil {
		return nil, err
	}

	// Create the clientset for the Kubernetes APIs
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	// Add Sprig template functions (removing functions that leak host info)
	// Derived from Helm: https://github.com/helm/helm/blob/main/pkg/engine/funcs.go
	funcMap := sprig.TxtFuncMap()
	delete(funcMap, "env")
	delete(funcMap, "expandenv")

	// Define template files
	templateFiles := []string{
		ksp.TemplatePaths.HTTPIngress,
		ksp.TemplatePaths.GRPCIngress,
		ksp.TemplatePaths.Service,
		ksp.TemplatePaths.Deployment,
		ksp.TemplatePaths.PVC,
	}

	// Parse the template definitions
	templates := template.Must(template.New("").Funcs(funcMap).ParseFiles(templateFiles...))

	// Calculate the combined sha256 sum of all the template files
	h := sha256.New()
	for _, f := range templateFiles {
		d, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		h.Write(d)
	}
	templatesChecksum := hex.EncodeToString(h.Sum(nil))

	return &KubernetesProvisioner{
		Spec:              ksp,
		clientset:         clientset,
		templates:         templates,
		templatesChecksum: templatesChecksum,
	}, nil
}

func (p *KubernetesProvisioner) Type() string {
	return "kubernetes"
}

func (p *KubernetesProvisioner) Supports(rt provisioner.ResourceType) bool {
	return rt == provisioner.ResourceTypeRuntime
}

func (p *KubernetesProvisioner) Close() error {
	return nil
}

func (p *KubernetesProvisioner) Provision(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Parse args
	args, err := provisioner.NewRuntimeArgs(opts.Args)
	if err != nil {
		return nil, err
	}

	// Resolve "latest" version to the current Rill version
	version := args.Version
	if version == "" {
		version = "latest"
	}
	if version == "latest" && opts.RillVersion != "" {
		version = opts.RillVersion
	}

	// Use 'prod' if no environment is specified
	environment := args.Environment
	if environment == "" {
		environment = "prod"
	}

	// Get Kubernetes resource names
	provisionID := getProvisionID(r.ID)
	names := p.getResourceNames(provisionID)

	// Create unique host
	host := p.getHost(provisionID)

	// Define template data
	data := &TemplateData{
		ImageTag:     version,
		Image:        p.Spec.Image,
		ProvisionID:  provisionID,
		Host:         strings.Split(host, "//")[1], // Remove protocol
		CPU:          1 * args.Slots,
		MemoryGB:     4 * args.Slots,
		StorageBytes: 40 * int64(args.Slots) * int64(datasize.GB),
		Slots:        args.Slots,
		Names:        names,
		Annotations:  opts.Annotations,
		Environment:  environment,
	}

	// Define the structured Kubernetes API resources
	httpIng := &netv1ac.IngressApplyConfiguration{}
	grpcIng := &netv1ac.IngressApplyConfiguration{}
	svc := &corev1ac.ServiceApplyConfiguration{}
	pvc := &corev1ac.PersistentVolumeClaimApplyConfiguration{}
	depl := &appsv1ac.DeploymentApplyConfiguration{}

	// Resolve the templates and decode into Kubernetes API resources
	for k, v := range map[string]any{
		p.Spec.TemplatePaths.HTTPIngress: httpIng,
		p.Spec.TemplatePaths.GRPCIngress: grpcIng,
		p.Spec.TemplatePaths.Service:     svc,
		p.Spec.TemplatePaths.PVC:         pvc,
		p.Spec.TemplatePaths.Deployment:  depl,
	} {
		// Resolve template
		buf := &bytes.Buffer{}
		err := p.templates.Lookup(filepath.Base(k)).Execute(buf, data)
		if err != nil {
			return nil, fmt.Errorf("kubernetes provisioner resolve template error: %w", err)
		}

		// Decode into Kubernetes resource
		dec := yaml.NewYAMLOrJSONDecoder(buf, 1000)
		err = dec.Decode(v)
		if err != nil {
			return nil, fmt.Errorf("kubernetes provisioner decode resource error: %w", err)
		}
	}

	applyOptions := metav1.ApplyOptions{FieldManager: "rill-cloud-admin", Force: true}
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "rill-cloud-admin",
		"app.kubernetes.io/instance":   provisionID,
	}
	annotations := map[string]string{
		"checksum/templates": p.templatesChecksum,
	}

	// If the PVC already exists we need to make sure the volume is not decreased, since the Kubernetes storage drivers in general only supports volume expansion
	oldPvc, err := p.clientset.CoreV1().PersistentVolumeClaims(p.Spec.Namespace).Get(ctx, names.PVC, metav1.GetOptions{})
	if err != nil && !k8serrs.IsNotFound(err) {
		return nil, err
	}
	if !k8serrs.IsNotFound(err) {
		if oldPvc.Spec.Resources.Requests.Storage().Cmp(*pvc.Spec.Resources.Requests.Storage()) == 1 {
			pvc.Spec.WithResources(&corev1ac.VolumeResourceRequirementsApplyConfiguration{
				Requests: &oldPvc.Spec.Resources.Requests,
			})
		}
	}

	// Server-Side apply all the Kubernetes resources, for more info on this methodology see https://kubernetes.io/docs/reference/using-api/server-side-apply/
	_, err = p.clientset.CoreV1().PersistentVolumeClaims(p.Spec.Namespace).Apply(ctx, pvc.WithName(names.PVC).WithLabels(labels).WithAnnotations(annotations), applyOptions)
	if err != nil {
		return nil, err
	}

	_, err = p.clientset.AppsV1().Deployments(p.Spec.Namespace).Apply(ctx, depl.WithName(names.Deployment).WithLabels(labels).WithAnnotations(annotations), applyOptions)
	if err != nil {
		return nil, err
	}

	_, err = p.clientset.CoreV1().Services(p.Spec.Namespace).Apply(ctx, svc.WithName(names.Service).WithLabels(labels).WithAnnotations(annotations), applyOptions)
	if err != nil {
		return nil, err
	}

	_, err = p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Apply(ctx, httpIng.WithName(names.HTTPIngress).WithLabels(labels).WithAnnotations(annotations), applyOptions)
	if err != nil {
		return nil, err
	}

	_, err = p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Apply(ctx, grpcIng.WithName(names.GRPCIngress).WithLabels(labels).WithAnnotations(annotations), applyOptions)
	if err != nil {
		return nil, err
	}

	state := &runtimeState{
		Slots:   data.Slots,
		Version: version,
	}

	cfg := &provisioner.RuntimeConfig{
		Host:         host,
		Audience:     host,
		CPU:          data.CPU,
		MemoryGB:     data.MemoryGB,
		StorageBytes: data.StorageBytes,
	}

	return &provisioner.Resource{
		ID:     r.ID,
		Type:   r.Type,
		State:  state.AsMap(),
		Config: cfg.AsMap(),
	}, nil
}

func (p *KubernetesProvisioner) Deprovision(ctx context.Context, r *provisioner.Resource) error {
	// Get Kubernetes resource names
	provisionID := getProvisionID(r.ID)
	names := p.getResourceNames(provisionID)

	// Common delete options
	delPolicy := metav1.DeletePropagationForeground
	delOptions := metav1.DeleteOptions{
		PropagationPolicy: &delPolicy,
	}

	// Delete HTTP ingress
	err1 := p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Delete(ctx, names.HTTPIngress, delOptions)

	// Delete GRPC ingress
	err2 := p.clientset.NetworkingV1().Ingresses(p.Spec.Namespace).Delete(ctx, names.GRPCIngress, delOptions)

	// Delete service
	err3 := p.clientset.CoreV1().Services(p.Spec.Namespace).Delete(ctx, names.Service, delOptions)

	// Delete deployment
	err4 := p.clientset.AppsV1().Deployments(p.Spec.Namespace).Delete(ctx, names.Deployment, delOptions)

	// Delete PVC
	err5 := p.clientset.CoreV1().PersistentVolumeClaims(p.Spec.Namespace).Delete(ctx, names.PVC, delOptions)

	// We ignore not found errors for idempotency
	errs := []error{err1, err2, err3, err4, err5}
	for i := 0; i < len(errs); i++ {
		if k8serrs.IsNotFound(errs[i]) {
			errs[i] = nil
		}
	}

	// This returns 'nil' if all errors are 'nil'
	return multierr.Combine(errs...)
}

func (p *KubernetesProvisioner) AwaitReady(ctx context.Context, r *provisioner.Resource) error {
	// Get Kubernetes resource names
	provisionID := getProvisionID(r.ID)
	names := p.getResourceNames(provisionID)

	// Wait for the deployment to be ready (with timeout)
	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Duration(p.Spec.TimeoutSeconds)*time.Second, true, func(ctx context.Context) (done bool, err error) {
		depl, err := p.clientset.AppsV1().Deployments(p.Spec.Namespace).Get(ctx, names.Deployment, metav1.GetOptions{})
		if err != nil {
			return false, nil
		}
		return depl.Status.AvailableReplicas > 0 && depl.Status.AvailableReplicas == depl.Status.Replicas && depl.Generation == depl.Status.ObservedGeneration, nil
	})
	if err != nil {
		return err
	}

	// As a final step we make sure the runtime can be reached, we retry on failure, to account for potential small delays in network config propagation
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 5
	retryClient.RetryWaitMin = 2 * time.Second
	retryClient.RetryWaitMax = 10 * time.Second
	retryClient.Logger = nil // Disable inbuilt logger
	pingURL, err := url.JoinPath(p.getHost(provisionID), "/v1/ping")
	if err != nil {
		return err
	}
	resp, err := retryClient.Get(pingURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (p *KubernetesProvisioner) Check(ctx context.Context) error {
	// No-op
	return nil
}

func (p *KubernetesProvisioner) CheckResource(ctx context.Context, r *provisioner.Resource, opts *provisioner.ResourceOptions) (*provisioner.Resource, error) {
	// Parse the resource state
	state, err := newRuntimeState(r.State)
	if err != nil {
		return nil, err
	}

	// Get the Kubernetes deployment
	provisionID := getProvisionID(r.ID)
	names := p.getResourceNames(provisionID)
	depl, err := p.clientset.AppsV1().Deployments(p.Spec.Namespace).Get(ctx, names.Deployment, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	// Determine if we should re-provision, and exit early if not
	trigger := false
	trigger = trigger || state.Version != opts.RillVersion                                                         // Version changed
	trigger = trigger || depl.ObjectMeta.Annotations["checksum/templates"] != p.templatesChecksum                  // Templates changed
	trigger = trigger || depl.ObjectMeta.Annotations["organization_plan"] != opts.Annotations["organization_plan"] // Billing plan changed

	if !trigger {
		return r, nil
	}

	// Reprovision the resource
	r, err = p.Provision(ctx, r, opts)
	if err != nil {
		return nil, err
	}
	err = p.AwaitReady(ctx, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (p *KubernetesProvisioner) getResourceNames(provisionID string) ResourceNames {
	return ResourceNames{
		Deployment:  fmt.Sprintf("runtime-%s", provisionID),
		PVC:         fmt.Sprintf("runtime-%s", provisionID),
		Service:     fmt.Sprintf("runtime-%s", provisionID),
		HTTPIngress: fmt.Sprintf("http-runtime-%s", provisionID),
		GRPCIngress: fmt.Sprintf("grpc-runtime-%s", provisionID),
	}
}

func (p *KubernetesProvisioner) getHost(provisionID string) string {
	return strings.ReplaceAll(p.Spec.Host, "*", provisionID)
}

func getProvisionID(resourceID string) string {
	return strings.ReplaceAll(resourceID, "-", "")
}

// runtimeState describes the Kubernetes provisioner's state for a provisioned runtime resource.
type runtimeState struct {
	Slots   int    `mapstructure:"slots"`
	Version string `mapstructure:"version"`
}

func newRuntimeState(state map[string]any) (*runtimeState, error) {
	res := &runtimeState{}
	err := mapstructure.Decode(state, res)
	if err != nil {
		return nil, fmt.Errorf("failed to parse runtime state: %w", err)
	}
	return res, nil
}

func (r *runtimeState) AsMap() map[string]any {
	res := make(map[string]any)
	err := mapstructure.Decode(r, &res)
	if err != nil {
		panic(err)
	}
	return res
}
