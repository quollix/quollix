package certificates

import (
	"fmt"
	"os"
	"path/filepath"
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type CertificatePersister interface {
	LoadCertificateFromHostSystemToDatabaseIfExist() error
}

type CertificatePersisterImpl struct {
	Config            *tools.GlobalConfig
	ConfigsRepository configs.ConfigsRepository
	DirectoryProvider tools.DirectoryProvider
}

func (c *CertificatePersisterImpl) LoadCertificateFromHostSystemToDatabaseIfExist() error {
	certificatesDirectoryInfo, err := os.Stat(c.DirectoryProvider.GetCacheDir())
	if err != nil {
		u.Logger.Warn("tried to find the certificate which should only be looked up in the development profile, but failed unexpectedly", "error", err.Error())
		return nil
	}

	if !certificatesDirectoryInfo.IsDir() {
		return fmt.Errorf("%s is not a directory", c.DirectoryProvider.GetCacheDir())
	}

	certificatePemPath := filepath.Join(c.DirectoryProvider.GetCacheDir(), "certificate.pem")
	pemBundleBytes, err := os.ReadFile(certificatePemPath) // #nosec G304 (CWE-22): Potential file inclusion via variable
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	certBundle, err := NewCertificateBundleFromPemBytes(pemBundleBytes)
	if err != nil {
		return err
	}

	return c.ConfigsRepository.SetConfig(configs.ConfigKeys.CertificatePemBundle, certBundle.GetString())
}
