package register

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
	"net/http"
	"tiggerops/apistructs"
	"tiggerops/internal/register/client/pipelinedefinitionclient"
	"tiggerops/pkg/responsehandler"
	"tiggerops/pkg/schema/pipeline"
	"tiggerops/pkg/token"
)

type PipelineNameVersionUri struct {
	Name    string `uri:"name" binding:"required"`
	Version string `uri:"version"`
}

func (r *Service) GetPipelineVersion(c *gin.Context) {
	var nameVersion = PipelineNameVersionUri{}
	if err := c.ShouldBindUri(&nameVersion); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get name or version error: %v", err), nil))
		return
	}

	name := nameVersion.Name
	version := nameVersion.Version
	creater := c.Query("creater")
	if creater == "" {
		creater = token.GetUserName(c)
	}

	dbPipelineDefinition, pipelineDefinitionFind, err := r.pipelineVersionDefinitionClient.GetPipelineDefinition(nil, name, creater)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline definition error: %v", err), nil))
		return
	}
	if !pipelineDefinitionFind {
		c.JSON(responsehandler.Build(http.StatusOK, "", nil))
		return
	}

	dbPipelineVersionDefinition, pipelineVersionDefinitionFind, err := r.pipelineVersionDefinitionClient.GetPipelineVersionDefinition(nil, name, version, creater)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline definition error: %v", err), nil))
		return
	}
	if !pipelineVersionDefinitionFind {
		c.JSON(responsehandler.Build(http.StatusOK, "", nil))
		return
	}

	if !dbPipelineDefinition.Public &&
		(dbPipelineDefinition.Creater != token.GetUserName(c) || dbPipelineVersionDefinition.Creater != token.GetUserName(c)) {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("auth failed"), nil))
		return
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", dbPipelineVersionDefinition.ToApiStructs()))
	return
}

func (r *Service) GetPipeline(c *gin.Context) {
	var nameVersion = PipelineNameVersionUri{}
	if err := c.ShouldBindUri(&nameVersion); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get name or version error: %v", err), nil))
		return
	}

	name := nameVersion.Name
	creater := c.Query("creater")
	if creater == "" {
		creater = token.GetUserName(c)
	}

	dbPipelineDefinition, pipelineDefinitionFind, err := r.pipelineVersionDefinitionClient.GetPipelineDefinition(nil, name, creater)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline definition error: %v", err), nil))
		return
	}
	if !pipelineDefinitionFind {
		c.JSON(responsehandler.Build(http.StatusOK, "", nil))
		return
	}
	if !dbPipelineDefinition.Public &&
		(dbPipelineDefinition.Creater != token.GetUserName(c)) {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("auth failed"), nil))
		return
	}

	versionList, err := r.pipelineVersionDefinitionClient.ListPipelineVersionDefinition(nil, &pipelinedefinitionclient.PipelineVersionQueryCompose{
		VersionQueryList: []pipelinedefinitionclient.PipelineVersionQuery{
			{
				Name:    dbPipelineDefinition.Name,
				Creater: dbPipelineDefinition.Creater,
			},
		},
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("get pipeline definition version list error: %v", err), nil))
		return
	}

	var result = dbPipelineDefinition.ToApiStructs()
	for _, version := range versionList {
		result.VersionList = append(result.VersionList, version.ToApiStructs())
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", result))
	return
}

