//go:build integration || !unit

package devstack

import (
	"testing"

	"github.com/bacalhau-project/bacalhau/pkg/devstack"
	"github.com/bacalhau-project/bacalhau/pkg/job"
	"github.com/bacalhau-project/bacalhau/pkg/model"
	"github.com/bacalhau-project/bacalhau/pkg/test/scenario"
	"github.com/stretchr/testify/suite"
)

type DisabledFeatureTestSuite struct {
	scenario.ScenarioRunner
}

func TestDisabledFeatureSuite(t *testing.T) {
	suite.Run(t, new(DisabledFeatureTestSuite))
}

var disabledTestSpec = scenario.Scenario{
	Stack: &scenario.StackConfig{
		DevStackOptions: &devstack.DevStackOptions{
			NumberOfRequesterOnlyNodes: 1,
			NumberOfComputeOnlyNodes:   1,
		},
	},
	Spec: scenario.WasmHelloWorld.Spec,
	JobCheckers: []job.CheckStatesFunction{
		func(js model.JobState) (bool, error) {
			return js.State == model.JobStateError, nil
		},
	},
}

func (s *DisabledFeatureTestSuite) TestDisabledEngine() {
	testCase := disabledTestSpec
	testCase.Stack.DevStackOptions.DisabledFeatures.Engines = []model.Engine{model.EngineWasm}

	s.RunScenario(testCase)
}

func (s *DisabledFeatureTestSuite) TestDisabledStorage() {
	testCase := disabledTestSpec
	testCase.Stack.DevStackOptions.DisabledFeatures.Storages = []model.StorageSourceType{model.StorageSourceInline}

	s.RunScenario(testCase)
}

func (s *DisabledFeatureTestSuite) TestDisabledVerifier() {
	testCase := disabledTestSpec
	testCase.Spec.Verifier = model.VerifierNoop
	testCase.Stack.DevStackOptions.DisabledFeatures.Verifiers = []model.Verifier{model.VerifierNoop}

	s.RunScenario(testCase)
}

func (s *DisabledFeatureTestSuite) TestDisabledPublisher() {
	testCase := disabledTestSpec
	testCase.Spec.Publisher = model.PublisherNoop
	testCase.Stack.DevStackOptions.DisabledFeatures.Publishers = []model.Publisher{model.PublisherNoop}

	s.RunScenario(testCase)
}
