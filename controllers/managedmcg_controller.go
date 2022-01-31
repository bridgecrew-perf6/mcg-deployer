/*
Copyright 2022.

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

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/go-logr/logr"
	odfv1alpha1 "github.com/red-hat-data-services/odf-operator/api/v1alpha1"
	ocsv1 "github.com/red-hat-storage/ocs-operator/api/v1"
	mcgv1alpha1 "github.com/rexagod/mcg-deployer/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	managedMCGObjectName = "managedmcg"
	storageSystemName    = "ocs-storagecluster-storagesystem"
)

// ManagedMCGReconciler reconciles a ManagedMCG object
type ManagedMCGReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger

	ctx            context.Context
	namespace      string
	managedMCG     *mcgv1alpha1.ManagedMCG
	storagesystem  *odfv1alpha1.StorageSystem
	storagecluster *ocsv1.StorageCluster
}

// Required for ManagedMCG
//+kubebuilder:rbac:groups=mcg.openshift.io,resources=managedmcgs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=mcg.openshift.io,resources=managedmcgs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=mcg.openshift.io,resources=managedmcgs/finalizers,verbs=update
// Required for StorageSystem
//+kubebuilder:rbac:groups=storage.openshift.io,resources=storagesystems,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=storage.openshift.io,resources=storagesystems/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=storage.openshift.io,resources=storagesystems/finalizers,verbs=update
// Required for Noobaa
//+kubebuilder:rbac:groups=noobaas.noobaa.io,resources=noobaas,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=noobaas.noobaa.io,resources=noobaas/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=noobaas.noobaa.io,resources=noobaas/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *ManagedMCGReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var err error
	log := r.Log.WithValues("req.Namespace", req.Namespace, "req.Name", req.Name)
	log.Info("Starting reconcile for ManagedMCG")

	// Set default values in the reconciler
	r.ctx = context.Background()
	r.namespace = req.NamespacedName.Namespace
	r.managedMCG = &mcgv1alpha1.ManagedMCG{}
	r.managedMCG.Name = req.NamespacedName.Name
	r.managedMCG.Namespace = r.namespace
	r.storagecluster = &ocsv1.StorageCluster{}
	r.storagesystem = &odfv1alpha1.StorageSystem{}
	r.storagesystem.Name = storageSystemName

	// Fetch the ManagedMCG instance
	err = r.Get(r.ctx, req.NamespacedName, r.managedMCG)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			log.Error(err, "unable to fetch ManagedMCG")
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// TODO Check if ODF operator is successfully installed and up, reconcile till it isn't
	// Create a StorageSystem CR if it does not exist
	err = r.Get(r.ctx, types.NamespacedName{Name: r.storagesystem.Name, Namespace: r.namespace}, r.storagesystem)
	if err != nil {
		if errors.IsNotFound(err) {
			log.Info("Creating StorageSystem CR")
			err = r.Create(r.ctx, r.storagesystem)
			if err != nil {
				log.Error(err, "unable to create StorageSystem CR")
				return ctrl.Result{}, err
			}
		} else {
			log.Error(err, "unable to fetch StorageSystem CR")
			return ctrl.Result{}, err
		}
	}

	// TODO Reconcile till StorageCluster CR is Ready

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ManagedMCGReconciler) SetupWithManager(mgr ctrl.Manager) error {
	ctrlOptions := controller.Options{
		MaxConcurrentReconciles: 1,
	}
	managedMCGPredicates := builder.WithPredicates(
		predicate.GenerationChangedPredicate{},
	)
	enqueueManagedMCGRequest := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      e.Object.GetName(),
				Namespace: e.Object.GetNamespace(),
			}})
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      e.ObjectNew.GetName(),
				Namespace: e.ObjectNew.GetNamespace(),
			}})
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      e.Object.GetName(),
				Namespace: e.Object.GetNamespace(),
			}})
		},
		GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      e.Object.GetName(),
				Namespace: e.Object.GetNamespace(),
			}})
		},
	}
	return ctrl.NewControllerManagedBy(mgr).
		WithOptions(ctrlOptions).
		For(&mcgv1alpha1.ManagedMCG{}, managedMCGPredicates).
		Watches(
			&source.Kind{Type: &ocsv1.StorageCluster{}},
			&enqueueManagedMCGRequest,
			managedMCGPredicates,
		).
		Complete(r)
}
