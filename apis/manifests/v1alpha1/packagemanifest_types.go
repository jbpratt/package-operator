package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"

	corev1alpha1 "package-operator.run/apis/core/v1alpha1"
)

const (
	// PackagePhaseAnnotation annotation to assign objects to a phase.
	PackagePhaseAnnotation = "package-operator.run/phase"
	// PackageConditionMapAnnotation specifies object conditions to map back into Package Operator APIs.
	// Example: Available => my-own-prefix/Available.
	PackageConditionMapAnnotation = "package-operator.run/condition-map"
	// PackageExternalObjectAnnotation when set to "True", indicates
	// that the referenced object should only be observed during a phase
	// rather than reconciled.
	PackageExternalObjectAnnotation = "package-operator.run/external"
)

const (
	// PackageLabel contains the name of the Package from the PackageManifest.
	PackageLabel = "package-operator.run/package"
	// PackageSourceImageAnnotation references the package container image originating this object.
	PackageSourceImageAnnotation = "package-operator.run/package-source-image"
	// PackageConfigAnnotation contains the configuration for this object.
	PackageConfigAnnotation = "package-operator.run/package-config"
	// PackageInstanceLabel contains the name of the Package instance.
	PackageInstanceLabel = "package-operator.run/instance"
)

// PackageManifest defines the manifest of a package.
// +kubebuilder:object:root=true
type PackageManifest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec PackageManifestSpec `json:"spec,omitempty"`
	Test PackageManifestTest `json:"test,omitempty"`
}

// PackageManifestScope declares the available scopes to install this package in.
type PackageManifestScope string

const (
	// PackageManifestScopeCluster scope allows the package to be installed for the whole cluster.
	// The package needs to default installation namespaces and create them.
	PackageManifestScopeCluster PackageManifestScope = "Cluster"
	// PackageManifestScopeNamespaced scope allows the package to be installed for specific namespaces.
	PackageManifestScopeNamespaced PackageManifestScope = "Namespaced"
)

// PackageManifestSpec represents the spec of the packagemanifest containing the
// details about phases and availability probes.
type PackageManifestSpec struct {
	// Scopes declare the available installation scopes for the package.
	// Either Cluster, Namespaced, or both.
	Scopes []PackageManifestScope `json:"scopes"`
	// Phases correspond to the references to the phases which are going to be the
	// part of the ObjectDeployment/ClusterObjectDeployment.
	Phases []PackageManifestPhase `json:"phases"`
	// Availability Probes check objects that are part of the package.
	// All probes need to succeed for a package to be considered Available.
	// Failing probes will prevent the reconciliation of objects in later phases.
	// +optional
	AvailabilityProbes []corev1alpha1.ObjectSetProbe `json:"availabilityProbes,omitempty"`
	// Configuration specification.
	Config PackageManifestSpecConfig `json:"config,omitempty"`
	// List of images to be resolved
	Images []PackageManifestImage `json:"images"`
	// Configuration for multi-component packages. If this field is not set it is assumed
	// that the containing package is a single-component package.
	// +optional
	Components *PackageManifestComponentsConfig `json:"components,omitempty"`
	// Constraints limit what environments a package can be installed into.
	// e.g. can only be installed on OpenShift.
	// +optional
	Constraints []PackageManifestConstraint `json:"constraints,omitempty"`
	// Repository references that are used to validate constraints and resolve dependencies.
	Repositories []PackageManifestRepository `json:"repositories,omitempty"`
	// Dependency references to resolve and use within this package.
	Dependencies []PackageManifestDependency `json:"dependencies,omitempty"`
}

type PackageManifestRepository struct {
	// References a file in the filesystem to load.
	// +example=../myrepo.yaml
	File string `json:"file,omitempty"`
	// References an image in a container image registry.
	// +example=quay.io/package-operator/my-repo:latest
	Image string `json:"image,omitempty"`
}

// Uses a solver to find the latest version package image.
type PackageManifestDependency struct {
	// Resolves the dependency as a image url and digest and commits it to the PackageManifestLock.
	Image *PackageManifestDependencyImage `json:"image,omitempty"`
}

