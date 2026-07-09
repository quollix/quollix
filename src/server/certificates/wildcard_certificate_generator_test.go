package certificates

import (
	"testing"

	"github.com/quollix/common/assert"
)

func TestGetDnsRecordStub(t *testing.T) {
	generatorStub := &WildcardCertificateServiceMock{}
	_, info, err := generatorStub.StartDns01Session()
	assert.Nil(t, err)
	assert.Equal(t, acmeChallengePrefix+"localhost", info.RecordName)
	assert.Equal(t, SampleWildcardKeyAuth, info.WildcardKeyAuth)
}
