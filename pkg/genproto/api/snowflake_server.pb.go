// Code generated by protoc-gen-yggdrasil-rpc. DO NOT EDIT.

package api

import (
	context "context"
	errors "github.com/imkuqin-zw/yggdrasil/pkg/errors"
	code "google.golang.org/genproto/googleapis/rpc/code"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

type SnowflakeServer interface {
	FetchNext(context.Context, *FetchSnowflakeNextReq) (*UUID, error)
	UnsafeSnowflakeServer
}

type UnsafeSnowflakeServer interface {
	mustEmbedUnimplementedSnowflakeServer()
}

// UnimplementedSnowflakeServer must be embedded to have forward compatible implementations.
type UnimplementedSnowflakeServer struct {
}

func (UnimplementedSnowflakeServer) FetchNext(context.Context, *FetchSnowflakeNextReq) (*UUID, error) {
	return nil, errors.Errorf(code.Code_UNAUTHENTICATED, "method FetchNext not implemented")
}

func (UnimplementedSnowflakeServer) mustEmbedUnimplementedSnowflakeServer() {}