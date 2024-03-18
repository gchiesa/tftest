package scenario

import (
	"path/filepath"
	"testing"

	"github.com/excoriate/tftest/pkg/cloudprovider"

	"github.com/excoriate/tftest/pkg/validation"

	"github.com/gruntwork-io/terratest/modules/terraform"
	test_structure "github.com/gruntwork-io/terratest/modules/test-structure"
)

type Options struct {
	// vars is the Terraform variables
	vars map[string]interface{}
	// varFiles is the Terraform variable files
	varFiles []string
	// enableAWS is a flag to enable the AWS Cloud Provider (Client)
	enableAWS bool
	// awsRegion is the AWS region. If not set, it defaults to "us-west-2"
	awsRegion string
}

type OptFn func(*Options) error

type Client struct {
	// t is the testing instance
	t *testing.T
	// opts is the Terraform Options
	opts *terraform.Options
	// Stg is the StageClient
	Stg *StageClient
	// awsCfg is the AWS Cloud Provider (Client)
	awsCloud cloudprovider.AWSAdapter
}

type Config interface {
	GetTerraformOptions() *terraform.Options
	GetAWS() cloudprovider.AWSAdapter
}

func (c *Client) GetTerraformOptions() *terraform.Options {
	if c.opts == nil {
		return &terraform.Options{}
	}

	return c.opts
}

func (c *Client) GetAWS() cloudprovider.AWSAdapter {
	return c.awsCloud
}

func WithVars(vars map[string]interface{}) OptFn {
	return func(o *Options) error {
		o.vars = vars
		return nil
	}
}

func WithAWS(region string) OptFn {
	return func(o *Options) error {
		if region == "" {
			region = "us-west-2"
		}

		o.enableAWS = true
		o.awsRegion = region

		return nil
	}
}

func WithVarFiles(workdir string, varFiles ...string) OptFn {
	return func(o *Options) error {
		if err := validation.IsValidTFDir(workdir); err != nil {
			return err
		}

		for _, vf := range varFiles {
			if err := validation.IsValidTFVarFile(filepath.Join(workdir, vf)); err != nil {
				return err
			}
		}

		o.varFiles = varFiles
		return nil
	}
}

func NewWithOptions(t *testing.T, workdir string, opts ...OptFn) (*Client, error) {
	o := &Options{}
	for _, opt := range opts {
		if err := opt(o); err != nil {
			return nil, err
		}
	}

	if err := validation.IsValidTFDir(workdir); err != nil {
		return nil, err
	}

	c := &Client{}

	tfOptions := &terraform.Options{
		TerraformDir: workdir,
		PlanFilePath: DefaultPlanOutput,
		NoColor:      true,
	}

	if o.enableAWS {
		cfg, err := cloudprovider.NewAWS(o.awsRegion)
		t.Logf("Enabling AWS Cloud Provider (Client) with region: %s", o.awsRegion)
		if err != nil {
			return nil, err
		}

		c.awsCloud = cfg
	}

	if len(o.vars) > 0 {
		t.Logf("Setting Terraform variables: %v", o.vars)
		tfOptions.Vars = o.vars
	}

	if len(o.varFiles) > 0 {
		t.Logf("Setting Terraform variable files: %v", o.varFiles)
		tfOptions.VarFiles = o.varFiles
	}

	c.opts = tfOptions
	test_structure.SaveTerraformOptions(t, workdir, c.opts)

	return c, nil
}

// New creates a new Terraform options with default retryable errors and saves it to the workdir
// This is a wrapper around terraform.WithDefaultRetryableErrors
func New(t *testing.T, workdir string) (*Client, error) {
	if err := validation.IsValidTFDir(workdir); err != nil {
		return nil, err
	}

	terraformOptions := terraform.WithDefaultRetryableErrors(t, &terraform.Options{
		TerraformDir: workdir,
		NoColor:      true,
		PlanFilePath: "plan.out",
	})

	test_structure.SaveTerraformOptions(t, workdir, terraformOptions)

	return &Client{
		t:    t,
		opts: terraformOptions,
		Stg:  &StageClient{},
	}, nil
}