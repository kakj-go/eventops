package pipelinedefinitionclient

import (
	"database/sql/driver"
	"encoding/json"
	"eventops/apistructs"
	"fmt"
	"gorm.io/gorm"
	"strings"
	"time"
)

type Client struct {
	client *gorm.DB
}

func NewPipelineDefinitionClient(client *gorm.DB) *Client {
	return &Client{client: client}
}

type PipelineDefinition struct {
	Id      uint64 `json:"id"`
	Name    string `json:"name"`
	Public  bool   `json:"public"`
	Desc    string `json:"desc"`
	Creater string `json:"creater"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (d PipelineDefinition) ToApiStructs() apistructs.PipelineDefinition {
	return apistructs.PipelineDefinition{
		Name:     d.Name,
		Desc:     d.Desc,
		Public:   d.Public,
		CreateAt: d.CreatedAt,
		Creater:  d.Creater,
	}
}

type PipelineVersionDefinition struct {
	Id            uint64                                     `json:"id"`
	Name          string                                     `json:"name"`
	Version       string                                     `json:"version"`
	Content       string                                     `json:"content"`
	Creater       string                                     `json:"creater"`
	Status        apistructs.PipelineVersionDefinitionStatus `json:"status"`
	StatusMessage string                                     `json:"status_message"`
	Latest        bool                                       `json:"latest"`

	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt
}

func (version *PipelineVersionDefinition) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	b, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("value is not []byte, value: %v", value)
	}

	return json.Unmarshal(b, &version)
}

func (version *PipelineVersionDefinition) Value() (driver.Value, error) {
	if version == nil {
		return nil, nil
	}

	return json.Marshal(version)
}

func (version *PipelineVersionDefinition) ToApiStructs() apistructs.PipelineVersionDefinition {
	return apistructs.PipelineVersionDefinition{
		Name:     version.Name,
		Version:  version.Version,
		Status:   version.Status,
		Content:  version.Content,
		CreateAt: version.CreatedAt,
		Creater:  version.Creater,
		Latest:   version.Latest,
	}
}

type ListPipelineDefinitionQuery struct {
	NameLike string
	Creater  string
	Public   *bool
	Page     int
	PageSize int
}

func (client *Client) PagePipelineDefinition(tx *gorm.DB, listPipelineDefinitionQuery ListPipelineDefinitionQuery) ([]PipelineDefinition, int64, error) {
	if tx == nil {
		tx = client.client
	}

	if listPipelineDefinitionQuery.NameLike != "" {
		tx = tx.Where("name like %?%", listPipelineDefinitionQuery.NameLike)
	}
	if listPipelineDefinitionQuery.Creater != "" {
		tx = tx.Where("creater = ?", listPipelineDefinitionQuery.Creater)
	}
	if &listPipelineDefinitionQuery.Public != nil {
		tx = tx.Where("public = ?", listPipelineDefinitionQuery.Public)
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var pipelineDefinitionList []PipelineDefinition
	offset := (listPipelineDefinitionQuery.Page - 1) * listPipelineDefinitionQuery.PageSize
	err := tx.Offset(offset).Limit(listPipelineDefinitionQuery.PageSize).Find(&pipelineDefinitionList).Error
	if err != nil {
		return nil, 0, err
	}

	return pipelineDefinitionList, total, nil
}

func (client *Client) GetPipelineDefinition(tx *gorm.DB, name string, creater string) (*PipelineDefinition, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var pipelineDefinition PipelineDefinition
	err := tx.Where("name = ? and creater = ?", name, creater).First(&pipelineDefinition).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &pipelineDefinition, true, nil
}

type PipelineVersionQuery struct {
	Name    string
	Creater string

	Version string
	Public  *bool
}

type PipelineVersionQueryCompose struct {
	VersionQueryList []PipelineVersionQuery
	Status           []apistructs.PipelineVersionDefinitionStatus
	Creater          string
}

func (client *Client) ListPipelineVersionDefinition(tx *gorm.DB, queryList *PipelineVersionQueryCompose) ([]PipelineVersionDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	tx = tx.Model(&PipelineVersionDefinition{}).Joins("left join pipeline_definitions on pipeline_definitions.name = pipeline_version_definitions.name and pipeline_definitions.creater = pipeline_version_definitions.creater")
	if len(queryList.VersionQueryList) > 0 {
		for _, query := range queryList.VersionQueryList {
			var sql = "1 = 1"
			var values []interface{}
			if query.Name != "" {
				sql += " and pipeline_version_definitions.name = ?"
				values = append(values, query.Name)
			}

			if query.Version != "" {
				if query.Version == apistructs.LatestVersion {
					sql += " and pipeline_version_definitions.latest = ?"
					values = append(values, true)
				} else {
					sql += " and pipeline_version_definitions.version = ?"
					values = append(values, query.Version)
				}
			}

			if query.Creater != "" {
				sql += " and pipeline_version_definitions.creater = ?"
				values = append(values, query.Creater)
			}
			if query.Public != nil {
				sql += " and pipeline_definitions.public = ?"
				values = append(values, query.Public)
			}
			tx = tx.Or(sql, values...)
		}
	} else {
		if queryList.Status != nil {
			tx = tx.Where("pipeline_version_definitions.status in ?", queryList.Status)
		}
		if queryList.Creater != "" {
			tx = tx.Where("pipeline_version_definitions.creater = ?", queryList.Creater)
		}
	}

	var pipelineVersionDefinitionList []PipelineVersionDefinition
	err := tx.Order("pipeline_version_definitions.version desc").Find(&pipelineVersionDefinitionList).Error
	if err != nil {
		return nil, err
	}

	return pipelineVersionDefinitionList, nil
}

func (client *Client) GetPipelineVersionDefinition(tx *gorm.DB, name string, version string, creater string) (*PipelineVersionDefinition, bool, error) {
	if tx == nil {
		tx = client.client
	}

	var pipelineVersionDefinition PipelineVersionDefinition
	var err error
	if version == apistructs.LatestVersion || strings.TrimSpace(version) == "" {
		err = tx.Where("name = ? and latest = ? and creater = ?", name, true, creater).First(&pipelineVersionDefinition).Error
	} else {
		err = tx.Where("name = ? and version = ? and creater = ?", name, version, creater).First(&pipelineVersionDefinition).Error
	}
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, false, nil
		}
		return nil, false, err
	}
	return &pipelineVersionDefinition, true, nil
}

func (client *Client) CreatePipelineDefinition(tx *gorm.DB, p *PipelineDefinition) (*PipelineDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (client *Client) UpdatePipelineDefinition(tx *gorm.DB, p *PipelineDefinition) (*PipelineDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Updates(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (client *Client) CreatePipelineVersionDefinition(tx *gorm.DB, pv *PipelineVersionDefinition) (*PipelineVersionDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Create(pv).Error
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (client *Client) UpdatePipelineVersionDefinition(tx *gorm.DB, pv *PipelineVersionDefinition) (*PipelineVersionDefinition, error) {
	if tx == nil {
		tx = client.client
	}

	err := tx.Model(pv).Select("*").Where("id = ?", pv.Id).Updates(pv).Error
	if err != nil {
		return nil, err
	}
	return pv, nil
}

func (client *Client) DeletePipelineDefinition(tx *gorm.DB, name, creater string) error {
	if tx == nil {
		tx = client.client
		return tx.Transaction(func(tx *gorm.DB) error {
			return client.deletePipelineDefinition(tx, name, creater)
		})
	} else {
		return client.deletePipelineDefinition(tx, name, creater)
	}
}

func (client *Client) deletePipelineDefinition(tx *gorm.DB, name, creater string) error {
	err := tx.Where("name = ? and creater = ?", name, creater).Delete(&PipelineDefinition{}).Error
	if err != nil {
		return err
	}
	err = tx.Where("name = ? and creater = ?", name, creater).Delete(&PipelineVersionDefinition{}).Error
	if err != nil {
		return err
	}
	return nil
}

func (client *Client) DeletePipelineVersionDefinition(tx *gorm.DB, name, version, creater string) error {
	if tx == nil {
		tx = client.client
	}

	return tx.Model(&PipelineVersionDefinition{}).Where("name = ? and version = ? and creater = ?", name, version, creater).Delete(&PipelineVersionDefinition{}).Error
}
