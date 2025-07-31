package service

import (
	"encoding/json"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"mime/multipart"
	"sync"
	"trojan-panel/core"
	"trojan-panel/dao"
	"trojan-panel/model"
	"trojan-panel/model/constant"
	"trojan-panel/model/dto"
	"trojan-panel/model/vo"
	"trojan-panel/util"
)

func SelectNodeServerById(id *uint) (*model.NodeServer, error) {
	return dao.SelectNodeServer(map[string]interface{}{"id": *id})
}

func CreateNodeServer(nodeServer *model.NodeServer) error {
	count, err := dao.CountNodeServerByName(nil, nodeServer.Name)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(constant.NodeServerNameExist)
	}
	return dao.CreateNodeServer(nodeServer)
}

func SelectNodeServerPage(queryName *string, queryIp *string, pageNum *uint, pageSize *uint, c *gin.Context) (*vo.NodeServerPageVo, error) {
	nodeServerPage, total, err := dao.SelectNodeServerPage(queryName, queryIp, pageNum, pageSize)
	if err != nil {
		return nil, err
	}
	nodeServerVos := make([]vo.NodeServerVo, 0)
	for _, item := range *nodeServerPage {
		nodeServerVo := vo.NodeServerVo{
			Id:         *item.Id,
			Name:       *item.Name,
			Ip:         *item.Ip,
			GrpcPort:   *item.GrpcPort,
			CreateTime: *item.CreateTime,
		}
		nodeServerVos = append(nodeServerVos, nodeServerVo)
	}

	account := GetCurrentAccount(c)
	if util.IsAdmin(account.Roles) {
		token := util.GetToken(c)
		var nodeMap sync.Map
		var wg sync.WaitGroup
		wg.Add(len(nodeServerVos))
		for i := range nodeServerVos {
			indexI := i
			go func() {
				var ip = nodeServerVos[indexI].Ip
				var grpcPort = nodeServerVos[indexI].GrpcPort
				nodeMapValue, ok := nodeMap.Load(ip)
				if ok {
					nodeServerVo := nodeMapValue.(vo.NodeServerVo)
					nodeServerVos[indexI].Status = nodeServerVo.Status
					nodeServerVos[indexI].TrojanPanelCoreVersion = nodeServerVo.TrojanPanelCoreVersion
				} else {
					var nodeServerState int
					var trojanPanelCoreVersion string
					stateVo, err := core.GetNodeServerState(token, ip, grpcPort)
					if err != nil {
						nodeServerState = 0
					} else {
						nodeServerState = 1
						trojanPanelCoreVersion = stateVo.GetVersion()
					}
					nodeServerVos[indexI].Status = nodeServerState
					nodeServerVos[indexI].TrojanPanelCoreVersion = trojanPanelCoreVersion
					nodeMap.Store(ip, nodeServerVos[indexI])
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}

	nodeServerPageVo := vo.NodeServerPageVo{
		BaseVoPage: vo.BaseVoPage{
			PageNum:  *pageNum,
			PageSize: *pageSize,
			Total:    total,
		},
		NodeServers: nodeServerVos,
	}
	return &nodeServerPageVo, nil
}

func DeleteNodeServerById(id *uint) error {
	count, err := dao.CountNodeByNameAndNodeServerId(nil, nil, id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(constant.NodeServerDeletedError)
	}

	return dao.DeleteNodeServerById(id)
}

func UpdateNodeServerById(dto *dto.NodeServerUpdateDto) error {
	count, err := dao.CountNodeByNameAndNodeServerId(nil, nil, dto.Id)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(constant.NodeServerDeletedError)
	}

	count, err = dao.CountNodeServerByName(dto.Id, dto.Name)
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New(constant.NodeServerNameExist)
	}

	nodeServer := model.NodeServer{
		Id:       dto.Id,
		Ip:       dto.Ip,
		Name:     dto.Name,
		GrpcPort: dto.GrpcPort,
	}
	return dao.UpdateNodeServerById(&nodeServer)
}

func CountNodeServer() (int, error) {
	return dao.CountNodeServer()
}

func CountNodeServerByName(id *uint, queryName *string) (int, error) {
	return dao.CountNodeServerByName(id, queryName)
}

func SelectNodeServerList(dto *dto.NodeServerDto) ([]vo.NodeServerListVo, error) {
	nodeServerList, err := dao.SelectNodeServerList(dto.Ip, dto.Name)
	if err != nil {
		return nil, err
	}
	nodeServerListVos := make([]vo.NodeServerListVo, 0)
	for _, item := range nodeServerList {
		nodeServerVo := vo.NodeServerListVo{
			Id:   *item.Id,
			Name: *item.Name,
		}
		nodeServerListVos = append(nodeServerListVos, nodeServerVo)

	}
	return nodeServerListVos, nil
}

func GetNodeServerInfo(token string, nodeServerId *uint) (*core.NodeServerInfoVo, error) {
	nodeServer, err := dao.SelectNodeServer(map[string]interface{}{"id": *nodeServerId})
	if err != nil {
		return nil, err
	}
	nodeServerInfoVo, err := core.GetNodeServerInfo(token, *nodeServer.Ip, *nodeServer.GrpcPort)
	if err != nil {
		return nil, err
	}
	return nodeServerInfoVo, nil
}

func ExportNodeServer(accountId uint, accountUsername string) error {
	nodeServerExportVos, err := dao.SelectNodeServerAll()
	if err != nil {
		return err
	}
	if err = ExportTaskJson(accountId, accountUsername, constant.TaskTypeNodeServerExport, "nodeServerExport", nodeServerExportVos); err != nil {
		return err
	}
	return nil
}

func ImportNodeServer(cover uint, file *multipart.FileHeader, accountId uint, accountUsername string) error {
	fileName := file.Filename

	var fileTaskType uint = constant.TaskTypeNodeServerImport
	var fileTaskStatus = constant.TaskDoing
	fileTask := model.FileTask{
		Name:            &fileName,
		Path:            nil,
		Type:            &fileTaskType,
		Status:          &fileTaskStatus,
		AccountId:       &accountId,
		AccountUsername: &accountUsername,
	}
	fileTaskId, err := dao.CreateFileTask(&fileTask)
	if err != nil {
		return err
	}

	go func(fileTaskId uint) {
		var fail = constant.TaskFail
		var success = constant.TaskSuccess
		fileTask := model.FileTask{
			Id:     &fileTaskId,
			Status: &fail,
		}

		src, err := file.Open()
		defer src.Close()
		if err != nil {
			logrus.Errorf("ImportNodeServer file Open err: %v", err)
			fileUploadError := constant.FileUploadError
			fileTask.ErrMsg = &fileUploadError
			if err = dao.UpdateFileTaskById(&fileTask); err != nil {
				logrus.Errorf("ImportNodeServer UpdateFileTaskById err: %v", err)
			}
			return
		}

		var nodeServers []model.NodeServer
		decoder := json.NewDecoder(src)
		if err = decoder.Decode(&nodeServers); err != nil {
			logrus.Errorf("ImportNodeServer decoder Decode err: %v", err)
			fileUploadError := constant.RowNotEnough
			fileTask.ErrMsg = &fileUploadError
			if err = dao.UpdateFileTaskById(&fileTask); err != nil {
				logrus.Errorf("ImportNodeServer UpdateFileTaskById err: %v", err)
			}
			return
		}
		if len(nodeServers) == 0 {
			logrus.Errorf("ImportNodeServer err: %s", constant.RowNotEnough)
			fileUploadError := constant.RowNotEnough
			fileTask.ErrMsg = &fileUploadError
			if err = dao.UpdateFileTaskById(&fileTask); err != nil {
				logrus.Errorf("ImportNodeServer UpdateFileTaskById err: %v", err)
			}
			return
		}
		// 在这里可以处理数据并将其存储到数据库中 todo 这里可能存在性能问题
		for _, item := range nodeServers {
			if err = dao.CreateOrUpdateNodeServer(item, cover); err != nil {
				continue
			}
		}
		fileTask.Status = &success
		// 更新文件任务状态
		if err = dao.UpdateFileTaskById(&fileTask); err != nil {
			logrus.Errorf("ImportNodeServer UpdateFileTaskById err: %v", err)
		}
	}(fileTaskId)
	return nil
}
