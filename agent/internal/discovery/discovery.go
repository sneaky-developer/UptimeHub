package discovery

import (
	"context"
	"fmt"
	"log"
	"sync"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// LabelMonitoringEnabled is the label to enable monitoring for a service
	LabelMonitoringEnabled = "monitoring.enabled"
	// AnnotationMonitoringPath is the annotation for the health check path
	AnnotationMonitoringPath = "monitoring.path"
	// AnnotationMonitoringPort is the annotation for the health check port
	AnnotationMonitoringPort = "monitoring.port"
)

// DiscoveredService represents a K8s service discovered for monitoring
type DiscoveredService struct {
	Key       string // stable identity across syncs, e.g. "svc/default/my-api"
	Name      string
	Namespace string
	URL       string
	Path      string
	Port      string
}

// Discovery handles Kubernetes service discovery
type Discovery struct {
	client    kubernetes.Interface
	namespace string
	mu        sync.RWMutex
	services  []DiscoveredService
}

// NewDiscovery creates a new Discovery instance
func NewDiscovery(inCluster bool, namespace string) (*Discovery, error) {
	var cfg *rest.Config
	var err error

	if inCluster {
		cfg, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
		}
	} else {
		// Use kubeconfig from default location (~/.kube/config)
		loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
		configOverrides := &clientcmd.ConfigOverrides{}
		kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

		cfg, err = kubeConfig.ClientConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to get kubeconfig: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return &Discovery{
		client:    clientset,
		namespace: namespace,
	}, nil
}

// Discover finds all services with the monitoring label. The returned bool
// reports whether the discovery covered all sources (Services and Ingresses);
// callers should only treat the list as authoritative when it is true.
func (d *Discovery) Discover(ctx context.Context) ([]DiscoveredService, bool, error) {
	labelSelector := LabelMonitoringEnabled + "=true"

	namespace := d.namespace
	if namespace == "" {
		namespace = metav1.NamespaceAll
	}

	// Discover from Services
	svcList, err := d.client.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		return nil, false, fmt.Errorf("failed to list services: %w", err)
	}

	complete := true

	var discovered []DiscoveredService

	for _, svc := range svcList.Items {
		path := "/health"
		if p, ok := svc.Annotations[AnnotationMonitoringPath]; ok {
			path = p
		}

		port := "80"
		if p, ok := svc.Annotations[AnnotationMonitoringPort]; ok {
			port = p
		} else if len(svc.Spec.Ports) > 0 {
			port = fmt.Sprintf("%d", svc.Spec.Ports[0].Port)
		}

		// Construct the internal cluster URL
		url := fmt.Sprintf("http://%s.%s.svc.cluster.local:%s%s",
			svc.Name, svc.Namespace, port, path)

		discovered = append(discovered, DiscoveredService{
			Key:       fmt.Sprintf("svc/%s/%s", svc.Namespace, svc.Name),
			Name:      svc.Name,
			Namespace: svc.Namespace,
			URL:       url,
			Path:      path,
			Port:      port,
		})
	}

	// Discover from Ingresses
	ingList, err := d.client.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		log.Printf("Warning: failed to list ingresses (may need RBAC): %v", err)
		complete = false
	} else {
		for _, ing := range ingList.Items {
			path := "/health"
			if p, ok := ing.Annotations[AnnotationMonitoringPath]; ok {
				path = p
			}

			for _, rule := range ing.Spec.Rules {
				if rule.Host != "" {
					scheme := "https"
					if len(ing.Spec.TLS) == 0 {
						scheme = "http"
					}

					url := fmt.Sprintf("%s://%s%s", scheme, rule.Host, path)
					discovered = append(discovered, DiscoveredService{
						Key:       fmt.Sprintf("ing/%s/%s/%s", ing.Namespace, ing.Name, rule.Host),
						Name:      ing.Name,
						Namespace: ing.Namespace,
						URL:       url,
						Path:      path,
					})
				}
			}
		}
	}

	d.mu.Lock()
	d.services = discovered
	d.mu.Unlock()

	log.Printf("🔍 Discovered %d services (%d from K8s Services, %d from Ingresses)",
		len(discovered), len(svcList.Items), len(discovered)-len(svcList.Items))

	return discovered, complete, nil
}

// GetServices returns the last discovered services
func (d *Discovery) GetServices() []DiscoveredService {
	d.mu.RLock()
	defer d.mu.RUnlock()
	result := make([]DiscoveredService, len(d.services))
	copy(result, d.services)
	return result
}
