package base

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	vanClient "github.com/skupperproject/skupper/client"
	"github.com/skupperproject/skupper/pkg/kube"
	"github.com/skupperproject/skupper/test/utils/k8s"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
)

// ClusterContext represents a cluster that is available for testing
type ClusterContext struct {
	Namespace  string
	nsCreated  bool
	KubeConfig string
	VanClient  *vanClient.VanClient
	Private    bool
	Id         int
}

func _exec(command string) ([]byte, error) {
	var output []byte
	var err error
	fmt.Println(command)
	cmd := exec.Command("sh", "-c", command)
	output, err = cmd.CombinedOutput()
	fmt.Println(string(output))
	return output, err
}

func (cc *ClusterContext) exec(main_command string, sub_command string) ([]byte, error) {
	return _exec("KUBECONFIG=" + cc.KubeConfig + " " + main_command + " " + cc.Namespace + " " + sub_command)
}

func (cc *ClusterContext) KubectlExec(command string) ([]byte, error) {
	return cc.exec("kubectl -n ", command)
}

func (cc *ClusterContext) CreateNamespace() error {
	if ShouldSkipNamespaceSetup() {
		log.Printf("Skipping namespace creation for %v", cc.Namespace)
		ns, err := cc.VanClient.KubeClient.CoreV1().Namespaces().Get(context.TODO(), cc.Namespace, metav1.GetOptions{})
		if err == nil {
			if ns != nil {
				// As we're skipping the creation of namespaces, we're adopting whatever
				// we find; we will destroy these when DeleteNamespace is called, unless
				// ShouldSkipNamespaceTeardown returns true.
				log.Printf("Reusing existing namespace %v", cc.Namespace)
				cc.nsCreated = true
			} else {
				// Assertion; this should never happen
				return fmt.Errorf("Namespace check returned nil response, but no errors")
			}
		} else {
			if errors.IsNotFound(err) {
				return fmt.Errorf("Namespace %v did not exist and namespace creation skipping was requested", cc.Namespace)
			}
			return err
		}
	} else if ShouldForceNamespaceCleanup() {
		if k8s.DeleteNamespaceAndWait(cc.VanClient.KubeClient, cc.Namespace) == nil {
			log.Printf("Removed existing namespace %v", cc.Namespace)
		}
	}
	_, err := kube.NewNamespace(cc.Namespace, cc.VanClient.KubeClient)
	if err == nil {
		cc.nsCreated = true
	}
	return err
}

func (cc *ClusterContext) DeleteNamespace() error {
	if !cc.nsCreated {
		log.Printf("namespace [%s] will not be deleted as it was not created by ClusterContext", cc.Namespace)
		return nil
	}
	if ShouldSkipNamespaceTeardown() {
		log.Print("Skipping namespace tear down, per env variables")
		return nil
	}
	if err := k8s.DeleteNamespaceAndWait(cc.VanClient.KubeClient, cc.Namespace); err != nil {
		return err
	}

	cc.nsCreated = false
	return nil
}

// As the name says, it will add label to this namespace
func (cc *ClusterContext) LabelNamespace(label string, value string) (err error) {

	payload := fmt.Sprintf(`{"metadata": {"labels": {"%v": "%v"}}}`, label, value)

	_, err = cc.VanClient.KubeClient.CoreV1().Namespaces().Patch(context.TODO(), cc.Namespace, types.MergePatchType, []byte(payload), metav1.PatchOptions{})

	return
}

func (cc *ClusterContext) waitForSkupperServiceToBeCreated(name string, retryFn func() (*apiv1.Service, error), backoff wait.Backoff) (*apiv1.Service, error) {
	var service *apiv1.Service = nil
	var err error
	isError := func(err error) bool {
		return err != nil
	}

	_retryFn := func() (*apiv1.Service, error) {
		cc.KubectlExec("get pods -o wide")
		return cc.VanClient.KubeClient.CoreV1().Services(cc.Namespace).Get(context.TODO(), name, metav1.GetOptions{})
	}

	if retryFn == nil {
		retryFn = _retryFn
	}

	return service, retry.OnError(backoff, isError, func() error {
		service, err = retryFn()
		return err
	})
}

