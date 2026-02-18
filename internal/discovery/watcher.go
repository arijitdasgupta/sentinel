package discovery

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/arijitdasgupta/sentinel/internal/config"
)

type IngressWatcher struct {
	mu      sync.RWMutex
	targets []config.Target
	onChange func([]config.Target)
}

func New(onChange func([]config.Target)) (*IngressWatcher, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes client: %w", err)
	}

	w := &IngressWatcher{
		onChange: onChange,
	}

	factory := informers.NewSharedInformerFactory(clientset, 0)
	ingressInformer := factory.Networking().V1().Ingresses().Informer()

	ingressInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    func(obj interface{}) { w.rebuild(ingressInformer.GetStore()) },
		UpdateFunc: func(_, obj interface{}) { w.rebuild(ingressInformer.GetStore()) },
		DeleteFunc: func(obj interface{}) { w.rebuild(ingressInformer.GetStore()) },
	})

	w.targets = []config.Target{}

	go factory.Start(context.Background().Done())
	factory.WaitForCacheSync(context.Background().Done())

	slog.Info("ingress watcher started")

	return w, nil
}

func (w *IngressWatcher) Targets() []config.Target {
	w.mu.RLock()
	defer w.mu.RUnlock()
	result := make([]config.Target, len(w.targets))
	copy(result, w.targets)
	return result
}

func (w *IngressWatcher) rebuild(store cache.Store) {
	seen := make(map[string]bool)
	var targets []config.Target

	for _, obj := range store.List() {
		ingress, ok := obj.(*networkingv1.Ingress)
		if !ok {
			continue
		}

		for _, rule := range ingress.Spec.Rules {
			host := rule.Host
			if host == "" || seen[host] {
				continue
			}
			seen[host] = true

			url := fmt.Sprintf("https://%s", host)
			targets = append(targets, config.Target{
				URL:  url,
				Host: host,
			})

			slog.Info("discovered ingress host", "host", host, "namespace", ingress.Namespace, "ingress", ingress.Name)
		}
	}

	w.mu.Lock()
	w.targets = targets
	w.mu.Unlock()

	slog.Info("ingress targets updated", "count", len(targets))

	if w.onChange != nil {
		w.onChange(targets)
	}
}
