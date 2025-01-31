package action

import (
	"fmt"
	"io"
	"strings"

	"github.com/iter8-tools/iter8/action/report"
	"github.com/iter8-tools/iter8/base"
	"github.com/iter8-tools/iter8/base/log"
	"github.com/iter8-tools/iter8/driver"
)

const (
	// TextOutputFormatKey is the output format used to create text output
	TextOutputFormatKey = "text"

	// HTMLOutputFormatKey is the output format used to create html output
	HTMLOutputFormatKey = "html"
)

// ReportOpts are the options used for generating reports from experiment result
type ReportOpts struct {
	// OutputFormat specifies the output format to be used by report
	OutputFormat string
	// RunOpts enables fetching local experiment spec and result
	RunOpts
	// KubeDriver enables fetching Kubernetes experiment spec and result
	*driver.KubeDriver
}

// NewReportOpts initializes and returns report opts
func NewReportOpts(kd *driver.KubeDriver) *ReportOpts {
	return &ReportOpts{
		RunOpts: RunOpts{
			RunDir: ".",
		},
		OutputFormat: TextOutputFormatKey,
		KubeDriver:   kd,
	}
}

// KubeRun generates report for a Kubernetes experiment
func (rOpts *ReportOpts) KubeRun(out io.Writer) error {
	if err := rOpts.KubeDriver.Init(); err != nil {
		return err
	}
	return rOpts.Run(rOpts, out)
}

// Run generates the text or HTML report
func (rOpts *ReportOpts) Run(eio base.Driver, out io.Writer) error {
	var e *base.Experiment
	var err error
	if e, err = base.BuildExperiment(eio); err != nil {
		return err
	}
	switch strings.ToLower(rOpts.OutputFormat) {
	case TextOutputFormatKey:
		reporter := report.TextReporter{
			Reporter: &report.Reporter{
				Experiment: e,
			},
		}
		return reporter.Gen(out)
	case HTMLOutputFormatKey:
		reporter := report.HTMLReporter{
			Reporter: &report.Reporter{
				Experiment: e,
			},
		}
		return reporter.Gen(out)
	default:
		e := fmt.Errorf("unsupported report format %v", rOpts.OutputFormat)
		log.Logger.Error(e)
		return e
	}
}