//func (r *Service) PagePipeline(c *gin.Context) {
//	public, err := strconv.ParseBool(c.Query("public"))
//	if err != nil {
//		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("public parse error: %v", err), nil))
//		return
//	}
//
//	page, pageSize, err := pagehelper.GetPageAndPageSizeFromGinContext(c)
//	if err != nil {
//		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("page or pageSize get error: %v", err), nil))
//		return
//	}
//
//	var listPipelineDefinitionQuery = client.ListPipelineDefinitionQuery{
//		NameLike: c.Query("search"),
//		Page:     int(page),
//		PageSize: int(pageSize),
//	}
//
//	if public {
//		listPipelineDefinitionQuery.Public = &[]bool{public}[0]
//	} else {
//		listPipelineDefinitionQuery.Creater = token.GetUserName(c)
//	}
//
//	result, total, err := r.pipelineDefinitionDbClient.PagePipelineDefinition(nil, listPipelineDefinitionQuery)
//	if err != nil {
//		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("list pipeline definition list error: %v", err), nil))
//		return
//	}
//	var pipelineDefinitionPage apistructs.PipelineDefinitionPage
//	pipelineDefinitionPage.Total = total
//	for _, definition := range result {
//		pipelineDefinitionPage.List = append(pipelineDefinitionPage.List, apistructs.PipelineDefinition{
//			Name:     definition.Name,
//			Desc:     definition.Desc,
//			Creater:  definition.Creater,
//			Public:   definition.Public,
//			CreateAt: definition.CreatedAt,
//		})
//	}
//
//	c.JSON(responsehandler.Build(http.StatusOK, "", pipelineDefinitionPage))
//	return
//}

func (r *Service) ListMyPipelineVersion(c *gin.Context) {
	dbResult, err := r.pipelineVersionDefinitionClient.ListPipelineVersionDefinition(nil, &pipelinedefinitionclient.PipelineVersionQueryCompose{
		Creater: token.GetUserName(c),
	})
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get pipeline version definition error: %v", err), nil))
		return
	}

	var versionList []apistructs.PipelineVersionDefinition
	for _, vd := range dbResult {
		versionList = append(versionList, vd.ToApiStructs())
	}

	c.JSON(responsehandler.Build(http.StatusOK, "", versionList))
}

type ApplyPipelineRequest struct {
	PipelineContent string `json:"pipelineContent"`
}

func (r *Service) ApplyPipeline(c *gin.Context) {
	var applyInfo ApplyPipelineRequest
	if err := c.ShouldBind(&applyInfo); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	pipeInfo, applyContent, err := r.applyPipelineMutatingAndCheck(c, applyInfo.PipelineContent)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}

	err = r.applyPipeline(c, pipeInfo.Name, pipeInfo.Version, applyContent)
	if err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, err.Error(), nil))
		return
	}
	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
}

func (r *Service) applyPipelineMutatingAndCheck(c *gin.Context, content string) (*pipeline.Pipeline, string, error) {
	var pipeInfo = &pipeline.Pipeline{}
	err := yaml.Unmarshal([]byte(content), pipeInfo)
	if err != nil {
		return nil, "", fmt.Errorf("pipeline yaml content unmarshal error: %v", err)
	}

	pipeInfo.PipelineTypeTaskImageMutating(token.GetUserName(c))

	// 获取流水线类型的任务对应的流水线定义
	taskImagePipelineDefinitionMap, err := r.getTaskImagePipelineInfoMap(token.GetUserName(c), pipeInfo)
	if err != nil {
		return nil, "", err
	}

	if err := pipeInfo.Mutating(taskImagePipelineDefinitionMap); err != nil {
		return nil, "", fmt.Errorf("pipeline Mutating fieled error: %v", err)
	}

	yamlContent, err := yaml.Marshal(pipeInfo)
	if err != nil {
		return nil, "", fmt.Errorf("pipeline yaml content marshal error: %v", err)
	}
	applyContent := string(yamlContent)

	if err := pipeInfo.Check(applyContent, taskImagePipelineDefinitionMap); err != nil {
		return nil, "", fmt.Errorf("pipeline yaml check failed: %v", err)
	}
	return pipeInfo, applyContent, nil
}

