package tasks

import (
	lakeModels "github.com/merico-dev/lake/models"
	domainlayerBase "github.com/merico-dev/lake/plugins/domainlayer/models/base"
	"github.com/merico-dev/lake/plugins/domainlayer/models/ticket"
	"github.com/merico-dev/lake/plugins/domainlayer/okgen"
	jiraModels "github.com/merico-dev/lake/plugins/jira/models"
	"gorm.io/gorm/clause"
)

func ConvertIssues(boardId uint64) error {

	jiraIssue := &jiraModels.JiraIssue{}
	// select all issues belongs to the board
	cursor, err := lakeModels.Db.Model(jiraIssue).
		Select("jira_issues.*").
		Joins("left join jira_board_issues on jira_board_issues.issue_id = jira_issues.id").
		Where("jira_board_issues.board_id = ?", boardId).
		Rows()
	if err != nil {
		return err
	}

	boardOriginKey := okgen.NewOriginKeyGenerator(&jiraModels.JiraBoard{}).Generate(boardId)
	issueOriginKeyGenerator := okgen.NewOriginKeyGenerator(&jiraModels.JiraIssue{})

	// iterate all rows
	for cursor.Next() {
		err = lakeModels.Db.ScanRows(cursor, jiraIssue)
		if err != nil {
			return err
		}
		issue := &ticket.Issue{
			DomainEntity: domainlayerBase.DomainEntity{
				OriginKey: issueOriginKeyGenerator.Generate(jiraIssue.ID),
			},
			BoardOriginKey: boardOriginKey,
			Url:            jiraIssue.Self,
			Key:            jiraIssue.Key,
			Summary:        jiraIssue.Summary,
			EpicKey:        jiraIssue.EpicKey,
			Type:           jiraIssue.StdType,
			Status:         jiraIssue.StdStatus,
			StoryPoint:     jiraIssue.StdStoryPoint,
			ResolutionDate: jiraIssue.ResolutionDate,
			CreatedDate:    jiraIssue.Created,
			UpdatedDate:    jiraIssue.Updated,
			LeadTime:       jiraIssue.LeadTime,
		}

		err = lakeModels.Db.Clauses(clause.OnConflict{UpdateAll: true}).Create(issue).Error
		if err != nil {
			return err
		}
	}
	return nil
}
