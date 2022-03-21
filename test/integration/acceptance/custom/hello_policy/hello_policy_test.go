//go:build integration || cli || examples || policy
// +build integration cli examples policy

// TODO:
// - link and policy
//   - policy destroys link on source - rebuild
//   - policy destroys link on dest - ??
// - "Not authorized service" on skupper service status
// - t.Run names (standard, Polarion)
// - Enhance to include annotation-based exposing
// - Console

package hello_policy

import (
	"log"
	"os"
	"testing"

	"github.com/skupperproject/skupper/api/types"
	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/skupper/cli"
	"github.com/skupperproject/skupper/test/utils/skupper/cli/link"
	"github.com/skupperproject/skupper/test/utils/skupper/cli/service"
	"github.com/skupperproject/skupper/test/utils/skupper/cli/token"
	"gotest.tools/assert"
)

// TestHelloPolicy is a test that runs the hello-world-example
// scenario using just the "skupper" binary, which must be available
// in the PATH.
// It is a copy of the test at test/integration/examples/custom/helloworld/,
// adapted for Policy testing
func TestHelloPolicy(t *testing.T) {

	// First, validate if skupper binary is in the PATH, or fail the test
	log.Printf("Running 'skupper --help' to determine if skupper binary is available")
	_, _, err := cli.RunSkupperCli([]string{"--help"})
	if err != nil {
		t.Fatalf("skupper binary is not available")
	}

	needs := base.ClusterNeeds{
		NamespaceId:     "hello-policy",
		PublicClusters:  1,
		PrivateClusters: 1,
	}
	runner := &base.ClusterTestRunnerBase{}
	if err := runner.Validate(needs); err != nil {
		t.Skipf("%s", err)
	}
	_, err = runner.Build(needs, nil)
	assert.Assert(t, err)

	// getting public and private contexts
	pub, err := runner.GetPublicContext(1)
	assert.Assert(t, err)
	prv, err := runner.GetPrivateContext(1)
	assert.Assert(t, err)

	// creating namespaces
	assert.Assert(t, pub.CreateNamespace())
	assert.Assert(t, prv.CreateNamespace())

	// teardown once test completes
	tearDownFn := func() {
		t.Log("entering teardown")
		t.Log("Removing pub namespace")
		_ = pub.DeleteNamespace()
		t.Log("Removing prv namespace")
		_ = prv.DeleteNamespace()
		t.Log("Removing cluster role skupper-service-controller from the CRD definition")
		pub.VanClient.KubeClient.RbacV1().ClusterRoles().Delete("skupper-service-controller", nil)
		t.Log("Removing CRD")
		pub.KubectlExec("delete crd skupperclusterpolicies.skupper.io")
		t.Log("tearDown completed")
	}
	defer tearDownFn()
	base.HandleInterruptSignal(func() {
		tearDownFn()
	})

	// Creating a local directory for storing the token
	testPath := "./tmp/"
	_ = os.Mkdir(testPath, 0755)

	// deploying frontend and backend services
	assert.Assert(t, deployResources(pub, prv))

	// These test scenarios allow defining a set of skupper cli
	// commands to be executed as a workflow, against specific
	// clusters. Each execution is validated accordingly by its
	// SkupperCommandTester implementation.
	//
	// The idea is to cover most of the main skupper commands
	// as we run the hello-world-example so that all manipulation
	// is performed just by the skupper binary, while each
	// SkupperCommandTester implementation validates necessary
	// output or resources in the cluster to certify the command
	// was executed correctly.
	initSteps := []cli.TestScenario{
		{
			Name: "initialize",
			Tasks: []cli.SkupperTask{
				{Ctx: pub, Commands: []cli.SkupperCommandTester{
					// skupper init - interior mode, enabling console and internal authentication
					&cli.InitTester{
						ConsoleAuth:         "internal",
						ConsoleUser:         "internal",
						ConsolePassword:     "internal",
						RouterMode:          "interior",
						EnableConsole:       true,
						EnableRouterConsole: true,
					},
					// skupper status - verify initialized as interior
					&cli.StatusTester{
						RouterMode:          "interior",
						ConsoleEnabled:      true,
						ConsoleAuthInternal: true,
					},
				}},
				{Ctx: prv, Commands: []cli.SkupperCommandTester{
					// skupper init - edge mode, no console and unsecured
					&cli.InitTester{
						ConsoleAuth:           "unsecured",
						ConsoleUser:           "admin",
						ConsolePassword:       "admin",
						Ingress:               "none",
						RouterDebugMode:       "gdb",
						RouterLogging:         "trace",
						RouterMode:            "edge",
						SiteName:              "private",
						EnableConsole:         false,
						EnableRouterConsole:   false,
						RouterCPU:             "100m",
						RouterMemory:          "32Mi",
						ControllerCPU:         "50m",
						ControllerMemory:      "16Mi",
						RouterCPULimit:        "600m",
						RouterMemoryLimit:     "500Mi",
						ControllerCPULimit:    "600m",
						ControllerMemoryLimit: "500Mi",
						//ConsoleIngress:      "none",
					},
					// skupper status - verify initialized as edge
					&cli.StatusTester{
						RouterMode: "edge",
						SiteName:   "private",
					},
				}},
			},
		},
	}

	connectSteps := cli.TestScenario{
		Name: "connect-sites",
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper token create - verify token has been created
				&token.CreateTester{
					Name:     "public",
					FileName: testPath + "public-hello-world-1.token.yaml",
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper link create - connect to public and verify connection created
				&link.CreateTester{
					TokenFile: testPath + "public-hello-world-1.token.yaml",
					Name:      "public",
					Cost:      1,
				},
			}},
		},
	}

	//validateConnSteps Steps to confirm a link exists from the private namespace to the public one
	validateConnSteps := cli.TestScenario{
		Name: "validate-sites-connected",
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper status - verify sites are connected
				&cli.StatusTester{
					RouterMode:          "interior",
					ConnectedSites:      1,
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper status - verify sites are connected
				&cli.StatusTester{
					RouterMode:     "edge",
					SiteName:       "private",
					ConnectedSites: 1,
				},
				// skupper link status - testing all links
				&link.StatusTester{
					Name:   "public",
					Active: true,
				},
				// skupper link status - now using link name and a 10 secs wait
				&link.StatusTester{
					Name:   "public",
					Active: true,
					Wait:   10,
				},
			}},
		},
	}

	serviceCreateBindSteps := cli.TestScenario{
		Name: "service-create-bind",
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper service create - creates the frontend service and verify
				&service.CreateTester{
					Name:    "hello-world-frontend",
					Port:    8080,
					Mapping: "http",
				},
				// skupper service status - verify frontend service is exposed
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
					},
				},
				// skupper status - verify frontend service is exposed
				&cli.StatusTester{
					RouterMode:          "interior",
					ConnectedSites:      1,
					ExposedServices:     1,
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper service create - creates the backend service and verify
				&service.CreateTester{
					Name:    "hello-world-backend",
					Port:    8080,
					Mapping: "http",
				},
				// skupper service status - validate status of the two created services without targets
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
					},
				},
				// skupper status - verify two services are now exposed
				&cli.StatusTester{
					RouterMode:      "edge",
					SiteName:        "private",
					ConnectedSites:  1,
					ExposedServices: 2,
				},
			}},
			// Binding the services
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper service bind - bind service to deployment and validate target has been defined
				&service.BindTester{
					ServiceName: "hello-world-frontend",
					TargetType:  "deployment",
					TargetName:  "hello-world-frontend",
					Protocol:    "http",
					TargetPort:  8080,
				},
				// skupper service status - validate status expecting frontend now has a target
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}, Targets: []types.ServiceInterfaceTarget{
							{Name: "hello-world-frontend", TargetPorts: map[int]int{8080: 8080}, Service: "hello-world-frontend"},
						}},
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
					},
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper service bind - bind service to deployment and validate target has been defined
				&service.BindTester{
					ServiceName: "hello-world-backend",
					TargetType:  "deployment",
					TargetName:  "hello-world-backend",
					Protocol:    "http",
					TargetPort:  8080,
				},
				// skupper service status - validate backend service now has a target
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}, Targets: []types.ServiceInterfaceTarget{
							{Name: "hello-world-backend", TargetPorts: map[int]int{8080: 8080}, Service: "hello-world-backend"},
						}},
					},
				},
			}},
		},
	}

	serviceUnbindDeleteSteps := cli.TestScenario{
		Name: "service-unbind-delete",
		Tasks: []cli.SkupperTask{
			// unbinding frontend and validating service status for public cluster
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper service unbind - unbind and verify service no longer has a target
				&service.UnbindTester{
					ServiceName: "hello-world-frontend",
					TargetType:  "deployment",
					TargetName:  "hello-world-frontend",
				},
				// skupper service status - validates no more target for frontend service
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
					},
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper service unbind - unbind and verify service no longer has a target
				&service.UnbindTester{
					ServiceName: "hello-world-backend",
					TargetType:  "deployment",
					TargetName:  "hello-world-backend",
				},
				// skupper service status - validates no more target for frontend service
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
					},
				},
			}},
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper service delete - removes exposed service and certify it is removed
				&service.DeleteTester{
					Name: "hello-world-frontend",
				},
				// skupper service status - verify only backend is available
				&service.StatusTester{
					ServiceInterfaces: []types.ServiceInterface{
						{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
					},
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper service delete - removes exposed service and certify it is removed
				&service.DeleteTester{
					Name: "hello-world-backend",
				},
				// skupper status - verify there is no exposed service
				&cli.StatusTester{
					RouterMode:      "edge",
					SiteName:        "private",
					ConnectedSites:  1,
					ExposedServices: 0,
				},
			}},
		},
	}

	exposeSteps := cli.TestScenario{
		Name: "expose",
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper expose - expose and ensure service is available
				&cli.ExposeTester{
					TargetType: "deployment",
					TargetName: "hello-world-frontend",
					Address:    "hello-world-frontend",
					Port:       8080,
					Protocol:   "http",
					TargetPort: 8080,
				},
				// skupper status - asserts that 1 service is exposed
				&cli.StatusTester{
					RouterMode:          "interior",
					ConnectedSites:      1,
					ExposedServices:     1,
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper expose - exposes backend and certify it is available
				&cli.ExposeTester{
					TargetType: "deployment",
					TargetName: "hello-world-backend",
					Address:    "hello-world-backend",
					Port:       8080,
					Protocol:   "http",
					TargetPort: 8080,
				},
				// skupper status - asserts that there are 2 exposed services
				&cli.StatusTester{
					RouterMode:      "edge",
					SiteName:        "private",
					ConnectedSites:  1,
					ExposedServices: 2,
				},
			}},
		},
	}

	unexposeSteps := cli.TestScenario{
		Name: "unexpose",
		Tasks: []cli.SkupperTask{
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				// skupper unexpose - unexpose and verify it has been removed
				&cli.UnexposeTester{
					TargetType: "deployment",
					TargetName: "hello-world-frontend",
					Address:    "hello-world-frontend",
				},
				// skupper status - verify only 1 service is exposed
				&cli.StatusTester{
					RouterMode:          "interior",
					ConnectedSites:      1,
					ExposedServices:     1,
					ConsoleEnabled:      true,
					ConsoleAuthInternal: true,
				},
			}},
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				// skupper unexpose - unexpose and verify it has been removed
				&cli.UnexposeTester{
					TargetType: "deployment",
					TargetName: "hello-world-backend",
					Address:    "hello-world-backend",
				},
				// skupper status - verify there is no exposed services
				&cli.StatusTester{
					RouterMode:      "edge",
					SiteName:        "private",
					ConnectedSites:  1,
					ExposedServices: 0,
				},
			}},
		},
	}

	versionSteps := cli.TestScenario{
		Name: "version",
		Tasks: []cli.SkupperTask{
			// skupper version - verify version is being reported accordingly
			{Ctx: pub, Commands: []cli.SkupperCommandTester{
				&cli.VersionTester{},
			}},
			// skupper version - verify version is being reported accordingly
			{Ctx: prv, Commands: []cli.SkupperCommandTester{
				&cli.VersionTester{},
			}},
		},
	}

	mainSteps := []cli.TestScenario{
		connectSteps,
		validateConnSteps,
		serviceCreateBindSteps,
		serviceUnbindDeleteSteps,
		exposeSteps,
		unexposeSteps,
		versionSteps,
	}

	checkStuffCameBackUp := []cli.TestScenario{
		validateConnSteps,
	}

	checkStuffCameDown := []cli.TestScenario{
		{
			Name: "validate-sites-disconnected",
			Tasks: []cli.SkupperTask{
				{Ctx: pub, Commands: []cli.SkupperCommandTester{
					// skupper status - verify sites are connected
					&cli.StatusTester{
						RouterMode:          "interior",
						ConnectedSites:      0,
						ConsoleEnabled:      true,
						ConsoleAuthInternal: true,
						PolicyEnabled:       true,
					},
				}},
				{Ctx: prv, Commands: []cli.SkupperCommandTester{
					// skupper status - verify sites are connected
					&cli.StatusTester{
						RouterMode:     "edge",
						SiteName:       "private",
						ConnectedSites: 0,
						PolicyEnabled:  true,
					},
					// skupper link status - testing all links
					&link.StatusTester{
						Name:   "public",
						Active: false,
					},
					// skupper link status - now using link name and a 10 secs wait
					&link.StatusTester{
						Name:   "public",
						Active: false,
						Wait:   10,
					},
				}},
			},
		}, {
			Name: "services-destroyed",
			Tasks: []cli.SkupperTask{
				{Ctx: pub, Commands: []cli.SkupperCommandTester{
					// skupper service status - verify frontend service is exposed
					&service.StatusTester{
						ServiceInterfaces: []types.ServiceInterface{
							{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
						},
						Absent: true,
					},
					// skupper status - verify frontend service is exposed
					&cli.StatusTester{
						RouterMode:          "interior",
						ConnectedSites:      0,
						ExposedServices:     0,
						ConsoleEnabled:      true,
						ConsoleAuthInternal: true,
						PolicyEnabled:       true,
					},
				}},
				{Ctx: prv, Commands: []cli.SkupperCommandTester{
					// skupper service status - validate status of the two created services without targets
					&service.StatusTester{
						ServiceInterfaces: []types.ServiceInterface{
							{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
							{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
						},
						Absent: true,
					},
					// skupper status - verify two services are now exposed
					&cli.StatusTester{
						RouterMode:      "edge",
						SiteName:        "private",
						ConnectedSites:  0,
						ExposedServices: 0,
						PolicyEnabled:   true,
					},
				}},
				// Binding the services
				{Ctx: pub, Commands: []cli.SkupperCommandTester{
					// skupper service bind - bind service to deployment and validate target has been defined
					&service.BindTester{
						ServiceName:     "hello-world-frontend",
						TargetType:      "deployment",
						TargetName:      "hello-world-frontend",
						Protocol:        "http",
						TargetPort:      8080,
						ExpectAuthError: true,
					},
					// skupper service status - validate status expecting frontend now has a target
					&service.StatusTester{
						ServiceInterfaces: []types.ServiceInterface{
							{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}, Targets: []types.ServiceInterfaceTarget{
								{Name: "hello-world-frontend", TargetPorts: map[int]int{8080: 8080}, Service: "hello-world-frontend"},
							}},
							{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}},
						},
						Absent: true,
					},
				}},
				{Ctx: prv, Commands: []cli.SkupperCommandTester{
					// skupper service bind - bind service to deployment and validate target has been defined
					&service.BindTester{
						ServiceName:     "hello-world-backend",
						TargetType:      "deployment",
						TargetName:      "hello-world-backend",
						Protocol:        "http",
						TargetPort:      8080,
						ExpectAuthError: true,
					},
					// skupper service status - validate backend service now has a target
					&service.StatusTester{
						ServiceInterfaces: []types.ServiceInterface{
							{Address: "hello-world-frontend", Protocol: "http", Ports: []int{8080}},
							{Address: "hello-world-backend", Protocol: "http", Ports: []int{8080}, Targets: []types.ServiceInterfaceTarget{
								{Name: "hello-world-backend", TargetPorts: map[int]int{8080: 8080}, Service: "hello-world-backend"},
							}},
						},
						Absent: true,
					},
				}},
			},
		},
	}

	deleteSteps := []cli.TestScenario{
		{
			Name: "skupper delete",
			Tasks: []cli.SkupperTask{
				// skupper delete - delete and verify resources have been removed
				{Ctx: pub, Commands: []cli.SkupperCommandTester{
					&cli.DeleteTester{},
					&cli.StatusTester{
						NotEnabled: true,
					},
				}},
				// skupper delete - delete and verify resources have been removed
				{Ctx: prv, Commands: []cli.SkupperCommandTester{
					&cli.DeleteTester{},
					&cli.StatusTester{
						NotEnabled: true,
					},
				}},
			},
		},
	}

	//	scenarios := append(append(initSteps, mainSteps...), deleteSteps...)

	// Running the scenarios
	t.Run("init", func(t *testing.T) { cli.RunScenarios(t, initSteps) })
	//	mainSteps = mainSteps
	t.Run("No CRD, all works", func(t *testing.T) { cli.RunScenarios(t, mainSteps) })
	t.Run("Re-expose service, for next test", func(t *testing.T) { cli.RunScenarios(t, []cli.TestScenario{exposeSteps}) })
	applyCrd(t, pub)
	t.Run("CRD added and no policy, all comes down", func(t *testing.T) { cli.RunScenarios(t, checkStuffCameDown) })
	t.Log("Removing CRD again, some resources should come back up")
	removeCrd(t, pub)
	t.Run("CRD removed, link should come back up", func(t *testing.T) { cli.RunScenarios(t, checkStuffCameBackUp) })
	t.Run("closing", func(t *testing.T) { cli.RunScenarios(t, deleteSteps) })

}
