package apistructs

import "time"

const LatestVersion = "latest"

type PipelineVersionDefinition struct {
	Name     string                          `json:"name"`
	Version  string                          `json:"version"`
	Status   PipelineVersionDefinitionStatus `json:"status"`
	Content  string                          `json:"content"`
	CreateAt time.Time                       `json:"createTime"`
	Creater  string                          `json:"creater"`
	Latest   bool                            `yaml:"latest"`
}

type PipelineDefinition struct {
	Name     string    `json:"name"`
	Desc     string    `json:"desc"`
	Public   bool      `json:"public"`
	CreateAt time.Time `json:"createTime"`
	Creater  string    `json:"creater"`

	VersionList []PipelineVersionDefinition `json:"versionList"`
}

type PipelineDefinitionPage struct {
	List  []PipelineDefinition `json:"list"`
	Total int64                `json:"total"`
}
