package component

import (
	"encoding/json"
	"server/certificates"
	"server/tools"

	"github.com/quollix/common/assert"
)

type QuollixCertificatesClient struct {
	quollix *QuollixClient
}

func (c *QuollixCertificatesClient) Reset() {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateReset, nil)
	assert.Nil(c.quollix.T, err)
}

func (c *QuollixCertificatesClient) TryDns01Challenge() (*certificates.Dns01ChallengeInfo, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsStartDns01CertificateChallenge, nil)
	if err != nil {
		return nil, err
	}
	var info certificates.Dns01ChallengeInfo
	err = json.Unmarshal(body, &info)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

func (c *QuollixCertificatesClient) Upload(content []byte) {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateUpload, tools.BinaryFile{
		FileName: "certificate.pem",
		Content:  content,
	})
	assert.Nil(c.quollix.T, err)
}

func (c *QuollixCertificatesClient) DownloadBundleBytes() []byte {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateDownload, nil)
	assert.Nil(c.quollix.T, err)

	var certificateFile tools.BinaryFile
	err = json.Unmarshal(body, &certificateFile)
	assert.Nil(c.quollix.T, err)
	assert.Equal(c.quollix.T, "certificate.pem", certificateFile.FileName)
	assertBundleParsesAsTlsKeyPair(c.quollix.T, certificateFile.Content)
	return certificateFile.Content
}
