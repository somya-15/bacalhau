package computenode

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/filecoin-project/bacalhau/pkg/computenode"
	_ "github.com/filecoin-project/bacalhau/pkg/logger"
	"github.com/stretchr/testify/assert"
)

// test that when we have RejectStatelessJobs turned on
// we don't accept a job with no volumes
// but when it's not turned on the job is actually selected
func TestJobSelectionNoVolumes(t *testing.T) {
	runTest := func(rejectSetting, expectedResult bool) {
		computeNode, _, cm := SetupTest(t, computenode.ComputeNodeConfig{
			JobSelectionPolicy: computenode.JobSelectionPolicy{
				RejectStatelessJobs: rejectSetting,
			},
		})
		defer cm.Cleanup()

		result, err := computeNode.SelectJob(context.Background(), GetProbeData(""))
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)
	}

	runTest(true, false)
	runTest(false, true)
}

func TestJobSelectionLocality(t *testing.T) {

	// get the CID so we can use it in the tests below but without it actually being
	// added to the server (so we can test locality anywhere)
	EXAMPLE_TEXT := "hello"
	cid, err := (func() (string, error) {
		_, ipfsStack, cm := SetupTest(t, computenode.NewDefaultComputeNodeConfig())
		defer cm.Cleanup()
		return ipfsStack.AddTextToNodes(1, []byte(EXAMPLE_TEXT))
	}())
	assert.NoError(t, err)

	runTest := func(locality computenode.JobSelectionDataLocality, shouldAddData, expectedResult bool) {

		computeNode, ipfsStack, cm := SetupTest(t, computenode.ComputeNodeConfig{
			JobSelectionPolicy: computenode.JobSelectionPolicy{
				Locality: locality,
			},
		})
		defer cm.Cleanup()

		if shouldAddData {
			_, err := ipfsStack.AddTextToNodes(1, []byte(EXAMPLE_TEXT))
			assert.NoError(t, err)
		}

		result, err := computeNode.SelectJob(context.Background(), GetProbeData(cid))
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)
	}

	// we are local - we do have the file - we should accept
	runTest(computenode.Local, true, true)

	// we are local - we don't have the file - we should reject
	runTest(computenode.Local, false, false)

	// we are anywhere - we do have the file - we should accept
	runTest(computenode.Anywhere, true, true)

	// we are anywhere - we don't have the file - we should accept
	runTest(computenode.Anywhere, false, true)
}

func TestJobSelectionHttp(t *testing.T) {
	runTest := func(failMode, expectedResult bool) {
		svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.Method, "POST")
			if failMode {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("500 - Something bad happened!"))
			} else {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("200 - Everything is good!"))
			}

		}))
		defer svr.Close()

		computeNode, _, cm := SetupTest(t, computenode.ComputeNodeConfig{
			JobSelectionPolicy: computenode.JobSelectionPolicy{
				ProbeHTTP: svr.URL,
			},
		})
		defer cm.Cleanup()

		result, err := computeNode.SelectJob(context.Background(), GetProbeData(""))
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)
	}

	runTest(true, false)
	runTest(false, true)
}

func TestJobSelectionExec(t *testing.T) {
	runTest := func(failMode, expectedResult bool) {
		command := "exit 0"
		if failMode {
			command = "exit 1"
		}
		computeNode, _, cm := SetupTest(t, computenode.ComputeNodeConfig{
			JobSelectionPolicy: computenode.JobSelectionPolicy{
				ProbeExec: command,
			},
		})
		defer cm.Cleanup()

		result, err := computeNode.SelectJob(context.Background(), GetProbeData(""))
		assert.NoError(t, err)
		assert.Equal(t, result, expectedResult)
	}

	runTest(true, false)
	runTest(false, true)
}
