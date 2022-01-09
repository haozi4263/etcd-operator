package controllers

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// 定义执行动作接口
type Action interface {
	Execute(ctx context.Context) error
}

type PatchStatus struct {
	client   client.Client
	original runtime.Object
	new      runtime.Object
}

func (s *PatchStatus) Execute(ctx context.Context) error {
	if reflect.DeepEqual(s.original, s.new) {
		return nil
	}
	// 更新状态
	if err := s.client.Status().Patch(ctx, s.new, client.MergeFrom(s.original)); err != nil {
		return fmt.Errorf("patching status error: %s", err)
	}
	return nil
}

// 创建一个新的资源对象
type CreateObject struct {
	client client.Client
	obj    runtime.Object
}

func (o *CreateObject) Execute(ctx context.Context) error {
	if err := o.client.Create(ctx, o.obj); err != nil {
		return fmt.Errorf("create object error: %s", err)
	}

	return nil
}

// updateObject
