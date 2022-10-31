package tolerantinformer

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	rolloutinformers "github.com/argoproj/argo-rollouts/pkg/client/informers/externalversions/rollouts/v1alpha1"
	rolloutlisters "github.com/argoproj/argo-rollouts/pkg/client/listers/rollouts/v1alpha1"
)

func NewTolerantISDTemplateInformer(factory dynamicinformer.DynamicSharedInformerFactory) rolloutinformers.ISDTemplateInformer {
	return &tolerantISDTemplateInformer{
		delegate: factory.ForResource(v1alpha1.ISDTemplateGVR),
	}
}

type tolerantISDTemplateInformer struct {
	delegate informers.GenericInformer
}

func (i *tolerantISDTemplateInformer) Informer() cache.SharedIndexInformer {
	return i.delegate.Informer()
}

func (i *tolerantISDTemplateInformer) Lister() rolloutlisters.ISDTemplateLister {
	return &tolerantISDTemplateLister{
		delegate: i.delegate.Lister(),
	}
}

type tolerantISDTemplateLister struct {
	delegate cache.GenericLister
}

func (t *tolerantISDTemplateLister) List(selector labels.Selector) ([]*v1alpha1.ISDTemplate, error) {
	objects, err := t.delegate.List(selector)
	if err != nil {
		return nil, err
	}
	return convertObjectsToISDTemplates(objects)
}

func (t *tolerantISDTemplateLister) ISDTemplates(namespace string) rolloutlisters.ISDTemplateNamespaceLister {
	return &tolerantISDTemplateNamespaceLister{
		delegate: t.delegate.ByNamespace(namespace),
	}
}

type tolerantISDTemplateNamespaceLister struct {
	delegate cache.GenericNamespaceLister
}

func (t *tolerantISDTemplateNamespaceLister) Get(name string) (*v1alpha1.ISDTemplate, error) {
	object, err := t.delegate.Get(name)
	if err != nil {
		return nil, err
	}
	v := &v1alpha1.ISDTemplate{}
	err = convertObject(object, v)
	return v, err
}

func (t *tolerantISDTemplateNamespaceLister) List(selector labels.Selector) ([]*v1alpha1.ISDTemplate, error) {
	objects, err := t.delegate.List(selector)
	if err != nil {
		return nil, err
	}
	return convertObjectsToISDTemplates(objects)
}

func convertObjectsToISDTemplates(objects []runtime.Object) ([]*v1alpha1.ISDTemplate, error) {
	var firstErr error
	vs := make([]*v1alpha1.ISDTemplate, len(objects))
	for i, obj := range objects {
		vs[i] = &v1alpha1.ISDTemplate{}
		err := convertObject(obj, vs[i])
		if err != nil && firstErr != nil {
			firstErr = err
		}
	}
	return vs, firstErr
}
