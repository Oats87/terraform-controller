/*
Copyright 2021 Rancher Labs, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by main. DO NOT EDIT.

package v1

import (
	"context"
	"time"

	"github.com/rancher/lasso/pkg/client"
	"github.com/rancher/lasso/pkg/controller"
	v1 "github.com/rancher/terraform-controller/pkg/apis/terraformcontroller.cattle.io/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/condition"
	"github.com/rancher/wrangler/pkg/generic"
	"github.com/rancher/wrangler/pkg/kv"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type StateHandler func(string, *v1.State) (*v1.State, error)

type StateController interface {
	generic.ControllerMeta
	StateClient

	OnChange(ctx context.Context, name string, sync StateHandler)
	OnRemove(ctx context.Context, name string, sync StateHandler)
	Enqueue(namespace, name string)
	EnqueueAfter(namespace, name string, duration time.Duration)

	Cache() StateCache
}

type StateClient interface {
	Create(*v1.State) (*v1.State, error)
	Update(*v1.State) (*v1.State, error)
	UpdateStatus(*v1.State) (*v1.State, error)
	Delete(namespace, name string, options *metav1.DeleteOptions) error
	Get(namespace, name string, options metav1.GetOptions) (*v1.State, error)
	List(namespace string, opts metav1.ListOptions) (*v1.StateList, error)
	Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error)
	Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.State, err error)
}

type StateCache interface {
	Get(namespace, name string) (*v1.State, error)
	List(namespace string, selector labels.Selector) ([]*v1.State, error)

	AddIndexer(indexName string, indexer StateIndexer)
	GetByIndex(indexName, key string) ([]*v1.State, error)
}

type StateIndexer func(obj *v1.State) ([]string, error)

type stateController struct {
	controller    controller.SharedController
	client        *client.Client
	gvk           schema.GroupVersionKind
	groupResource schema.GroupResource
}

func NewStateController(gvk schema.GroupVersionKind, resource string, namespaced bool, controller controller.SharedControllerFactory) StateController {
	c := controller.ForResourceKind(gvk.GroupVersion().WithResource(resource), gvk.Kind, namespaced)
	return &stateController{
		controller: c,
		client:     c.Client(),
		gvk:        gvk,
		groupResource: schema.GroupResource{
			Group:    gvk.Group,
			Resource: resource,
		},
	}
}

func FromStateHandlerToHandler(sync StateHandler) generic.Handler {
	return func(key string, obj runtime.Object) (ret runtime.Object, err error) {
		var v *v1.State
		if obj == nil {
			v, err = sync(key, nil)
		} else {
			v, err = sync(key, obj.(*v1.State))
		}
		if v == nil {
			return nil, err
		}
		return v, err
	}
}

func (c *stateController) Updater() generic.Updater {
	return func(obj runtime.Object) (runtime.Object, error) {
		newObj, err := c.Update(obj.(*v1.State))
		if newObj == nil {
			return nil, err
		}
		return newObj, err
	}
}

func UpdateStateDeepCopyOnChange(client StateClient, obj *v1.State, handler func(obj *v1.State) (*v1.State, error)) (*v1.State, error) {
	if obj == nil {
		return obj, nil
	}

	copyObj := obj.DeepCopy()
	newObj, err := handler(copyObj)
	if newObj != nil {
		copyObj = newObj
	}
	if obj.ResourceVersion == copyObj.ResourceVersion && !equality.Semantic.DeepEqual(obj, copyObj) {
		return client.Update(copyObj)
	}

	return copyObj, err
}

func (c *stateController) AddGenericHandler(ctx context.Context, name string, handler generic.Handler) {
	c.controller.RegisterHandler(ctx, name, controller.SharedControllerHandlerFunc(handler))
}

func (c *stateController) AddGenericRemoveHandler(ctx context.Context, name string, handler generic.Handler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), handler))
}

func (c *stateController) OnChange(ctx context.Context, name string, sync StateHandler) {
	c.AddGenericHandler(ctx, name, FromStateHandlerToHandler(sync))
}

func (c *stateController) OnRemove(ctx context.Context, name string, sync StateHandler) {
	c.AddGenericHandler(ctx, name, generic.NewRemoveHandler(name, c.Updater(), FromStateHandlerToHandler(sync)))
}

func (c *stateController) Enqueue(namespace, name string) {
	c.controller.Enqueue(namespace, name)
}

func (c *stateController) EnqueueAfter(namespace, name string, duration time.Duration) {
	c.controller.EnqueueAfter(namespace, name, duration)
}

func (c *stateController) Informer() cache.SharedIndexInformer {
	return c.controller.Informer()
}

func (c *stateController) GroupVersionKind() schema.GroupVersionKind {
	return c.gvk
}

func (c *stateController) Cache() StateCache {
	return &stateCache{
		indexer:  c.Informer().GetIndexer(),
		resource: c.groupResource,
	}
}

func (c *stateController) Create(obj *v1.State) (*v1.State, error) {
	result := &v1.State{}
	return result, c.client.Create(context.TODO(), obj.Namespace, obj, result, metav1.CreateOptions{})
}

func (c *stateController) Update(obj *v1.State) (*v1.State, error) {
	result := &v1.State{}
	return result, c.client.Update(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *stateController) UpdateStatus(obj *v1.State) (*v1.State, error) {
	result := &v1.State{}
	return result, c.client.UpdateStatus(context.TODO(), obj.Namespace, obj, result, metav1.UpdateOptions{})
}

func (c *stateController) Delete(namespace, name string, options *metav1.DeleteOptions) error {
	if options == nil {
		options = &metav1.DeleteOptions{}
	}
	return c.client.Delete(context.TODO(), namespace, name, *options)
}

func (c *stateController) Get(namespace, name string, options metav1.GetOptions) (*v1.State, error) {
	result := &v1.State{}
	return result, c.client.Get(context.TODO(), namespace, name, result, options)
}

func (c *stateController) List(namespace string, opts metav1.ListOptions) (*v1.StateList, error) {
	result := &v1.StateList{}
	return result, c.client.List(context.TODO(), namespace, result, opts)
}

func (c *stateController) Watch(namespace string, opts metav1.ListOptions) (watch.Interface, error) {
	return c.client.Watch(context.TODO(), namespace, opts)
}

func (c *stateController) Patch(namespace, name string, pt types.PatchType, data []byte, subresources ...string) (*v1.State, error) {
	result := &v1.State{}
	return result, c.client.Patch(context.TODO(), namespace, name, pt, data, result, metav1.PatchOptions{}, subresources...)
}

type stateCache struct {
	indexer  cache.Indexer
	resource schema.GroupResource
}

func (c *stateCache) Get(namespace, name string) (*v1.State, error) {
	obj, exists, err := c.indexer.GetByKey(namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(c.resource, name)
	}
	return obj.(*v1.State), nil
}

func (c *stateCache) List(namespace string, selector labels.Selector) (ret []*v1.State, err error) {

	err = cache.ListAllByNamespace(c.indexer, namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1.State))
	})

	return ret, err
}

func (c *stateCache) AddIndexer(indexName string, indexer StateIndexer) {
	utilruntime.Must(c.indexer.AddIndexers(map[string]cache.IndexFunc{
		indexName: func(obj interface{}) (strings []string, e error) {
			return indexer(obj.(*v1.State))
		},
	}))
}

func (c *stateCache) GetByIndex(indexName, key string) (result []*v1.State, err error) {
	objs, err := c.indexer.ByIndex(indexName, key)
	if err != nil {
		return nil, err
	}
	result = make([]*v1.State, 0, len(objs))
	for _, obj := range objs {
		result = append(result, obj.(*v1.State))
	}
	return result, nil
}

type StateStatusHandler func(obj *v1.State, status v1.StateStatus) (v1.StateStatus, error)

type StateGeneratingHandler func(obj *v1.State, status v1.StateStatus) ([]runtime.Object, v1.StateStatus, error)

func RegisterStateStatusHandler(ctx context.Context, controller StateController, condition condition.Cond, name string, handler StateStatusHandler) {
	statusHandler := &stateStatusHandler{
		client:    controller,
		condition: condition,
		handler:   handler,
	}
	controller.AddGenericHandler(ctx, name, FromStateHandlerToHandler(statusHandler.sync))
}

func RegisterStateGeneratingHandler(ctx context.Context, controller StateController, apply apply.Apply,
	condition condition.Cond, name string, handler StateGeneratingHandler, opts *generic.GeneratingHandlerOptions) {
	statusHandler := &stateGeneratingHandler{
		StateGeneratingHandler: handler,
		apply:                  apply,
		name:                   name,
		gvk:                    controller.GroupVersionKind(),
	}
	if opts != nil {
		statusHandler.opts = *opts
	}
	controller.OnChange(ctx, name, statusHandler.Remove)
	RegisterStateStatusHandler(ctx, controller, condition, name, statusHandler.Handle)
}

type stateStatusHandler struct {
	client    StateClient
	condition condition.Cond
	handler   StateStatusHandler
}

func (a *stateStatusHandler) sync(key string, obj *v1.State) (*v1.State, error) {
	if obj == nil {
		return obj, nil
	}

	origStatus := obj.Status.DeepCopy()
	obj = obj.DeepCopy()
	newStatus, err := a.handler(obj, obj.Status)
	if err != nil {
		// Revert to old status on error
		newStatus = *origStatus.DeepCopy()
	}

	if a.condition != "" {
		if errors.IsConflict(err) {
			a.condition.SetError(&newStatus, "", nil)
		} else {
			a.condition.SetError(&newStatus, "", err)
		}
	}
	if !equality.Semantic.DeepEqual(origStatus, &newStatus) {
		if a.condition != "" {
			// Since status has changed, update the lastUpdatedTime
			a.condition.LastUpdated(&newStatus, time.Now().UTC().Format(time.RFC3339))
		}

		var newErr error
		obj.Status = newStatus
		newObj, newErr := a.client.UpdateStatus(obj)
		if err == nil {
			err = newErr
		}
		if newErr == nil {
			obj = newObj
		}
	}
	return obj, err
}

type stateGeneratingHandler struct {
	StateGeneratingHandler
	apply apply.Apply
	opts  generic.GeneratingHandlerOptions
	gvk   schema.GroupVersionKind
	name  string
}

func (a *stateGeneratingHandler) Remove(key string, obj *v1.State) (*v1.State, error) {
	if obj != nil {
		return obj, nil
	}

	obj = &v1.State{}
	obj.Namespace, obj.Name = kv.RSplit(key, "/")
	obj.SetGroupVersionKind(a.gvk)

	return nil, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects()
}

func (a *stateGeneratingHandler) Handle(obj *v1.State, status v1.StateStatus) (v1.StateStatus, error) {
	objs, newStatus, err := a.StateGeneratingHandler(obj, status)
	if err != nil {
		return newStatus, err
	}

	return newStatus, generic.ConfigureApplyForObject(a.apply, obj, &a.opts).
		WithOwner(obj).
		WithSetID(a.name).
		ApplyObjects(objs...)
}
