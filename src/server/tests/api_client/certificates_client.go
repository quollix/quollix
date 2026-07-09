package api_client

import (
	"encoding/json"
	"server/certificates"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type QuollixCertificatesClient struct {
	quollix *QuollixClient
}

func (c *QuollixCertificatesClient) Reset() error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateReset, nil)
	return err
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

func (c *QuollixCertificatesClient) Upload(content []byte) error {
	_, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateUpload, tools.BinaryFile{
		FileName: "certificate.pem",
		Content:  content,
	})
	return err
}

func (c *QuollixCertificatesClient) DownloadBundleBytes() ([]byte, error) {
	body, err := c.quollix.Parent.DoRequest(tools.Paths.BackendSettingsCertificateDownload, nil)
	if err != nil {
		return nil, err
	}

	var certificateFile tools.BinaryFile
	err = json.Unmarshal(body, &certificateFile)
	if err != nil {
		return nil, err
	}
	if certificateFile.FileName != "certificate.pem" {
		return nil, u.Logger.NewError("unexpected certificate file name", "file_name", certificateFile.FileName)
	}
	return certificateFile.Content, nil
}