type PackageManifestDependencyImage struct {
	// Name for the dependency.
	// +example=my-pkg
	Name string `json:"name"`
	// Package FQDN <package-name>.<repository name>
	// +example=my-pkg.my-repo
	Package string `json:"package"`
	// Semantic Versioning 2.0.0 version range.
	// +example=>=2.1
	Range string `json:"range"`
}

// PackageManifestConstraint configures environment constraints to block package installation.
type PackageManifestConstraint struct {
	// PackageManifestPlatformVersionConstraint enforces that the platform matches the given version range.
	// This constraint is ignored when running on a different platform.
	// e.g. a PlatformVersionConstraint OpenShift>=4.13.x is ignored when installed on a plain Kubernetes cluster.
	// Use the Platform constraint to enforce running on a specific platform.
	PlatformVersion *PackageManifestPlatformVersionConstraint `json:"platformVersion,omitempty"`
	// Valid platforms that support this package.
	// +example=[Kubernetes]
	Platform []PlatformName `json:"platform,omitempty"`
	// Constraints this package to be only installed once in the Cluster or once in the same Namespace.
	UniqueInScope *PackageManifestUniqueInScopeConstraint `json:"uniqueInScope,omitempty"`
}

// PlatformName holds the name of a specific platform flavor name.
// e.g. Kubernetes, OpenShift.
type PlatformName string

const (
	// Plain Kubernetes.
	Kubernetes PlatformName = "Kubernetes"
	// Red Hat OpenShift.
	OpenShift PlatformName = "OpenShift"
)

// PackageManifestPlatformVersionConstraint enforces that the platform matches the given version range.
// This constraint is ignored when running on a different platform.
// e.g. a PlatformVersionConstraint OpenShift>=4.13.x is ignored when installed on a plain Kubernetes cluster.
// Use the Platform constraint to enforce running on a specific platform.
type PackageManifestPlatformVersionConstraint struct {
	// Name of the platform this constraint should apply to.
	// +example=Kubernetes
	Name PlatformName `json:"name"`
	// Semantic Versioning 2.0.0 version range.
	// +example=>=1.20.x
	Range string `json:"range"`
}

type PackageManifestUniqueInScopeConstraint struct{}

// PackageManifestComponentsConfig configures components of a package.
type PackageManifestComponentsConfig struct{}

// PackageManifestSpecConfig configutes a package manifest.
type PackageManifestSpecConfig struct {
	// OpenAPIV3Schema is the OpenAPI v3 schema to use for validation and pruning.
	OpenAPIV3Schema *apiextensionsv1.JSONSchemaProps `json:"openAPIV3Schema,omitempty"`
}

// PackageManifestPhase defines a package phase.
type PackageManifestPhase struct {
	// Name of the reconcile phase. Must be unique within a PackageManifest
	Name string `json:"name"`
	// If non empty, phase reconciliation is delegated to another controller.
	// If set to the string "default" the built-in controller reconciling the object.
	// If set to any other string, an out-of-tree controller needs to be present to handle ObjectSetPhase objects.
	Class string `json:"class,omitempty"`
}

// PackageManifestImage specifies an image tag to be resolved.
type PackageManifestImage struct {
	// Image name to be use to reference it in the templates
	Name string `json:"name"`
	// Image identifier (REPOSITORY[:TAG])
	Image string `json:"image"`
}

// PackageManifestTest configures test cases.
type PackageManifestTest struct {
	// Template testing configuration.
	Template    []PackageManifestTestCaseTemplate `json:"template,omitempty"`
	Kubeconform *PackageManifestTestKubeconform   `json:"kubeconform,omitempty"`
}

// PackageManifestTestCaseTemplate template testing configuration.
type PackageManifestTestCaseTemplate struct {
	// Name describing the test case.
	Name string `json:"name"`
	// Template data to use in the test case.
	Context TemplateContext `json:"context,omitempty"`
}

