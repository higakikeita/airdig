package aws

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
)

// getNameTag は EC2 タグから Name タグの値を取得
func getNameTag(tags []types.Tag) string {
	for _, tag := range tags {
		if tag.Key != nil && *tag.Key == "Name" && tag.Value != nil {
			return *tag.Value
		}
	}
	return ""
}

// convertTags は EC2 タグを map[string]string に変換
func convertTags(tags []types.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// convertRDSTags は RDS タグを map[string]string に変換
func convertRDSTags(tags []rdstypes.Tag) map[string]string {
	result := make(map[string]string)
	for _, tag := range tags {
		if tag.Key != nil && tag.Value != nil {
			result[*tag.Key] = *tag.Value
		}
	}
	return result
}

// getStringPtr は *string から string を取得（nil の場合は空文字）
func getStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// getInt32Ptr は *int32 から int を取得（nil の場合は 0）
func getInt32Ptr(i *int32) int {
	if i == nil {
		return 0
	}
	return int(*i)
}

// getTimePtr は *time.Time から time.Time を取得（nil の場合は現在時刻）
func getTimePtr(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}
