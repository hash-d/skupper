package validate

import (
	"fmt"

	"github.com/skupperproject/skupper/test/frame2"
	"github.com/skupperproject/skupper/test/utils/tools"
)

type Curl struct {
	frame2.Validate
	CurlOptions tools.CurlOpts
	Url         string
	Fail400Plus bool
	Podname     string // Passed to tools.Curl.  Generally safe to leave empty.  Check tools.Curl docs
	DeployCurl  bool   // Not Implemented
}

func (c Curl) Run() error {
	if c.DeployCurl {
		return fmt.Errorf("validate.Curl.DeployCurl not implemented yet")
	}
	cluster, err := c.Namespace.Satisfy()
	if err != nil {
		return err
	}
	c.Logf("Calling Curl on %v", c.Url)
	resp, err := tools.Curl(
		cluster.VanClient.KubeClient,
		cluster.VanClient.RestConfig,
		cluster.Namespace,
		c.Podname,
		c.Url,
		c.CurlOptions,
	)
	c.Logf("- Output:\n%v", resp.Output)
	if err != nil {
		return fmt.Errorf("curl invokation failed: %w", err)
	}

	c.Logf("- status code %d", resp.StatusCode)
	c.Logf("- HTTP version: %v", resp.HttpVersion)
	c.Logf("- Reason phrase: %v", resp.ReasonPhrase)
	c.Logf("- Headers:\n%v", resp.Headers)
	c.Logf("- Body:\n", resp.Body)

	if c.Fail400Plus && resp.StatusCode >= 400 {
		return fmt.Errorf("curl invokation returned status code %d", resp.StatusCode)
	}

	return err
}
