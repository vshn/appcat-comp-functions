package controller

import (
	"context"
	"time"

	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/go-logr/logr"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

const (
	finalizerName = "appcat.io/deletionProtection"
)

type postgreSQLDeletionProtectionReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (p *postgreSQLDeletionProtectionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx).WithValues("namespace", req.Namespace, "instance", req.Name)

	inst := &vshnv1.VSHNPostgreSQL{}
	err := p.Get(ctx, req.NamespacedName, inst)
	if apierrors.IsNotFound(err) {
		log.Info("Instance deleted")
		return ctrl.Result{}, nil
	}

	protectionEnabled := inst.Spec.Parameters.Backup.DeletionProtection
	retention := inst.Spec.Parameters.Backup.DeletionRetention

	xInst, err := p.getPostgreSQLComposite(ctx, log, inst)
	if err != nil && apierrors.IsNotFound(err) {
		log.V(1).Info("Composite was not found")
		return ctrl.Result{RequeueAfter: time.Second * 30}, nil
	} else if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 30}, errors.Wrap(err, "Could not get composite")
	}

	xBaseObj := &vshnv1.XVSHNPostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name: xInst.Name,
		},
	}

	// TODO: we might want to think about moving this to it's own reconciler.
	// If the claim gets deleted, but the composite not for $reason, then it will never be reconciled again.
	_, err = handleDeletionProtection(ctx, log, xBaseObj, xInst, p.Client, protectionEnabled, retention)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 30}, err
	}

	baseObj := &vshnv1.VSHNPostgreSQL{
		ObjectMeta: metav1.ObjectMeta{
			Name:      inst.Name,
			Namespace: inst.Namespace,
		},
	}

	requeueTime, err := handleDeletionProtection(ctx, log, baseObj, inst, p.Client, protectionEnabled, retention)
	if err != nil {
		return ctrl.Result{RequeueAfter: time.Second * 30}, err
	}

	return ctrl.Result{RequeueAfter: requeueTime}, err
}

func (p *postgreSQLDeletionProtectionReconciler) getPostgreSQLComposite(ctx context.Context, log logr.Logger, inst *vshnv1.VSHNPostgreSQL) (*vshnv1.XVSHNPostgreSQL, error) {
	log.V(1).Info("Getting PostgreSQL composite", "name", inst.Spec.ResourceRef.Name)
	xInst := &vshnv1.XVSHNPostgreSQL{}
	err := p.Get(ctx, types.NamespacedName{Name: inst.Spec.ResourceRef.Name}, xInst)
	if err != nil {
		return nil, err
	}
	log.V(1).Info("Found composite")
	return xInst, nil
}

// SetupWithManager sets up the controller with the Manager.
func (p *postgreSQLDeletionProtectionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vshnv1.VSHNPostgreSQL{}).
		Complete(p)
}
