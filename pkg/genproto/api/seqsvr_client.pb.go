// Code generated by protoc-gen-yggdrasil-rpc. DO NOT EDIT.

package api

import (
	context "context"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

type AllocClient interface {
	FetchNext(context.Context, *FetchSeqNextReq) (*UUID, error)
}
