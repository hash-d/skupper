package execute

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/base"
)

type SkupperServiceCreate struct {

	// Required

	Namespace *base.ClusterContext
	Name      string
	Port      []string

	// Optional

	Aggregate    string
	EnableTls    bool
	EventChannel bool
	Protocol     string

	Runner       *frame2.Run
	AutoTeardown bool
}

// TODO move implemetnation to CLI, make this one base-independent
func (ssc SkupperServiceCreate) Execute() error {

	if len(ssc.Port) == 0 || ssc.Name == "" {
		return fmt.Errorf("SkupperServiceCreate configuration error: Name and Port must be specified")
	}

	args := []string{"service", "create", ssc.Name}

	args = append(args, strings.Join(ssc.Port, ","))

	if ssc.Aggregate != "" {
		args = append(args, "--aggregate", ssc.Aggregate)
	}

	if ssc.Protocol != "" {
		args = append(args, "--protocol", ssc.Protocol)
	}

	if ssc.EnableTls {
		args = append(args, "--enable-tls")
	}

	if ssc.EventChannel {
		args = append(args, "--event-channel")
	}

	phase := frame2.Phase{
		Runner: ssc.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: ssc.Namespace,
					Args:           args,
				},
			},
		},
	}
	phase.Run()

	return nil
}

func (ssc SkupperServiceCreate) Teardown() frame2.Executor {

	if ssc.AutoTeardown {
		return &SkupperServiceDelete{
			Namespace: ssc.Namespace,
			ArgName:   ssc.Name,
			// TODO: change this for some constant or env-variable
			Wait: 2 * time.Minute,
		}
	}
	return nil
}

type SkupperServiceDelete struct {
	// Required

	Namespace *base.ClusterContext
	ArgName   string

	// Optional

	Runner *frame2.Run
	Wait   time.Duration
	Ctx    context.Context
}

func (ssd SkupperServiceDelete) Execute() error {
	if ssd.ArgName == "" {
		return fmt.Errorf("SkupperServiceDelete configuration error: Name is requried")
	}

	retry := frame2.RetryOptions{}

	ctx := ssd.Ctx
	if ctx == nil {
		ctx = context.Background()
	}

	var validator frame2.Validator

	var fn context.CancelFunc
	if ssd.Wait != 0 {
		ctx, fn = context.WithTimeout(ctx, ssd.Wait)
		defer fn()
		validator = &K8SServiceGet{
			Namespace: ssd.Namespace,
			Name:      ssd.ArgName,
		}
		retry.KeepTrying = true
	}
	retry.Ctx = ctx

	phase := frame2.Phase{
		Runner: ssd.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: ssd.Namespace,
					Args:           []string{"service", "delete", ssd.ArgName},
				},
				Validator:      validator,
				ValidatorRetry: retry,
				ExpectError:    true,
			},
		},
	}
	return phase.Run()
}

type SkupperServiceBind struct {

	// Required

	Namespace  *base.ClusterContext
	Name       string
	TargetType string
	TargetName string

	// Optional

	Protocol               string
	PublishNotReadyAddress bool
	TargetPort             []string

	AutoTeardown bool
	Runner       *frame2.Run
}

// TODO move implemetnation to CLI, make this one base-independent
func (ssb SkupperServiceBind) Execute() error {

	if ssb.TargetType == "" || ssb.TargetName == "" || ssb.Name == "" {
		return fmt.Errorf("SkupperServiceBind configuration error: Name, TargetName and TargetType must be specified")
	}

	args := []string{"service", "bind", ssb.Name, ssb.TargetType, ssb.TargetName}

	if len(ssb.TargetPort) > 0 {
		args = append(args, "--target-port", strings.Join(ssb.TargetPort, ","))
	}

	if ssb.PublishNotReadyAddress {
		args = append(args, "--publish-not-ready-addresses")
	}

	if ssb.Protocol != "" {
		args = append(args, "--protocol", ssb.Protocol)
	}

	phase := frame2.Phase{
		Runner: ssb.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: ssb.Namespace,
					Args:           args,
				},
			},
		},
	}
	phase.Run()

	return nil
}

func (ssb SkupperServiceBind) Teardown() frame2.Executor {
	if ssb.AutoTeardown {
		return SkupperServiceUnbind{
			Namespace:  ssb.Namespace,
			Name:       ssb.Name,
			TargetType: ssb.TargetType,
			TargetName: ssb.TargetName,
		}
	}
	return nil
}

type SkupperServiceUnbind struct {

	// Required

	Namespace  *base.ClusterContext
	Name       string
	TargetType string
	TargetName string

	// Optional

	Runner *frame2.Run
}

// TODO move implemetnation to CLI, make this one base-independent
func (ssub SkupperServiceUnbind) Execute() error {

	if ssub.TargetType == "" || ssub.TargetName == "" || ssub.Name == "" {
		return fmt.Errorf("SkupperServiceUnbind configuration error: Name, TargetName and TargetType must be specified")
	}

	args := []string{"service", "unbind", ssub.Name, ssub.TargetType, ssub.TargetName}

	phase := frame2.Phase{
		Runner: ssub.Runner,
		MainSteps: []frame2.Step{
			{
				Modify: &CliSkupper{
					ClusterContext: ssub.Namespace,
					Args:           args,
				},
			},
		},
	}
	return phase.Run()
}
