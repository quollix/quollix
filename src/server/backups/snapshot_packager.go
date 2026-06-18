package backups

import (
	"fmt"
	"os"
	"path/filepath"
	"server/apps_basic"
	"server/backup_server"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type AppSnapshot struct {
	DockerComposeYamlContent []byte
	Volumes                  []string
	Meta                     *MetaData

	ContainerNames []string
}

type PreparedSnapshot struct {
	TempDir            string
	MountArgsForRestic string
	Volumes            []string
}

type SnapshotPackager interface {
	PrepareBackupFiles(dto BackupCreationDto, meta *MetaData) (*PreparedSnapshot, error)
	ReadSnapshotFromDir(dir string) (*AppSnapshot, error)
}

type SnapshotPackagerImpl struct {
	MetaCodec       MetaCodecImpl
	VolumeExtractor apps_basic.ComposeExtractor
	OsWrapper       u.OsWrapper
}

func (p *SnapshotPackagerImpl) PrepareBackupFiles(dto BackupCreationDto, meta *MetaData) (*PreparedSnapshot, error) {
	tempDir, err := os.MkdirTemp(tools.TempDir, "temp")
	if err != nil {
		return nil, err
	}

	dockerComposePath := filepath.Join(tempDir, backup_server.DockerComposeYaml)
	if err = os.WriteFile(dockerComposePath, dto.VersionContent, 0o600); err != nil {
		u.RemoveDir(tempDir)
		return nil, err
	}

	metaPath := filepath.Join(tempDir, backup_server.MetaYmlFileName)
	if err = p.MetaCodec.Save(metaPath, meta); err != nil {
		u.RemoveDir(tempDir)
		return nil, err
	}

	volumes, _, err := p.VolumeExtractor.Extract(dto.VersionContent)
	if err != nil {
		u.RemoveDir(tempDir)
		return nil, err
	}
	if len(volumes) == 0 {
		u.RemoveDir(tempDir)
		return nil, u.Logger.NewError(backup_server.CantBackupAppWithoutVolumes)
	}

	mountArgs := fmt.Sprintf("-v %s:/source/%s -v %s:/source/%s ", dockerComposePath, backup_server.DockerComposeYaml, metaPath, backup_server.MetaYmlFileName)
	snapshot := &PreparedSnapshot{
		TempDir:            tempDir,
		MountArgsForRestic: mountArgs,
		Volumes:            volumes,
	}
	return snapshot, nil
}

func (p *SnapshotPackagerImpl) ReadSnapshotFromDir(dir string) (*AppSnapshot, error) {
	metaPath := filepath.Join(dir, backup_server.MetaYmlFileName)
	if _, err := os.Stat(metaPath); err != nil { // #nosec G703: dir is the trusted extracted snapshot directory selected by application workflow
		return nil, err
	}

	meta, err := p.MetaCodec.Load(metaPath)
	if err != nil {
		return nil, err
	}

	var dockerComposePath = filepath.Join(dir, backup_server.DockerComposeYaml)
	content, err := p.OsWrapper.ReadFile(dockerComposePath) // #nosec G304 G703: dockerComposePath is a fixed file inside the trusted snapshot directory
	if err != nil {
		return nil, err
	}

	volumes, containerNames, err := p.VolumeExtractor.Extract(content)
	if err != nil {
		return nil, err
	}

	return &AppSnapshot{
		DockerComposeYamlContent: content,
		Volumes:                  volumes,
		Meta:                     meta,
		ContainerNames:           containerNames,
	}, nil
}
