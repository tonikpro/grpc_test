package service

import (
	"context"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/tonikpro/grpc_test/test/mocks/repository"
)

func Test_a2billing_GetAgentIdsByParentAgentID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := repository.NewMockA2billingRepository(ctrl)
	m.EXPECT().GetAgentIdsByParentAgentID(int32(1)).Return([]int32{2, 3, 4}, nil)
	m.EXPECT().GetAgentIdsByParentAgentID(int32(5)).Return([]int32{6, 7}, nil)
	m.EXPECT().GetAgentIdsByParentAgentID(int32(6)).Return([]int32{}, nil)
	svc := NewService(m)
	type args struct {
		ctx context.Context
		id  int32
	}
	tests := []struct {
		name       string
		s          A2billing
		args       args
		wantResult []int32
		wantErr    bool
	}{
		{
			"agent_id=1",
			svc,
			args{
				context.Background(),
				1,
			},
			[]int32{2, 3, 4},
			false,
		},
		{
			"agent_id=5",
			svc,
			args{
				context.Background(),
				5,
			},
			[]int32{6, 7},
			false,
		},
		{
			"agent_id=6",
			svc,
			args{
				context.Background(),
				6,
			},
			[]int32{},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResult, err := tt.s.GetAgentIdsByParentAgentID(tt.args.ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("a2billing.GetAgentIdsByParentAgentID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("a2billing.GetAgentIdsByParentAgentID() = %v, want %v", gotResult, tt.wantResult)
			}
		})
	}
}
