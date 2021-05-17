package aws

import (
	"context"

	"github.com/aws/aws-sdk-go/service/auditmanager"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
)

type evidenceInfo = struct {
	auditmanager.GetThreatIntelSetOutput
	ThreatIntelSetID string
	DetectorID       string
}

//// TABLE DEFINITION
func tableAwsAuditManagerEvidence(_ context.Context) *plugin.Table {
	return &plugin.Table{
		Name:        "aws_audit_manager_evidence",
		Description: "AWS Audit Manager Evidence",
		// Get: &plugin.GetConfig{
		// 	KeyColumns:        plugin.SingleColumn("id"),
		// 	ShouldIgnoreError: isNotFoundError([]string{"ResourceNotFoundException"}),
		// 	Hydrate:           getAwsAuditManagerEvidence,
		// },
		List: &plugin.ListConfig{
			Hydrate: listAwsAuditManagerEvidences,
		},
		GetMatrixItem: BuildRegionList,
		Columns: awsRegionalColumns([]*plugin.Column{
			{
				Name:        "id",
				Description: "The identifier for the evidence.",
				Type:        proto.ColumnType_STRING,
				// Hydrate:     getAwsAuditManagerEvidence,
			},
			{
				Name:        "assessment_report_selection",
				Description: "Specifies whether the evidence is inclded in the assessment report.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "attributes",
				Description: "The names and values used by the evidence event, including an attribute name and value.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "aws_account_id",
				Description: "The identifier for the specified AWS account.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "aws_organization",
				Description: "The AWS account from which the evidence is collected, and its AWS organization path.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "compliance_check",
				Description: "The evaluation status for evidence that falls under the compliance check category.",
				Type:        proto.ColumnType_JSON,
				// Hydrate:     getAwsAuditManagerAssessment,
			},
			{
				Name:        "data_source",
				Description: "The data source from which the specified evidence was collected.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "event_name",
				Description: "The name of the specified evidence event.",
				Type:        proto.ColumnType_STRING,
				// Hydrate:     getAwsAuditManagerAssessment,
			},
			{
				Name:        "event_source",
				Description: "The AWS service from which the evidence is collected.",
				Type:        proto.ColumnType_STRING,
				// Hydrate:     getAwsAuditManagerAssessment,
			},
			{
				Name:        "evidence_aws_account_id",
				Description: "The identifier for the specified AWS account.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "evidence_by_type",
				Description: "The type of automated evidence.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "evidence_folder_id",
				Description: "The identifier for the folder in which the evidence is stored.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "iam_id",
				Description: "The unique identifier for the IAM user or role associated with the evidence.",
				Type:        proto.ColumnType_STRING,
			},
			{
				Name:        "resources_included",
				Description: "The list of resources assessed to generate the evidence.",
				Type:        proto.ColumnType_JSON,
			},
			{
				Name:        "time",
				Description: "The timestamp that represents when the evidence was collected.",
				Type:        proto.ColumnType_TIMESTAMP,
			},
			// {
			// 	Name:        "tags_src",
			// 	Description: "The tags associated with the assessment.",
			// 	Type:        proto.ColumnType_JSON,
			// 	Hydrate:     getAwsAuditManagerAssessment,
			// 	Transform:   transform.FromField("Tags"),
			// },
			// // Standard columns for all tables
			// {
			// 	Name:        "title",
			// 	Description: resourceInterfaceDescription("title"),
			// 	Type:        proto.ColumnType_STRING,
			// 	Transform:   transform.FromField("Name"),
			// },
			// {
			// 	Name:        "tags",
			// 	Description: resourceInterfaceDescription("tags"),
			// 	Type:        proto.ColumnType_JSON,
			// 	Hydrate:     getAwsAuditManagerAssessment,
			// 	Transform:   transform.FromField("Tags").Transform(auditManagerAssessmentTagListToTurbotTags),
			// },
			// {
			// 	Name:        "akas",
			// 	Description: resourceInterfaceDescription("akas"),
			// 	Type:        proto.ColumnType_JSON,
			// 	Hydrate:     getAwsAuditManagerAssessment,
			// 	Transform:   transform.FromField("Arn").Transform(arnToAkas),
			// },
		}),
	}
}

//// LIST FUNCTION
func listAwsAuditManagerEvidences(ctx context.Context, d *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	var region string
	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
	if matrixRegion != nil {
		region = matrixRegion.(string)
	}
	plugin.Logger(ctx).Trace("listAwsAuditManagerEvidences", "AWS_REGION", region)
	// Create session
	svc, err := AuditManagerService(ctx, d, region)
	if err != nil {
		return nil, err
	}

	// assessmentId := d.KeyColumnQuals["assessment_id"].GetStringValue()
	// ControlSetId := d.KeyColumnQuals["ControlSetId"].GetStringValue()
	// EvidenceFolderId := d.KeyColumnQuals["EvidenceFolderId"].GetStringValue()

	// List call
	err = svc.GetEvidenceByEvidenceFolderPages(
		&auditmanager.GetEvidenceByEvidenceFolderInput{},
		func(page *auditmanager.GetEvidenceByEvidenceFolderOutput, isLast bool) bool {
			for _, evidence := range page.Evidence {
				d.StreamListItem(ctx, evidence)
			}
			return !isLast
		},
	)
	return nil, err
}

//// HYDRATE FUNCTIONS

// func getAwsAuditManagerEvidence(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
// 	logger := plugin.Logger(ctx)
// 	logger.Trace("getAwsAuditManagerEvidence")
// 	var region string
// 	matrixRegion := plugin.GetMatrixItem(ctx)[matrixKeyRegion]
// 	if matrixRegion != nil {
// 		region = matrixRegion.(string)
// 	}
// 	// var id string
// 	// if h.Item != nil {
// 	// 	i := h.Item.(*auditmanager.Evidence)
// 	// 	id = *i.Id
// 	// } else {
// 	// 	id = d.KeyColumnQuals["id"].GetStringValue()
// 	// }

// 	id = d.KeyColumnQuals["id"].GetStringValue()

// 	// Create Session
// 	svc, err := AuditManagerService(ctx, d, region)
// 	if err != nil {
// 		return nil, err
// 	}
// 	// Build the params
// 	params := &auditmanager.GetEvidenceInput{
// 		EvidenceId: &id,
// 	}
// 	// Get call
// 	data, err := svc.GetEvidence(params)
// 	if err != nil {
// 		logger.Debug("getAwsAuditManagerEvidence", "ERROR", err)
// 		return nil, err
// 	}
// 	return data, nil
// }

//// TRANSFORM FUNCTIONS
// func auditManagerAssessmentTagListToTurbotTags(ctx context.Context, d *transform.TransformData) (interface{}, error) {
// 	plugin.Logger(ctx).Trace("auditManagerAssessmentTagListToTurbotTags")
// 	tagList := d.HydrateItem.(*auditmanager.Assessment).Tags
// 	// Mapping the resource tags inside turbotTags
// 	var turbotTagsMap map[string]string
// 	if tagList != nil {
// 		turbotTagsMap = map[string]string{}
// 		for _, i := range tagList {
// 			turbotTagsMap[*i.Key] = *i.Value
// 		}
// 	} else {
// 		return nil, nil
// 	}
// 	return turbotTagsMap, nil
// }
