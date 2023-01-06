package validate

import (
	"fmt"
	"log"

	"github.com/skupperproject/skupper/test/utils/base"
	"github.com/skupperproject/skupper/test/utils/tools"
)

// Provides an interface to tools.Curl, with some enhancements.
//
// If CurlOptions.Timeout is zero, a default is set, instead.
type Curl struct {
	Namespace *base.ClusterContextPromise

	// CurlOptions is passed as-is to tools.Curl, with the exception that a
	// default of 60s is set for the timeout, if the original value is
	// zero.
	CurlOptions tools.CurlOpts
	Url         string
	Fail400Plus bool
	Podname     string // Passed to tools.Curl.  Generally safe to leave empty.  Check tools.Curl docs
	DeployCurl  bool   // Not Implemented

	// TODO: Add cli.Expect to inspect results?
}

func (c Curl) Validate() error {
	if c.DeployCurl {
		return fmt.Errorf("validate.Curl.DeployCurl not implemented yet")
	}
	if c.CurlOptions.Timeout == 0 {
		// There is no reason to give Curl no time to respond
		c.CurlOptions.Timeout = 60
	}
	cluster, err := c.Namespace.Satisfy()
	if err != nil {
		return err
	}
	log.Printf("Calling Curl on %v", c.Url)
	resp, err := tools.Curl(
		cluster.VanClient.KubeClient,
		cluster.VanClient.RestConfig,
		cluster.Namespace,
		c.Podname,
		c.Url,
		c.CurlOptions,
	)
	log.Printf("- Output:\n%v", resp.Output)
	if err != nil {
		log.Printf("- Err: %v", err)
		return fmt.Errorf("curl invokation failed: %w", err)
	}

	log.Printf("- status code %d", resp.StatusCode)
	log.Printf("- HTTP version: %v", resp.HttpVersion)
	log.Printf("- Reason phrase: %v", resp.ReasonPhrase)
	log.Printf("- Headers:\n%v", resp.Headers)
	log.Printf("- Body:\n%v", resp.Body)

	if c.Fail400Plus && resp.StatusCode >= 400 {
		return fmt.Errorf("curl invokation returned status code %d", resp.StatusCode)
	}

	return err
}
