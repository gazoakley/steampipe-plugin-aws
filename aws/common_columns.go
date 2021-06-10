package aws

import (
	"context"
	"strings"

	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/turbot/steampipe-plugin-sdk/grpc/proto"
	"github.com/turbot/steampipe-plugin-sdk/plugin"
	"github.com/turbot/steampipe-plugin-sdk/plugin/transform"
)

// define cacheable hydrate function to return common column data
var commonColumnsHydrate = plugin.CacheableHydrate(getCommonColumns, getCommonColumnsCacheKey)

// column definitions for the common columns
func commonAwsRegionalColumns() []*plugin.Column {
	return []*plugin.Column{
		{
			Name:        "partition",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS partition in which the resource is located (aws, aws-cn, or aws-us-gov).",
		},
		{
			Name:        "region",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS Region in which the resource is located.",
		},
		{
			Name:        "account_id",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS Account ID in which the resource is located.",
			Transform:   transform.FromCamel(),
		},
	}
}

// column definitions for the common columns
func commonS3Columns() []*plugin.Column {
	return []*plugin.Column{
		{
			Name:        "partition",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS partition in which the resource is located (aws, aws-cn, or aws-us-gov).",
		},
		{
			Name:        "account_id",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Transform:   transform.FromCamel(),
			Description: "The AWS Account ID in which the resource is located.",
		},
	}
}

func commonAwsColumns() []*plugin.Column {

	return []*plugin.Column{
		{
			Name:        "partition",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS partition in which the resource is located (aws, aws-cn, or aws-us-gov).",
		},
		{
			Name:        "region",
			Type:        proto.ColumnType_STRING,
			Transform:   transform.FromConstant("global"),
			Description: "The AWS Region in which the resource is located.",
		},
		{
			Name:        "account_id",
			Type:        proto.ColumnType_STRING,
			Hydrate:     commonColumnsHydrate,
			Description: "The AWS Account ID in which the resource is located.",
			Transform:   transform.FromCamel(),
		},
	}
}

// append the common aws columns for REGIONAL resources onto the column list
func awsRegionalColumns(columns []*plugin.Column) []*plugin.Column {
	return append(columns, commonAwsRegionalColumns()...)
}

// append the common aws columns for GLOBAL resources onto the column list
func awsColumns(columns []*plugin.Column) []*plugin.Column {
	return append(columns, commonAwsColumns()...)
}

func awsS3Columns(columns []*plugin.Column) []*plugin.Column {
	return append(columns, commonS3Columns()...)
}

// struct to store the common column data
type awsCommonColumnData struct {
	Partition, Region, AccountId string
}

// get columns which are returned with all tables: region, partition and account
func getCommonColumns(ctx context.Context, d *plugin.QueryData, h *plugin.HydrateData) (interface{}, error) {
	region := regionFromContext(ctx)
	plugin.Logger(ctx).Trace("getCommonColumns", "region", region)

	// retrieve sts service using cacheable hydrate function
	// note: this returns an interface so we will have to cast to a service when we use it
	stsSvc, err := plugin.ExecuteCacheableHydrate(ctx, d, h, StsService, plugin.ConstantCacheKey("sts"))
	if err != nil {
		return nil, err
	}

	callerIdentity, err := stsSvc.(*sts.STS).GetCallerIdentity(&sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, err
	}
	commonColumnData := &awsCommonColumnData{
		// extract partition from arn
		Partition: strings.Split(*callerIdentity.Arn, ":")[1],
		AccountId: *callerIdentity.Account,
		Region:    region,
	}

	plugin.Logger(ctx).Trace("getCommonColumns: ", "commonColumnData", commonColumnData)

	return commonColumnData, nil
}

func getCommonColumnsCacheKey(ctx context.Context, _ *plugin.QueryData, _ *plugin.HydrateData) (interface{}, error) {
	region := regionFromContext(ctx)

	cacheKey := "commonColumnData" + region
	return cacheKey, nil
}

func regionFromContext(ctx context.Context) string {
	var region string
	if plugin.GetMatrixItem(ctx)[matrixKeyRegion] != nil {
		region = plugin.GetMatrixItem(ctx)[matrixKeyRegion].(string)
	}
	if region == "" {
		region = "global"
	}
	return region
}