func (r *Service) getTaskImagePipelineInfoMap(creater string, pipeInfo *pipeline.Pipeline) (map[string]pipeline.Pipeline, error) {
	var taskImagePipelineDefinitionMap = map[string]pipeline.Pipeline{}
	pipelineTypeTaskDefinitions, err := r.getPipelineTaskYmlDefinition(creater, pipeInfo.GetPipelineTypeTask())
	if err != nil {
		return taskImagePipelineDefinitionMap, err
	}

	for image, definition := range pipelineTypeTaskDefinitions {
		var pipeInfo pipeline.Pipeline
		err = yaml.Unmarshal([]byte(definition.Content), &pipeInfo)
		if err != nil {
			return taskImagePipelineDefinitionMap, fmt.Errorf("task (image %v) Unmarshal yaml content error %v", image, err)
		}
		taskImagePipelineDefinitionMap[image] = pipeInfo
	}
	return taskImagePipelineDefinitionMap, nil
}

func (r *Service) getPipelineTaskYmlDefinition(creater string, pipelineTypeTasks []pipeline.Task) (map[string]*pipelinedefinitionclient.PipelineVersionDefinition, error) {
	var result = map[string]*pipelinedefinitionclient.PipelineVersionDefinition{}

	if len(pipelineTypeTasks) == 0 {
		return result, nil
	}

	var queryCompose = pipelinedefinitionclient.PipelineVersionQueryCompose{}
	for _, task := range pipelineTypeTasks {
		query := buildPipelineVersionQuery(creater, task)
		queryCompose.VersionQueryList = append(queryCompose.VersionQueryList, query)
	}

	definitions, err := r.pipelineVersionDefinitionClient.ListPipelineVersionDefinition(nil, &queryCompose)
	if err != nil {
		return result, err
	}

	for _, task := range pipelineTypeTasks {
		query := buildPipelineVersionQuery(creater, task)
		var findDefinition *pipelinedefinitionclient.PipelineVersionDefinition
		for _, definition := range definitions {
			if definition.Name != query.Name {
				continue
			}
			if query.Version == "" && !definition.Latest {
				continue
			}
			if query.Version != "" && query.Version != definition.Version {
				continue
			}
			if query.Creater != definition.Creater {
				continue
			}
			findDefinition = &definition
			break
		}
		if findDefinition == nil {
			return result, fmt.Errorf("task (alias %v) not find pipeline definition (image %v)", task.Alias, task.Image)
		}

		if findDefinition.Status != apistructs.CreatedStatus {
			return result, fmt.Errorf("task (alias %v) pipeline definition (image %v) not a %v status", task.Alias, task.Image, apistructs.CreatedStatus)
		}

		result[task.Image] = findDefinition
	}

	return result, nil
}

func buildPipelineVersionQuery(user string, task pipeline.Task) pipelinedefinitionclient.PipelineVersionQuery {
	query := pipelinedefinitionclient.PipelineVersionQuery{}

	if task.Type != pipeline.PipeType {
		return query
	}

	query.Name = task.GetPipelineTypeTaskName()
	query.Version = task.GetPipelineTypeTaskVersion()
	query.Creater = task.GetPipelineTypeTaskCreater()

	if query.Creater == "" {
		query.Creater = user
	} else {
		if query.Creater != user {
			query.Public = &[]bool{true}[0]
		}
	}
	return query
}

