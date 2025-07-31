package core

import (
	"context"
	"errors"
	"fmt"
	"github.com/avast/retry-go"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"time"
	"trojan-panel/model/constant"
)

func newGrpcInstance(token string, ip string, grpcPort uint, timeout time.Duration) (conn *grpc.ClientConn, ctx context.Context, clo func(), err error) {
	tokenParam := TokenValidateParam{
		Token: token,
	}

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithPerRPCCredentials(&tokenParam),
	}
	conn, err = grpc.Dial(fmt.Sprintf("%s:%d", ip, grpcPort),
		opts...)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	clo = func() {
		cancel()
		if conn != nil {
			conn.Close()
		}
	}
	if err != nil {
		logrus.Errorf("gRPC instance init err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		err = errors.New(constant.GrpcError)
	}
	return
}

func AddNode(token string, ip string, grpcPort uint, nodeAddDto *NodeAddDto) error {
	if err := retry.Do(func() error {
		conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
		defer clo()
		if err != nil {
			return err
		}
		client := NewApiNodeServiceClient(conn)
		send, err := client.AddNode(ctx, nodeAddDto)
		if err != nil {
			return err
		}
		if send.Success {
			return nil
		}
		return errors.New(send.Msg)
	}, []retry.Option{
		retry.Delay(8 * time.Second),
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	}...); err != nil {
		logrus.Errorf("api AddNode err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		return errors.New(constant.GrpcError)
	}
	return nil
}

func RemoveNode(token string, ip string, grpcPort uint, nodeRemoveDto *NodeRemoveDto) error {
	if err := retry.Do(func() error {
		conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
		defer clo()
		if err != nil {
			return err
		}
		client := NewApiNodeServiceClient(conn)
		send, err := client.RemoveNode(ctx, nodeRemoveDto)
		if err != nil {
			return err
		}
		if send.Success {
			return nil
		}
		return errors.New(send.Msg)
	}, []retry.Option{
		retry.Delay(8 * time.Second),
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	}...); err != nil {
		logrus.Errorf("api RemoveNode err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		return errors.New(constant.GrpcError)
	}
	return nil
}

func RemoveAccount(token string, ip string, grpcPort uint, accountRemoveDto *AccountRemoveDto) error {
	if err := retry.Do(func() error {
		conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
		defer clo()
		if err != nil {
			return err
		}
		client := NewApiAccountServiceClient(conn)
		send, err := client.RemoveAccount(ctx, accountRemoveDto)
		if err != nil {
			return err
		}
		if send.Success {
			return nil
		}
		return errors.New(send.Msg)
	}, []retry.Option{
		retry.Delay(8 * time.Second),
		retry.Attempts(3),
		retry.LastErrorOnly(true),
	}...); err != nil {
		logrus.Errorf("api RemoveAccount err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		return errors.New(constant.GrpcError)
	}
	return nil
}

// GetNodeState 查询节点状态
func GetNodeState(token string, ip string, grpcPort uint, nodeTypeId uint, port uint) (*NodeStateVo, error) {
	conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
	defer clo()
	if err != nil {
		return nil, err
	}
	client := NewApiStateServiceClient(conn)
	nodeStateDto := NodeStateDto{
		NodeTypeId: uint64(nodeTypeId),
		Port:       uint64(port),
	}
	send, err := client.GetNodeState(ctx, &nodeStateDto)
	if err != nil {
		logrus.Errorf("gRPC GetNodeState err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		return nil, errors.New(constant.GrpcError)
	}
	if send.Success {
		var nodeStateVo NodeStateVo
		if err = anypb.UnmarshalTo(send.Data, &nodeStateVo, proto.UnmarshalOptions{}); err != nil {
			logrus.Errorf("api GetNodeState UnmarshalTo err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
			return nil, errors.New(constant.GrpcError)
		}
		return &nodeStateVo, nil
	}
	logrus.Errorf("api GetNodeState err ip: %s grpcPort: %d err: %s", ip, grpcPort, send.Msg)
	return nil, errors.New(constant.GrpcError)
}

// GetNodeServerState 查询服务器状态
func GetNodeServerState(token string, ip string, grpcPort uint) (*NodeServerStateVo, error) {
	conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
	defer clo()
	if err != nil {
		return nil, err
	}
	client := NewApiStateServiceClient(conn)
	nodeServerStateDto := NodeServerStateDto{}
	send, err := client.GetNodeServerState(ctx, &nodeServerStateDto)
	if err != nil {
		logrus.Errorf("gRPC GetNodeServerState err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
		return nil, errors.New(constant.GrpcError)
	}
	if send.Success {
		var nodeServerStateVo NodeServerStateVo
		if err = anypb.UnmarshalTo(send.Data, &nodeServerStateVo, proto.UnmarshalOptions{}); err != nil {
			logrus.Errorf("api GetNodeServerState UnmarshalTo err ip: %s grpcPort: %d err: %v", ip, grpcPort, err)
			return nil, errors.New(constant.GrpcError)
		}
		return &nodeServerStateVo, nil
	}
	logrus.Errorf("api GetNodeServerState err ip: %s grpcPort: %d err: %s", ip, grpcPort, send.Msg)
	return nil, errors.New(constant.GrpcError)
}

func GetNodeServerInfo(token string, ip string, grpcPort uint) (*NodeServerInfoVo, error) {
	conn, ctx, clo, err := newGrpcInstance(token, ip, grpcPort, 4*time.Second)
	defer clo()
	if err != nil {
		return nil, err
	}
	client := NewApiNodeServerServiceClient(conn)
	nodeServerInfoDto := NodeServerInfoDto{}
	send, err := client.GetNodeServerInfo(ctx, &nodeServerInfoDto)
	if err != nil {
		logrus.Errorf("gRPC GetNodeServerInfo err ip: %s port: %d err: %v", ip, grpcPort, err)
		return nil, errors.New(constant.GrpcError)
	}
	if send.Success {
		var nodeServerInfoVo NodeServerInfoVo
		if err = anypb.UnmarshalTo(send.Data, &nodeServerInfoVo, proto.UnmarshalOptions{}); err != nil {
			logrus.Errorf("api GetNodeServerInfo UnmarshalTo err ip: %s port: %d err: %v", ip, grpcPort, err)
			return nil, errors.New(constant.GrpcError)
		}
		return &nodeServerInfoVo, nil
	}
	logrus.Errorf("api GetNodeServerInfo err ip: %s grpcPort: %d err: %s", ip, grpcPort, send.Msg)
	return nil, errors.New(constant.GrpcError)
}
