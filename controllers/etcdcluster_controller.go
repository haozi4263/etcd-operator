/*
Copyright 2021 jude.

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

package controllers

import (
	"context"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	etcdv1alpha1 "github.com/haozi4263/etcd-operator/api/v1alpha1"
)

// EtcdClusterReconciler reconciles a EtcdCluster object
type EtcdClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=etcd.shimo.im,resources=etcdclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=etcd.shimo.im,resources=etcdclusters/status,verbs=get;update;patch

// 生成rbac
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete

func (r *EtcdClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log.WithValues("etcdcluster", req.NamespacedName)

	// etcd 部署逻辑  获取EtcdCluster实例
	var etcdCluster etcdv1alpha1.EtcdCluster
	// r.Client.get 直接获取apiserver 没有获取本地的informer indexer
	// 更新时候可以用r.Client.update
	if err := r.Get(ctx, req.NamespacedName, &etcdCluster); err != nil {
		// 如果EtcdCluster是被删除的，那么我们应该忽略掉
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// 已经获取到EtcdCluster实例
	// 创建或更新对应的 StatefulSet和Headless Svc对象
	// CreateOrUpdate
	// 调谐： 观察当前到状态和期望的状态进行对比

	// CreateOrUpdate Service
	var svc corev1.Service
	svc.Name = etcdCluster.Name
	svc.Namespace = etcdCluster.Namespace
	// 如果svc不存在创建，存在则更新
	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &svc, func() error {
			// 调谐的函数必须在这里实现，实际上就是去拼装svc
			MutateHeadlessSvc(&etcdCluster, &svc)
			return controllerutil.SetControllerReference(&etcdCluster, &svc, r.Scheme)
		})
		log.Info("CreateOrUpdate Result", "Service", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	// CreateOrUpdate StateFulSet
	var sts appsv1.StatefulSet
	sts.Name = etcdCluster.Name
	sts.Namespace = etcdCluster.Namespace

	if err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		or, err := ctrl.CreateOrUpdate(ctx, r, &sts, func() error {
			// 调谐的函数必须在这里实现，实际上就是去拼装sts
			MutateStatefulSet(&etcdCluster, &sts)
			return controllerutil.SetControllerReference(&etcdCluster, &sts, r.Scheme)
		})
		log.Info("CreateOrUpdate Result", "StatefulSet", or)
		return err
	}); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *EtcdClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&etcdv1alpha1.EtcdCluster{}).
		Owns(&appsv1.StatefulSet{}). // 将sts加入watch
		Owns(&corev1.Service{}).     // 将svc加入watch
		Complete(r)
}
