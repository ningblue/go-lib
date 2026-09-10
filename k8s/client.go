// Package k8s 提供 Kubernetes 客户端构建与连通性探测（平台统一能力，各服务经此构建 client）。
// 业务语义（部署/实例/配额）不在此包。
package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// ResolveRESTConfig 从 kubeconfig 文本解析 rest.Config。
func ResolveRESTConfig(kubeconfig string) (*rest.Config, error) {
	cfg, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("parse kubeconfig: %w", err)
	}
	return cfg, nil
}

// NewClientset 由 rest.Config 构建 clientset。
func NewClientset(cfg *rest.Config) (*kubernetes.Clientset, error) {
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("create k8s client: %w", err)
	}
	return cs, nil
}

// NewClientsetFromKubeconfig 由 kubeconfig 文本一步构建 clientset。
func NewClientsetFromKubeconfig(kubeconfig string) (*kubernetes.Clientset, error) {
	cfg, err := ResolveRESTConfig(kubeconfig)
	if err != nil {
		return nil, err
	}
	return NewClientset(cfg)
}

// CheckKubeConfig 连通性探测：解析 kubeconfig 并请求 API server 版本（真实可达性验证）。
func CheckKubeConfig(ctx context.Context, kubeconfig string) error {
	cs, err := NewClientsetFromKubeconfig(kubeconfig)
	if err != nil {
		return err
	}
	if _, err := cs.Discovery().ServerVersion(); err != nil {
		return fmt.Errorf("k8s health check failed: %w", err)
	}
	return nil
}
