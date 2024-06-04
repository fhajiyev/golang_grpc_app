package repo

import (
	"context"

	rewardsvc "github.com/Buzzvil/buzzapis/go/reward"
	"google.golang.org/grpc"
)

// RewardServiceClient interface defintion
type RewardServiceClient interface {
	CheckRewardStatus(ctx context.Context, in *rewardsvc.CheckRewardStatusRequest, opts ...grpc.CallOption) (*rewardsvc.CheckRewardStatusResponse, error)
	IssueRewards(ctx context.Context, in *rewardsvc.IssueRewardsRequest, opts ...grpc.CallOption) (*rewardsvc.IssueRewardsResponse, error)
}
