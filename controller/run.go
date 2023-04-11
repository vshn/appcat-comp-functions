package controller

import (
	"github.com/go-logr/logr"
	"github.com/urfave/cli/v2"
	"github.com/vshn/appcat-comp-functions/runtime"
	vshnv1 "github.com/vshn/component-appcat/apis/vshn/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
)

func init() {
	_ = runtime.AddToScheme(vshnv1.SchemeBuilder.SchemeBuilder)

}

// RunController will run the controller mode of the composition function runner.
func RunController(cli *cli.Context) error {

	log := logr.FromContextOrDiscard(cli.Context)

	ctrl.SetLogger(log)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 runtime.GetScheme(),
		MetricsBindAddress:     cli.String("metrics-addr"),
		Port:                   9443,
		HealthProbeBindAddress: cli.String("health-addr"),
		LeaderElection:         cli.Bool("leader-elect"),
		LeaderElectionID:       "35t6u158.appcat.vshn.io",
	})
	if err != nil {
		return err
	}

	err = (&postgreSQLDeletionProtectionReconciler{
		Client: mgr.GetClient(),
		Scheme: mgr.GetScheme(),
	}).SetupWithManager(mgr)
	if err != nil {
		return err
	}

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		return err
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		return err
	}

	return mgr.Start(ctrl.SetupSignalHandler())
}
