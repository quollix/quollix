package backup_server

import (
	"server/configs"
	"server/tools"
	"testing"

	"github.com/quollix/common/assert"
)

var (
	resticDockerfilePath    = "/quollix/restic/Dockerfile"
	resticDockerfileBytes   = []byte("FROM alpine:3.24\n")
	resticDockerfileHash    = "new-hash"
	oldResticDockerfileHash = "old-hash"
)

type resticDockerImageServiceTestObjects struct {
	Service                 *ResticDockerImageServiceImpl
	AuthHelper              *tools.AuthHelperMock
	ConfigsRepository       *configs.ConfigsRepositoryMock
	DockerService           *tools.DockerServiceMock
	DirectoryProvider       *tools.DirectoryProviderMock
	OsWrapper               *tools.CommonOsWrapperMock
	ResticContainerExecutor *tools.ResticContainerExecutorMock
}

func newResticDockerImageServiceTestObjects(t *testing.T) resticDockerImageServiceTestObjects {
	authHelper := tools.NewAuthHelperMock(t)
	configsRepository := configs.NewConfigsRepositoryMock(t)
	dockerService := tools.NewDockerServiceMock(t)
	directoryProvider := tools.NewDirectoryProviderMock(t)
	osWrapper := tools.NewCommonOsWrapperMock(t)
	resticContainerExecutor := tools.NewResticContainerExecutorMock(t)

	service := &ResticDockerImageServiceImpl{
		AuthHelper:              authHelper,
		ConfigsRepository:       configsRepository,
		DockerService:           dockerService,
		DirectoryProvider:       directoryProvider,
		OsWrapper:               osWrapper,
		ResticContainerExecutor: resticContainerExecutor,
	}
	return resticDockerImageServiceTestObjects{
		Service:                 service,
		AuthHelper:              authHelper,
		ConfigsRepository:       configsRepository,
		DockerService:           dockerService,
		DirectoryProvider:       directoryProvider,
		OsWrapper:               osWrapper,
		ResticContainerExecutor: resticContainerExecutor,
	}
}

func assertResticDockerImageServiceExpectations(t *testing.T, testObjects resticDockerImageServiceTestObjects) {
	testObjects.AuthHelper.AssertExpectations(t)
	testObjects.ConfigsRepository.AssertExpectations(t)
	testObjects.DockerService.AssertExpectations(t)
	testObjects.DirectoryProvider.AssertExpectations(t)
	testObjects.OsWrapper.AssertExpectations(t)
	testObjects.ResticContainerExecutor.AssertExpectations(t)
}

func TestUpdateResticDockerImage_WhenHashMatchesAndImageExists_SkipsBuild(t *testing.T) {
	testObjects := newResticDockerImageServiceTestObjects(t)
	defer assertResticDockerImageServiceExpectations(t, testObjects)

	expectResticDockerfileIsReadAndHashed(testObjects)
	testObjects.DockerService.EXPECT().DoesDockerImageExist(tools.ResticImageName).Return(true, nil)
	expectStoredResticDockerfileHash(testObjects, resticDockerfileHash)

	err := testObjects.Service.UpdateResticDockerImage()

	assert.Nil(t, err)
}

func TestUpdateResticDockerImage_WhenHashChanged_BuildsAndStoresHashAndPreparesDirectories(t *testing.T) {
	testObjects := newResticDockerImageServiceTestObjects(t)
	defer assertResticDockerImageServiceExpectations(t, testObjects)

	expectResticDockerfileIsReadAndHashed(testObjects)
	testObjects.DockerService.EXPECT().DoesDockerImageExist(tools.ResticImageName).Return(true, nil)
	expectStoredResticDockerfileHash(testObjects, oldResticDockerfileHash)
	expectResticImageIsBuiltAndHashIsStored(testObjects)
	expectResticContainerDirectoriesArePrepared(testObjects)

	err := testObjects.Service.UpdateResticDockerImage()

	assert.Nil(t, err)
}

func TestUpdateResticDockerImage_WhenImageIsMissing_BuildsAndStoresHashAndPreparesDirectories(t *testing.T) {
	testObjects := newResticDockerImageServiceTestObjects(t)
	defer assertResticDockerImageServiceExpectations(t, testObjects)

	expectResticDockerfileIsReadAndHashed(testObjects)
	testObjects.DockerService.EXPECT().DoesDockerImageExist(tools.ResticImageName).Return(false, nil)
	expectStoredResticDockerfileHash(testObjects, resticDockerfileHash)
	expectResticImageIsBuiltAndHashIsStored(testObjects)
	expectResticContainerDirectoriesArePrepared(testObjects)

	err := testObjects.Service.UpdateResticDockerImage()

	assert.Nil(t, err)
}

func TestUpdateResticDockerImage_WhenNoStoredHashExists_BuildsAndStoresHashAndPreparesDirectories(t *testing.T) {
	testObjects := newResticDockerImageServiceTestObjects(t)
	defer assertResticDockerImageServiceExpectations(t, testObjects)

	expectResticDockerfileIsReadAndHashed(testObjects)
	testObjects.DockerService.EXPECT().DoesDockerImageExist(tools.ResticImageName).Return(true, nil)
	testObjects.ConfigsRepository.EXPECT().IsConfigSet(configs.ConfigKeys.ResticDockerfileHash).Return(false, nil)
	expectResticImageIsBuiltAndHashIsStored(testObjects)
	expectResticContainerDirectoriesArePrepared(testObjects)

	err := testObjects.Service.UpdateResticDockerImage()

	assert.Nil(t, err)
}

func expectResticDockerfileIsReadAndHashed(testObjects resticDockerImageServiceTestObjects) {
	testObjects.DirectoryProvider.EXPECT().GetResticImageDockerfilePath().Return(resticDockerfilePath)
	testObjects.OsWrapper.EXPECT().ReadFile(resticDockerfilePath).Return(resticDockerfileBytes, nil)
	testObjects.AuthHelper.EXPECT().GetSHA256Hash(string(resticDockerfileBytes)).Return(resticDockerfileHash)
}

func expectStoredResticDockerfileHash(testObjects resticDockerImageServiceTestObjects, storedHash string) {
	testObjects.ConfigsRepository.EXPECT().IsConfigSet(configs.ConfigKeys.ResticDockerfileHash).Return(true, nil)
	testObjects.ConfigsRepository.EXPECT().GetConfig(configs.ConfigKeys.ResticDockerfileHash).Return(storedHash, nil)
}

func expectResticImageIsBuiltAndHashIsStored(testObjects resticDockerImageServiceTestObjects) {
	testObjects.DockerService.EXPECT().BuildDockerImage(tools.ResticImageName, resticDockerfilePath, true).Return(nil)
	testObjects.ConfigsRepository.EXPECT().SetConfig(configs.ConfigKeys.ResticDockerfileHash, resticDockerfileHash).Return(nil)
}

func expectResticContainerDirectoriesArePrepared(testObjects resticDockerImageServiceTestObjects) {
	testObjects.ResticContainerExecutor.EXPECT().ExecuteSimple("mkdir -p /root/.config/rclone").Return(&tools.CommandOutput{}, nil)
	testObjects.ResticContainerExecutor.EXPECT().ExecuteSimple("mkdir -p /root/.ssh").Return(&tools.CommandOutput{}, nil)
}
