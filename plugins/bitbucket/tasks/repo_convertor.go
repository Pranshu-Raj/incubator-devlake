/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package tasks

import (
	"fmt"
	"reflect"

	"github.com/apache/incubator-devlake/plugins/core/dal"

	"github.com/apache/incubator-devlake/models/domainlayer"
	"github.com/apache/incubator-devlake/models/domainlayer/code"
	"github.com/apache/incubator-devlake/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/plugins/bitbucket/models"
	"github.com/apache/incubator-devlake/plugins/core"
	"github.com/apache/incubator-devlake/plugins/helper"
)

var ConvertRepoMeta = core.SubTaskMeta{
	Name:             "convertRepo",
	EntryPoint:       ConvertRepo,
	EnabledByDefault: true,
	Description:      "Convert tool layer table bitbucket_repos into  domain layer table repos and boards",
	DomainTypes:      core.DOMAIN_TYPES,
}

func ConvertRepo(taskCtx core.SubTaskContext) error {
	db := taskCtx.GetDal()
	data := taskCtx.GetData().(*BitbucketTaskData)
	repoId := data.Repo.BitbucketId

	cursor, err := db.Cursor(
		dal.From(&models.BitbucketRepo{}),
		dal.Where("bitbucket_id = ?", repoId),
	)
	if err != nil {
		return err
	}
	defer cursor.Close()

	repoIdGen := didgen.NewDomainIdGenerator(&models.BitbucketRepo{})

	converter, err := helper.NewDataConverter(helper.DataConverterArgs{
		InputRowType: reflect.TypeOf(models.BitbucketRepo{}),
		Input:        cursor,
		RawDataSubTaskArgs: helper.RawDataSubTaskArgs{
			Ctx: taskCtx,
			Params: BitbucketApiParams{
				ConnectionId: data.Options.ConnectionId,
				Owner:        data.Options.Owner,
				Repo:         data.Options.Repo,
			},
			Table: RAW_REPOSITORIES_TABLE,
		},
		Convert: func(inputRow interface{}) ([]interface{}, error) {
			repository := inputRow.(*models.BitbucketRepo)
			domainRepository := &code.Repo{
				DomainEntity: domainlayer.DomainEntity{
					Id: repoIdGen.Generate(data.Options.ConnectionId, repository.BitbucketId),
				},
				Name:        repository.Name,
				Url:         repository.HTMLUrl,
				Description: repository.Description,
				Language:    repository.Language,
				CreatedDate: repository.CreatedDate,
				UpdatedDate: repository.UpdatedDate,
			}

			domainBoard := &ticket.Board{
				DomainEntity: domainlayer.DomainEntity{
					Id: repoIdGen.Generate(data.Options.ConnectionId, repository.BitbucketId),
				},
				Name:        repository.Name,
				Url:         fmt.Sprintf("%s/%s", repository.HTMLUrl, "issues"),
				Description: repository.Description,
				CreatedDate: &repository.CreatedDate,
			}

			return []interface{}{
				domainRepository,
				domainBoard,
			}, nil
		},
	})
	if err != nil {
		return err
	}

	return converter.Execute()
}