func (cc *ClusterContext) DumpTestInfo(dirName string) {
	if !strings.HasPrefix(dirName, "tmp/") {
		dirName = fmt.Sprintf("tmp/%s", dirName)
	}
	f, err := os.Stat(dirName)
	if f != nil && !f.IsDir() {
		log.Printf("unable to dump test info: %s is not a directory", dirName)
		return
	} else if f == nil {
		if err := os.MkdirAll(dirName, 0755); err != nil {
			log.Printf("unable to dump test info: %v", err)
			return
		}
	}
	log.Printf("===> Dumping test information for: %s", cc.Namespace)
	tarBall, err := cc.VanClient.SkupperDump(context.Background(), fmt.Sprintf("%s/%s.tar.gz", dirName, cc.Namespace), cc.VanClient.GetVersion("service-controller", "service-controller"), "", "")
	if err == nil {
		absPath, _ := filepath.Abs(tarBall)
		log.Printf("Saved: %s", absPath)
	} else {
		log.Printf("error dumping test info: %v", err)
	}

	log.Printf("namespace status:")
	_, err = cc.KubectlExec("get -o wide job,pod,service,event")
	if err != nil {
		log.Printf("failed getting kube info: %v", err)
	}

	// These may be up and running now (and their output will show that).  However,
	// the last time they _terminated_, it was with a non-zero return, and that
	// may be valuable information for investigations.
	//
	// Here, we first get their pod and container names, and show some debugging
	// information...
	log.Printf("pod container whose last termination was non-zero:")
	out, err := cc.KubectlExec(`get pod -o go-template='
	    {{- range .items -}}
	      {{- with $pod := . -}}
		{{- range $pod.status.containerStatuses -}}
		  {{- with $cs := . -}}
		    {{- /* do we have a lastStae with exitCode on this cs? */ -}}
		    {{- if $cs.lastState.terminated.exitCode -}}
		      {{-  if ne $cs.lastState.terminated.exitCode 0 -}}
			{{ $pod.metadata.name }}{{ " " -}}
			{{ $cs.name }}{{ " " -}}
			lastExitCode: {{- $cs.lastState.terminated.exitCode }}{{ " " -}}
			lastReason: {{- $cs.lastState.terminated.reason }}{{ " " -}}
			lastStart: {{- $cs.lastState.terminated.startedAt }}{{ " " -}}
			lastFinish: {{- $cs.lastState.terminated.finishedAt }}{{ " " -}}
			restartCount: {{- $cs.restartCount }}{{ " " -}}
			started: {{- $cs.started }}{{ " " -}}
			podReady: {{- $cs.ready }}{{ "\n" -}}
		      {{- end  -}}
		    {{- end -}}
		  {{- end -}}
		{{- end -}}
	      {{- end -}}
	    {{- end -}}'
	`)
	if err != nil {
		log.Printf("failed gathering information on containers: %v", err)
	} else {

		lines := strings.Split(string(out), "\n")
		for _, line := range lines {

			if line == "" {
				continue
			}

			tokens := strings.Split(line, " ")

			if len(tokens) < 2 {
				fmt.Printf("The line %q is malformed for this process", line)
				continue
			}

			logCmd := fmt.Sprintf("logs %s -c %s -p --tail=2000 --timestamps", tokens[0], tokens[1])
			_, err = cc.KubectlExec(logCmd)
			if err != nil {
				log.Print("Failed fetching logs")
			}

		}
	}

	// On a healthy node, it will only report that it is ready.  If it faces any
	// pressures, though (disk, memory, pid), those will be listed
	log.Printf("node condition:")
	_, err = cc.KubectlExec(`get node -o jsonpath="{.items[*].status.conditions[?(@.status=='True')]}"`)
	if err != nil {
		log.Printf("failed gathering node condition: %v", err)
	}
}

// Returns a satisfied ClusterContextPromise, pointing to the ClusterContext where
// it was called.
// The Promise will be defective in that it will not have a reference to the
// ClusterTestRunnerBase (field cluster), so Runner() will always return nil.
func (cc *ClusterContext) GetPromise() *ClusterContextPromise {
	return &ClusterContextPromise{
		cluster:        nil,
		private:        cc.Private,
		id:             cc.Id,
		clusterContext: cc,
	}
}

// Some configuration may require a reference to a ClusterContext before it is
// created.  In that case, they can ask for a ClusterContextPromise, instead.
type ClusterContextPromise struct {
	cluster        *ClusterTestRunnerBase
	private        bool
	id             int
	clusterContext *ClusterContext
	dummy          bool // A dummy has only the reference to the clusterTestRunnerBase
}

// Satisfy the promise.  The returned error comes from the call to
// ClusterTestRunnerBase.GetContext()
// A successful run caches the ClusterContext for future calls
// (which will then never fail)
func (c *ClusterContextPromise) Satisfy() (*ClusterContext, error) {
	if c == nil {
		// to panic or not to panic?
		//
		// For development time, it's probably better to panic and get a proper
		// stack trace (TODO can it be added here?); on test time, it's better not
		// to panic, let the test fail and continue to the next test.
		return nil, fmt.Errorf("Satisfy() executed on a nil instance")
	}
	if c.dummy {
		return nil, fmt.Errorf("This is a dummy ClusterContextPromise")
	}
	var err error
	if c.clusterContext == nil {
		c.clusterContext, err = c.cluster.GetContext(c.private, c.id)
	}
	return c.clusterContext, err
}

// Returns the stored, private reference to the ClusterTestRunnerBase
func (c *ClusterContextPromise) Runner() *ClusterTestRunnerBase {
	return c.cluster
}