// PackageManifestTestKubeconform configures kubeconform testing.
type PackageManifestTestKubeconform struct {
	// Kubernetes version to use schemas from.
	KubernetesVersion string `json:"kubernetesVersion"`
	//nolint:lll
	// OpenAPI schema locations for kubeconform
	// defaults to:
	// - https://raw.githubusercontent.com/yannh/kubernetes-json-schema/master/{{ .NormalizedKubernetesVersion }}-standalone{{ .StrictSuffix }}/{{ .ResourceKind }}{{ .KindSuffix }}.json
	// - https://raw.githubusercontent.com/datreeio/CRDs-catalog/main/{{.Group}}/{{.ResourceKind}}_{{.ResourceAPIVersion}}.json
	SchemaLocations []string `json:"schemaLocations,omitempty"`
}

// TemplateContext is available within the package templating process.
type TemplateContext struct {
	// Package object.
	Package TemplateContextPackage `json:"package"`
	// Configuration as presented via the (Cluster)Package API after admission.
	Config *runtime.RawExtension `json:"config,omitempty"`
	// Environment specific information.
	Environment PackageEnvironment `json:"environment"`
}

// PackageEnvironment information.
type PackageEnvironment struct {
	// Kubernetes environment information. This section is always set.
	Kubernetes PackageEnvironmentKubernetes `json:"kubernetes"`
	// OpenShift environment information. This section is only set when OpenShift is detected.
	OpenShift *PackageEnvironmentOpenShift `json:"openShift,omitempty"`
	// Proxy configuration. Only available on OpenShift when the cluster-wide Proxy is enabled.
	// https://docs.openshift.com/container-platform/latest/networking/enable-cluster-wide-proxy.html
	Proxy *PackageEnvironmentProxy `json:"proxy,omitempty"`
	// HyperShift specific information. Only available when installed alongside HyperShift.
	// https://github.com/openshift/hypershift
	HyperShift *PackageEnvironmentHyperShift `json:"hyperShift,omitempty"`
}

// PackageEnvironmentKubernetes configures kubernetes environments.
type PackageEnvironmentKubernetes struct {
	// Kubernetes server version.
	Version string `json:"version"`
}

// PackageEnvironmentOpenShift configures openshift environments.
type PackageEnvironmentOpenShift struct {
	// OpenShift server version.
	Version string `json:"version"`
}

// PackageEnvironmentProxy configures proxy environments.
// On OpenShift, this config is taken from the cluster Proxy object.
// https://docs.openshift.com/container-platform/4.13/networking/enable-cluster-wide-proxy.html
type PackageEnvironmentProxy struct {
	// HTTP_PROXY
	HTTPProxy string `json:"httpProxy,omitempty"`
	// HTTPS_PROXY
	HTTPSProxy string `json:"httpsProxy,omitempty"`
	// NO_PROXY
	NoProxy string `json:"noProxy,omitempty"`
}

// PackageEnvironmentHyperShift contains HyperShift specific information.
// Only available when installed alongside HyperShift.
// https://github.com/openshift/hypershift
type PackageEnvironmentHyperShift struct {
	// Contains HyperShift HostedCluster specific information.
	// This information is only available when installed alongside HyperShift within a HostedCluster Namespace.
	// https://github.com/openshift/hypershift
	HostedCluster *PackageEnvironmentHyperShiftHostedCluster `json:"hostedCluster"`
}

// PackageEnvironmentHyperShiftHostedCluster contains HyperShift HostedCluster specific information.
// This information is only available when installed alongside HyperShift within a HostedCluster Namespace.
// https://github.com/openshift/hypershift
type PackageEnvironmentHyperShiftHostedCluster struct {
	TemplateContextObjectMeta `json:"metadata"`
	// Namespace of HostedCluster components belonging to this HostedCluster object.
	HostedClusterNamespace string `json:"hostedClusterNamespace"`
}

// TemplateContextPackage represents the (Cluster)Package object requesting this package content.
type TemplateContextPackage struct {
	TemplateContextObjectMeta `json:"metadata"`
}

// TemplateContextObjectMeta represents a simplified version of metav1.ObjectMeta for use in templates.
type TemplateContextObjectMeta struct {
	Name        string            `json:"name"`
	Namespace   string            `json:"namespace"`
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
}

func init() { register(&PackageManifest{}) }
