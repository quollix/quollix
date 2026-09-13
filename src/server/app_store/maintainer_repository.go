package app_store

import (
	"crypto/ed25519"
	"database/sql"
	"strings"
	"time"

	"server/tools"

	u "github.com/quollix/common/utils"
	"golang.org/x/crypto/ssh"
)

type MaintainerRepository interface {
	GetPublicKey(maintainer string) (ed25519.PublicKey, bool, error)
	ListPublicKeys() ([]MaintainerPublicKey, error)
	UpsertPublicKey(maintainer string, publicKeyRaw ed25519.PublicKey) error
	DeletePublicKey(maintainer string) error
}

type MaintainerPublicKey struct {
	Name                 string
	PublicKeyRaw         ed25519.PublicKey
	PublicKeyFingerprint string
	LastUpdatedAt        time.Time
}

type MaintainerRepositoryImpl struct {
	DbProvider tools.DatabaseConnector
}

func (r *MaintainerRepositoryImpl) GetPublicKey(maintainer string) (ed25519.PublicKey, bool, error) {
	var publicKeyRaw []byte
	err := r.DbProvider.GetDB().QueryRow(`
SELECT public_key_raw
FROM app_maintainers
WHERE name = $1
`, maintainer).Scan(&publicKeyRaw)
	if err == sql.ErrNoRows {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, u.Logger.NewError(err.Error(), tools.MaintainerField, maintainer)
	}
	publicKey := ed25519.PublicKey(publicKeyRaw)
	if _, err = maintainerPublicKeyFingerprint(publicKey); err != nil {
		return nil, false, err
	}
	return publicKey, true, nil
}

func (r *MaintainerRepositoryImpl) ListPublicKeys() ([]MaintainerPublicKey, error) {
	rows, err := r.DbProvider.GetDB().Query(`
SELECT name, public_key_raw, public_key_fingerprint, last_updated_at
FROM app_maintainers
ORDER BY name
`)
	if err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	defer u.Close(rows)

	result := make([]MaintainerPublicKey, 0)
	for rows.Next() {
		var maintainerKey MaintainerPublicKey
		var publicKeyRaw []byte
		err = rows.Scan(
			&maintainerKey.Name,
			&publicKeyRaw,
			&maintainerKey.PublicKeyFingerprint,
			&maintainerKey.LastUpdatedAt,
		)
		if err != nil {
			return nil, u.Logger.NewError(err.Error())
		}
		maintainerKey.PublicKeyRaw = ed25519.PublicKey(publicKeyRaw)
		if _, err = maintainerPublicKeyFingerprint(maintainerKey.PublicKeyRaw); err != nil {
			return nil, err
		}
		result = append(result, maintainerKey)
	}
	if err = rows.Err(); err != nil {
		return nil, u.Logger.NewError(err.Error())
	}
	return result, nil
}

func (r *MaintainerRepositoryImpl) UpsertPublicKey(maintainer string, publicKeyRaw ed25519.PublicKey) error {
	fingerprint, err := maintainerPublicKeyFingerprint(publicKeyRaw)
	if err != nil {
		return err
	}
	_, err = r.DbProvider.GetDB().Exec(`
INSERT INTO app_maintainers (
	name,
	public_key_raw,
	public_key_fingerprint,
	last_updated_at
)
VALUES ($1, $2, $3, NOW())
ON CONFLICT (name) DO UPDATE SET
	public_key_raw = EXCLUDED.public_key_raw,
	public_key_fingerprint = EXCLUDED.public_key_fingerprint,
	last_updated_at = NOW()
`, maintainer, []byte(publicKeyRaw), fingerprint)
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.MaintainerField, maintainer)
	}
	return nil
}

func (r *MaintainerRepositoryImpl) DeletePublicKey(maintainer string) error {
	_, err := r.DbProvider.GetDB().Exec(`
DELETE FROM app_maintainers
WHERE name = $1
`, maintainer)
	if err != nil {
		return u.Logger.NewError(err.Error(), tools.MaintainerField, maintainer)
	}
	return nil
}

// only used during testing
func (r *MaintainerRepositoryImpl) Wipe() {
	_, err := r.DbProvider.GetDB().Exec("DELETE FROM app_maintainers")
	if err != nil {
		u.Logger.Error(err)
	}
}

func maintainerPublicKeyFingerprint(publicKeyRaw ed25519.PublicKey) (string, error) {
	if len(publicKeyRaw) != ed25519.PublicKeySize {
		return "", u.Logger.NewError("invalid maintainer public key size")
	}
	sshPublicKey, err := ssh.NewPublicKey(publicKeyRaw)
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return ssh.FingerprintSHA256(sshPublicKey), nil
}

func MaintainerPublicKeyOpenSSH(publicKeyRaw ed25519.PublicKey) (string, error) {
	if len(publicKeyRaw) != ed25519.PublicKeySize {
		return "", u.Logger.NewError("invalid maintainer public key size")
	}
	sshPublicKey, err := ssh.NewPublicKey(publicKeyRaw)
	if err != nil {
		return "", u.Logger.NewError(err.Error())
	}
	return strings.TrimSpace(string(ssh.MarshalAuthorizedKey(sshPublicKey))), nil
}
