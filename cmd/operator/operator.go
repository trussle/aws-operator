package main

import (
	"flag"

	"github.com/SimonRichardson/flagset"
	"github.com/SimonRichardson/gexec"
	"github.com/pkg/errors"
	"github.com/trussle/aws-operator/pkg/iam"
	"github.com/trussle/aws-operator/pkg/sqs"
	apiextcs "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

const (
	defaultNamespace = "applications"
)

func runOperator(args []string) error {
	// flags for the operator command
	var (
		flags = flagset.NewFlagSet("operator", flag.ExitOnError)

		debug     = flags.Bool("debug", false, "debug logging")
		namespace = flags.String("namespace", defaultNamespace, "namespace for operator")
	)
	flags.Usage = usageFor(flags, "operator [flags]")
	if err := flags.Parse(args); err != nil {
		return nil
	}

	restCfg, err := rest.InClusterConfig()
	if err != nil {
		return errors.Wrap(err, "cluster config")
	}

	clientSet, err := apiextcs.NewForConfig(restCfg)
	if err != nil {
		return errors.Wrap(err, "api extensions config")
	}

	crdcs, _, err := iam.NewClient(restCfg)
	if err != nil {
		return errors.Wrap(err, "iam client")
	}

	sqsCrd, sqsScheme, err := sqs.NewClient(restCfg)
	if err != nil {
		return errors.Wrap(err, "sqs client")
	}

	// Execution group.
	g := gexec.NewGroup()
	gexec.Block(g)
	{
		controller, err := iam.New(*namespace,
			clientSet,
			crdcs,
		)
		if err != nil {
			return errors.Wrap(err, "iam")
		}

		g.Add(func() error {
			return controller.Run(clientSet, sqsCrd)
		}, func(error) {
			controller.Stop()
		})
	}
	{
		controller, err := sqs.New(*namespace,
			clientSet,
			crdcs,
		)
		if err != nil {
			return errors.Wrap(err, "sqs")
		}

		g.Add(func() error {
			return controller.Run(clientSet, sqsCrd)
		}, func(error) {
			controller.Stop()
		})
	}
	gexec.Interrupt(g)
	return g.Run()
}