func (r *Service) applyPipeline(c *gin.Context, name, version, applyContent string) error {

	_, pipelineDefinitionFind, err := r.pipelineVersionDefinitionClient.GetPipelineDefinition(nil, name, token.GetUserName(c))
	if err != nil {
		return fmt.Errorf("failed to get pipeline definition error: %v", err)
	}
	dbPipelineVersionDefinition, pipelineVersionDefinitionFind, err := r.pipelineVersionDefinitionClient.GetPipelineVersionDefinition(nil, name, version, token.GetUserName(c))
	if err != nil {
		return fmt.Errorf("failed to get pipeline definition error: %v", err)
	}

	err = r.dbClient.Transaction(func(tx *gorm.DB) error {
		if !pipelineDefinitionFind {
			var pipelineDefinition = pipelinedefinitionclient.PipelineDefinition{
				Name:    name,
				Public:  false,
				Creater: token.GetUserName(c),
			}
			_, err := r.pipelineVersionDefinitionClient.CreatePipelineDefinition(tx, &pipelineDefinition)
			if err != nil {
				return fmt.Errorf("create pipeline definition error: %v", err)
			}
		}

		latestVersionDefinition, _, err := r.pipelineVersionDefinitionClient.GetPipelineVersionDefinition(tx, name, apistructs.LatestVersion, token.GetUserName(c))
		if err != nil {
			return err
		}

		if pipelineVersionDefinitionFind {
			if dbPipelineVersionDefinition.Creater != token.GetUserName(c) {
				return fmt.Errorf("auth failed")
			}

			var preDbLatest = dbPipelineVersionDefinition.Latest
			setLatestVersion(dbPipelineVersionDefinition, latestVersionDefinition)
			// clear latest definition
			if latestVersionDefinition != nil && dbPipelineVersionDefinition.Latest {
				_, err = r.pipelineVersionDefinitionClient.UpdatePipelineVersionDefinition(tx, latestVersionDefinition)
				if err != nil {
					return fmt.Errorf("update pipeline definition error: %v", err)
				}
			}

			if dbPipelineVersionDefinition.Content == applyContent && preDbLatest == dbPipelineVersionDefinition.Latest {
				return nil
			}
			// clear status
			dbPipelineVersionDefinition.Content = applyContent
			dbPipelineVersionDefinition.Status = apistructs.CreatedStatus
			dbPipelineVersionDefinition.StatusMessage = ""
			_, err = r.pipelineVersionDefinitionClient.UpdatePipelineVersionDefinition(tx, dbPipelineVersionDefinition)
			if err != nil {
				return fmt.Errorf("update pipeline definition error: %v", err)
			}
		} else {

			var pipelineDefinition = pipelinedefinitionclient.PipelineVersionDefinition{
				Name:    name,
				Version: version,
				Content: applyContent,
				Status:  apistructs.CreatedStatus,
				Creater: token.GetUserName(c),
			}

			setLatestVersion(&pipelineDefinition, latestVersionDefinition)
			// clear latest definition
			if latestVersionDefinition != nil && pipelineDefinition.Latest {
				_, err = r.pipelineVersionDefinitionClient.UpdatePipelineVersionDefinition(tx, latestVersionDefinition)
				if err != nil {
					return fmt.Errorf("update pipeline definition error: %v", err)
				}
			}

			_, err := r.pipelineVersionDefinitionClient.CreatePipelineVersionDefinition(tx, &pipelineDefinition)
			if err != nil {
				return fmt.Errorf("create pipeline definition error: %v", err)
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *Service) DeletePipeline(c *gin.Context) {
	var nameVersion = PipelineNameVersionUri{}
	if err := c.ShouldBindUri(&nameVersion); err != nil {
		c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("failed to get name or version error: %v", err), nil))
		return
	}
	if nameVersion.Version == "" {
		err := r.pipelineVersionDefinitionClient.DeletePipelineDefinition(nil, nameVersion.Name, token.GetUserName(c))
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("delete pipeline definition failed. error: %v", err), nil))
			return
		}
	} else {
		err := r.pipelineVersionDefinitionClient.DeletePipelineVersionDefinition(nil, nameVersion.Name, nameVersion.Version, token.GetUserName(c))
		if err != nil {
			c.JSON(responsehandler.Build(http.StatusServiceUnavailable, fmt.Sprintf("delete pipeline version definition failed. error: %v", err), nil))
			return
		}
	}
	c.JSON(responsehandler.Build(http.StatusOK, "", nil))
	return
}

func setLatestVersion(thisDefinition *pipelinedefinitionclient.PipelineVersionDefinition, latestDefinition *pipelinedefinitionclient.PipelineVersionDefinition) {
	if latestDefinition == nil && thisDefinition == nil {
		return
	}

	if latestDefinition == nil {
		thisDefinition.Latest = true
	} else {
		if thisDefinition.Version > latestDefinition.Version {
			thisDefinition.Latest = true
			latestDefinition.Latest = false
		}
	}
}
