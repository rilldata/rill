package server

import (
	"encoding/base64"
	"time"

	adminv1 "github.com/rilldata/rill/proto/gen/rill/admin/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const _defaultPageSize = 20

func marshalPageToken(val string) string {
	token := &adminv1.StringPageToken{Val: val}
	bytes, err := proto.Marshal(token)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func unmarshalPageToken(reqToken string) (*adminv1.StringPageToken, error) {
	token := &adminv1.StringPageToken{}
	if reqToken != "" {
		in, err := base64.URLEncoding.DecodeString(reqToken)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse request token: %s", err.Error())
		}

		if err := proto.Unmarshal(in, token); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse request token: %s", err.Error())
		}
	}
	return token, nil
}

func marshalStringTimestampPageToken(s string, ts time.Time) string {
	tkn := &adminv1.StringTimestampPageToken{
		Str: s,
		Ts:  timestamppb.New(ts),
	}
	bytes, err := proto.Marshal(tkn)
	if err != nil {
		panic(err)
	}

	return base64.URLEncoding.EncodeToString(bytes)
}

func unmarshalStringTimestampPageToken(tknStr string) (*adminv1.StringTimestampPageToken, error) {
	tkn := &adminv1.StringTimestampPageToken{}
	if tknStr != "" {
		in, err := base64.URLEncoding.DecodeString(tknStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse request token: %s", err.Error())
		}

		if err := proto.Unmarshal(in, tkn); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse request token: %s", err.Error())
		}
	}

	return tkn, nil
}

func validPageSize(pageSize uint32) int {
	if pageSize == 0 {
		return _defaultPageSize
	}
	return int(pageSize)
}
