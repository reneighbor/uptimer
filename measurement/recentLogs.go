package measurement

import (
	"bytes"
	"fmt"
	"log"

	"github.com/cloudfoundry/uptimer/appLogValidator"
	"github.com/cloudfoundry/uptimer/cmdRunner"
	"github.com/cloudfoundry/uptimer/cmdStartWaiter"
)

type recentLogs struct {
	name                           string
	summaryPhrase                  string
	recentLogsCommandGeneratorFunc func() []cmdStartWaiter.CmdStartWaiter
	runner                         cmdRunner.CmdRunner
	runnerOutBuf                   *bytes.Buffer
	runnerErrBuf                   *bytes.Buffer
	appLogValidator                appLogValidator.AppLogValidator
}

func (r *recentLogs) Name() string {
	return r.name
}

func (r *recentLogs) SummaryPhrase() string {
	return r.summaryPhrase
}

func (r *recentLogs) PerformMeasurement(logger *log.Logger) bool {
	defer r.runnerOutBuf.Reset()
	defer r.runnerErrBuf.Reset()

	if err := r.runner.RunInSequence(r.recentLogsCommandGeneratorFunc()...); err != nil {
		r.logFailure(logger, err.Error(), r.runnerOutBuf.String(), r.runnerErrBuf.String())
		return false
	}

	logIsNewer, err := r.appLogValidator.IsNewer(r.runnerOutBuf.String())
	if err != nil {
		r.logFailure(logger, fmt.Sprintf("App log validation failed with: %s", err.Error()), r.runnerOutBuf.String(), r.runnerErrBuf.String())
		return false
	} else if !logIsNewer {
		r.logFailure(logger, "App log fetched was not newer than previous app log fetched", r.runnerOutBuf.String(), r.runnerErrBuf.String())
		return false
	}

	return true
}

func (r *recentLogs) logFailure(logger *log.Logger, errString, cmdOut, cmdErr string) {
	logger.Printf(
		"\x1b[31mFAILURE(%s): %s\x1b[0m\nstdout:\n%s\nstderr:\n%s\n\n",
		r.name,
		errString,
		cmdOut,
		cmdErr,
	)
}

func (r *recentLogs) Failed(rs ResultSet) bool {
	return rs.Failed() > 0
}
