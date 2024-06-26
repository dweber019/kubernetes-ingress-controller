package fallback

import (
	"fmt"

	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kong/kubernetes-ingress-controller/v3/internal/store"
	kongv1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1"
	kongv1beta1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/configuration/v1beta1"
	incubatorv1alpha1 "github.com/kong/kubernetes-ingress-controller/v3/pkg/apis/incubator/v1alpha1"
)

// resolveKongPluginDependencies resolves potential dependencies for a KongPlugin object:
// - Secret.
func resolveKongPluginDependencies(cache store.CacheStores, kongPlugin *kongv1.KongPlugin) []client.Object {
	var dependencies []client.Object
	if cf := kongPlugin.ConfigFrom; cf != nil {
		if s, ok := fetchSecret(
			cache,
			k8stypes.NamespacedName{
				Namespace: kongPlugin.Namespace,
				Name:      cf.SecretValue.Secret,
			},
		); ok {
			dependencies = append(dependencies, s)
		}
	}
	for _, cp := range kongPlugin.ConfigPatches {
		if s, ok := fetchSecret(
			cache,
			k8stypes.NamespacedName{
				Namespace: kongPlugin.Namespace,
				Name:      cp.ValueFrom.SecretValue.Secret,
			},
		); ok {
			dependencies = append(dependencies, s)
		}
	}
	return dependencies
}

// resolveKongClusterPluginDependencies resolves potential dependencies for a KongClusterPlugin object:
// - Secret.
func resolveKongClusterPluginDependencies(cache store.CacheStores, kongClusterPlugin *kongv1.KongClusterPlugin) []client.Object {
	var dependencies []client.Object
	if cf := kongClusterPlugin.ConfigFrom; cf != nil {
		if s, ok := fetchSecret(
			cache,
			k8stypes.NamespacedName{
				Namespace: cf.SecretValue.Namespace,
				Name:      cf.SecretValue.Secret,
			},
		); ok {
			dependencies = append(dependencies, s)
		}
	}
	for _, cp := range kongClusterPlugin.ConfigPatches {
		if s, ok := fetchSecret(
			cache,
			k8stypes.NamespacedName{
				Namespace: cp.ValueFrom.SecretValue.Namespace,
				Name:      cp.ValueFrom.SecretValue.Secret,
			},
		); ok {
			dependencies = append(dependencies, s)
		}
	}
	return dependencies
}

// resolveKongConsumerDependencies resolves potential dependencies for a KongConsumer object:
// - KongPlugin
// - KongClusterPlugin.
func resolveKongConsumerDependencies(cache store.CacheStores, kongConsumer *kongv1.KongConsumer) []client.Object {
	return resolveObjectDependenciesPlugin(cache, kongConsumer)
}

// resolveKongConsumerGroupDependencies resolves potential dependencies for a KongConsumerGroup object:
// - KongPlugin
// - KongClusterPlugin.
func resolveKongConsumerGroupDependencies(cache store.CacheStores, kongConsumerGroup *kongv1beta1.KongConsumerGroup) []client.Object {
	return resolveObjectDependenciesPlugin(cache, kongConsumerGroup)
}

// resolveUDPIngressDependencies resolves potential dependencies for a UDPIngress object:
// - KongPlugin
// - KongClusterPlugin
// - Service.
func resolveUDPIngressDependencies(cache store.CacheStores, udpIngress *kongv1beta1.UDPIngress) []client.Object {
	dependencies := resolveObjectDependenciesPlugin(cache, udpIngress)
	for _, rule := range udpIngress.Spec.Rules {
		if service, exists, err := cache.Service.GetByKey(
			fmt.Sprintf("%s/%s", udpIngress.GetNamespace(), rule.Backend.ServiceName),
		); err == nil && exists {
			dependencies = append(dependencies, service.(client.Object))
		}
	}
	return dependencies
}

// resolveTCPIngressDependencies resolves potential dependencies for a TCPIngress object:
// - KongPlugin
// - KongClusterPlugin
// - Service.
func resolveTCPIngressDependencies(cache store.CacheStores, tcpIngress *kongv1beta1.TCPIngress) []client.Object {
	dependencies := resolveObjectDependenciesPlugin(cache, tcpIngress)
	for _, rule := range tcpIngress.Spec.Rules {
		if service, exists, err := cache.Service.GetByKey(
			fmt.Sprintf("%s/%s", tcpIngress.GetNamespace(), rule.Backend.ServiceName),
		); err == nil && exists {
			dependencies = append(dependencies, service.(client.Object))
		}
	}
	return dependencies
}

// resolveKongServiceFacadeDependencies resolves potential dependencies for a KongServiceFacade object:
// - KongPlugin
// - KongClusterPlugin
// - KongUpstreamPolicy.
func resolveKongServiceFacadeDependencies(cache store.CacheStores, kongServiceFacade *incubatorv1alpha1.KongServiceFacade) []client.Object {
	return resolveDependenciesForServiceLikeObj(cache, kongServiceFacade)
}
