// Code generated by protoc-gen-yggdrasil-error. DO NOT EDIT.

package api

import (
	code "google.golang.org/genproto/googleapis/rpc/code"
)

// This is a compile-time assertion to ensure that this generated file
// is compatible with the yggdrasil package it is being compiled against.

var SeqsvrErrReason_code = map[int32]code.Code{
	0: code.Code(0),
	1: code.Code(9),
	2: code.Code(9),
}

func (r SeqsvrErrReason) Reason() string {
	return SeqsvrErrReason_name[int32(r)]
}

func (r SeqsvrErrReason) Domain() string {
	return "com.github.imkuqin_zw.uuid_generator.api"
}

func (r SeqsvrErrReason) Code() code.Code {
	return SeqsvrErrReason_code[int32(r)]
}