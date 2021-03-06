package hcl2template

import (
	"fmt"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/packer/packer"
)

// ProvisionerBlock references a detected but unparsed provisioner
type ProvisionerBlock struct {
	PType       string
	PName       string
	PauseBefore time.Duration
	MaxRetries  int
	Timeout     time.Duration
	HCL2Ref
}

func (p *ProvisionerBlock) String() string {
	return fmt.Sprintf(buildProvisionerLabel+"-block %q %q", p.PType, p.PName)
}

func (p *Parser) decodeProvisioner(block *hcl.Block) (*ProvisionerBlock, hcl.Diagnostics) {
	var b struct {
		Name        string   `hcl:"name,optional"`
		PauseBefore string   `hcl:"pause_before,optional"`
		MaxRetries  int      `hcl:"max_retries,optional"`
		Timeout     string   `hcl:"timeout,optional"`
		Rest        hcl.Body `hcl:",remain"`
	}
	diags := gohcl.DecodeBody(block.Body, nil, &b)
	if diags.HasErrors() {
		return nil, diags
	}

	provisioner := &ProvisionerBlock{
		PType:      block.Labels[0],
		PName:      b.Name,
		MaxRetries: b.MaxRetries,
		HCL2Ref:    newHCL2Ref(block, b.Rest),
	}

	if b.PauseBefore != "" {
		pauseBefore, err := time.ParseDuration(b.PauseBefore)
		if err != nil {
			return nil, append(diags, &hcl.Diagnostic{
				Summary: "Failed to parse pause_before duration",
				Detail:  err.Error(),
			})
		}
		provisioner.PauseBefore = pauseBefore
	}

	if b.Timeout != "" {
		timeout, err := time.ParseDuration(b.Timeout)
		if err != nil {
			return nil, append(diags, &hcl.Diagnostic{
				Summary: "Failed to parse timeout duration",
				Detail:  err.Error(),
			})
		}
		provisioner.Timeout = timeout
	}

	if !p.ProvisionersSchemas.Has(provisioner.PType) {
		diags = append(diags, &hcl.Diagnostic{
			Summary:  fmt.Sprintf("Unknown "+buildProvisionerLabel+" type %q", provisioner.PType),
			Subject:  block.LabelRanges[0].Ptr(),
			Detail:   fmt.Sprintf("known "+buildProvisionerLabel+"s: %v", p.ProvisionersSchemas.List()),
			Severity: hcl.DiagError,
		})
		return nil, diags
	}
	return provisioner, diags
}

func (cfg *PackerConfig) startProvisioner(source *SourceBlock, pb *ProvisionerBlock, ectx *hcl.EvalContext, generatedVars map[string]string) (packer.Provisioner, hcl.Diagnostics) {
	var diags hcl.Diagnostics

	provisioner, err := cfg.provisionersSchemas.Start(pb.PType)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Summary: fmt.Sprintf("failed loading %s", pb.PType),
			Subject: pb.HCL2Ref.LabelsRanges[0].Ptr(),
			Detail:  err.Error(),
		})
		return nil, diags
	}
	flatProvisionerCfg, moreDiags := decodeHCL2Spec(pb.HCL2Ref.Rest, ectx, provisioner)
	diags = append(diags, moreDiags...)
	if diags.HasErrors() {
		return nil, diags
	}
	// manipulate generatedVars from builder to add to the interfaces being
	// passed to the provisioner Prepare()

	// configs := make([]interface{}, 2)
	// configs = append(, flatProvisionerCfg)
	// configs = append(configs, generatedVars)
	err = provisioner.Prepare(source.builderVariables(), flatProvisionerCfg, generatedVars)
	if err != nil {
		diags = append(diags, &hcl.Diagnostic{
			Severity: hcl.DiagError,
			Summary:  fmt.Sprintf("Failed preparing %s", pb),
			Detail:   err.Error(),
			Subject:  pb.HCL2Ref.DefRange.Ptr(),
		})
		return nil, diags
	}
	return provisioner, diags
}
