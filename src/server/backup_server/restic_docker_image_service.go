package backup_server

import (
	"server/configs"
	"server/tools"

	u "github.com/quollix/common/utils"
)

type ResticDockerImageService interface {
	UpdateResticDockerImage() error
}

type ResticDockerImageServiceImpl struct {
	AuthHelper              u.AuthHelper
	ConfigsRepository       configs.ConfigsRepository
	DockerService           tools.DockerService
	DirectoryProvider       tools.DirectoryProvider
	OsWrapper               u.OsWrapper
	ResticContainerExecutor tools.ResticContainerExecutor
}

func (r *ResticDockerImageServiceImpl) UpdateResticDockerImage() error {
	dockerfilePath := r.DirectoryProvider.GetResticImageDockerfilePath()
	dockerfileBytes, err := r.OsWrapper.ReadFile(dockerfilePath)
	if err != nil {
		return err
	}

	dockerfileHash := r.AuthHelper.GetSHA256Hash(string(dockerfileBytes))
	doesExist, err := r.DockerService.DoesDockerImageExist(tools.ResticImageName)
	if err != nil {
		return err
	}

	isStoredHashSet, err := r.ConfigsRepository.IsConfigSet(configs.ConfigKeys.ResticDockerfileHash)
	if err != nil {
		return err
	}
	if isStoredHashSet {
		storedHash, err := r.ConfigsRepository.GetConfig(configs.ConfigKeys.ResticDockerfileHash)
		if err != nil {
			return err
		}
		if storedHash == dockerfileHash && doesExist {
			u.Logger.Info("restic image is up to date, skipping creation")
			return nil
		}
	}

	u.Logger.Info("restic image is outdated or missing, creating it, this may take a while")
	if err := r.DockerService.BuildDockerImage(tools.ResticImageName, dockerfilePath, true); err != nil {
		return err
	}

	if err := r.ConfigsRepository.SetConfig(configs.ConfigKeys.ResticDockerfileHash, dockerfileHash); err != nil {
		return err
	}

	u.Logger.Info("restic image was created successfully")
	return r.prepareResticContainerDirectories()
}

func (r *ResticDockerImageServiceImpl) prepareResticContainerDirectories() error {
	_, err := r.ResticContainerExecutor.ExecuteSimple("mkdir -p /root/.config/rclone")
	if err != nil {
		return err
	}
	_, err = r.ResticContainerExecutor.ExecuteSimple("mkdir -p /root/.ssh")
	if err != nil {
		return err
	}
	return nil
}
