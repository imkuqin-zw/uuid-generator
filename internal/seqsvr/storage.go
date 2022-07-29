package seqsvr

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/imkuqin-zw/uuid-generator/internal/seqsvr/proto"
	"github.com/imkuqin-zw/uuid-generator/pkg/etcdv3"
	"github.com/imkuqin-zw/yggdrasil/pkg/config"
	"github.com/imkuqin-zw/yggdrasil/pkg/types"
	"github.com/imkuqin-zw/yggdrasil/pkg/utils/xnet"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type RegisterEndpoint struct {
	Scheme   string            `json:"scheme,omitempty"`
	Kind     types.ServerKind  `json:"kind,omitempty"`
	Host     string            `json:"host,omitempty"`
	Metadata map[string]string `json:"metadata,omitempty"`
}

type RegisterNodeInfo struct {
	IP        string              `json:"ip,omitempty"`
	Endpoints []*RegisterEndpoint `json:"endpoints,omitempty"`
}

type Storage interface {
	Register(ctx context.Context, node *RegisterNodeInfo) error
	Unregister(ctx context.Context) error
	Heartbeat(ctx context.Context, version uint64) (*proto.Router, error)
	IncrAndGetMax(ctx context.Context, ID uint32, nowVal, step uint64) (uint64, error)
}

type etcdv3Storage struct {
	leaseID          clientv3.LeaseID
	heartbeatTimeout int64
	setID            uint32
	etcdv3Cli        *clientv3.Client
	privateIP        string
	routerKey        string
	routerVersionKey string
	setsKey          string
	nodeKey          string
}

func NewEtcdv3Storage() Storage {
	setID := config.Get("seqsvr.setID").Int(-1)
	if setID == -1 {
		log.Fatalf("seqsvr set ID can not be empty")
		return nil
	}
	privateIP, _ := xnet.Extract("")
	client, err := etcdv3.StdConfig().Build()
	if err != nil {
		log.Fatalf("fault to init etcd client, err: %+v", err)
		return nil
	}
	c := &etcdv3Storage{
		etcdv3Cli: client,
		privateIP: privateIP,
		setID:     uint32(setID),
	}
	c.routerKey = "/router"
	c.routerVersionKey = fmt.Sprintf("%s/version", c.routerKey)
	c.setsKey = fmt.Sprintf("%s/sets", c.routerKey)
	c.nodeKey = fmt.Sprintf("/nodes/set/%d/%s", c.setID, c.privateIP)
	return c
}

func (c *etcdv3Storage) Register(ctx context.Context, node *RegisterNodeInfo) error {
	resp, err := c.etcdv3Cli.Grant(ctx, c.heartbeatTimeout)
	if err != nil {
		return err
	}
	c.leaseID = resp.ID
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	data, _ := json.Marshal(&RegisterNodeInfo{
		Endpoints: node.Endpoints,
	})
	_, err = c.etcdv3Cli.Put(ctx, c.nodeKey, string(data), clientv3.WithLease(c.leaseID))
	if err != nil {
		return err
	}
	return nil
}

func (c *etcdv3Storage) Unregister(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*2)
	defer cancel()
	_, err := c.etcdv3Cli.Revoke(ctx, c.leaseID)
	return err
}

func (c *etcdv3Storage) Heartbeat(ctx context.Context, version uint64) (*proto.Router, error) {
	// 续约节点信息
	_, err := c.etcdv3Cli.KeepAliveOnce(ctx, c.leaseID)
	if err != nil {
		return nil, err
	}
	// 路由更新
	txnRes, err := c.etcdv3Cli.Txn(ctx).
		If(clientv3.Compare(clientv3.Value(c.routerVersionKey), ">", fmt.Sprintf("%d", version))).
		Then(clientv3.OpGet(c.routerKey, clientv3.WithPrefix())).
		Commit()
	if err != nil {
		return nil, err
	}
	if !txnRes.Succeeded {
		return nil, nil
	}
	router := &proto.Router{}
	sets := make(map[string]*proto.Set)
	rangeRes := txnRes.Responses[0].GetResponseRange()
	for _, kv := range rangeRes.Kvs {
		k := string(kv.Key)
		switch {
		case k == c.routerVersionKey:
			version, _ := strconv.ParseUint(string(kv.Value), 10, 64)
			router.Version = version
		case strings.HasPrefix(k, "/router/sets"):
			keys := strings.Split(k, "/")
			set, ok := sets[keys[3]]
			if !ok {
				setKey := strings.Split(keys[3], "_")
				setID, _ := strconv.ParseInt(setKey[0], 10, 64)
				setSize, _ := strconv.ParseInt(setKey[1], 10, 64)
				sectionSize, _ := strconv.ParseInt(setKey[2], 10, 64)
				set = &proto.Set{
					ID:          uint32(setID),
					Size:        uint32(setSize),
					SectionSize: uint32(sectionSize),
				}
				sets[keys[3]] = set
			}
			node := &proto.AllocNode{}
			if err := json.Unmarshal(kv.Value, node); err != nil {
				return nil, err
			}
			node.IP = keys[4]
			set.Nodes = append(set.Nodes, node)
		default:
		}
	}
	router.Sets = make([]*proto.Set, 0, len(sets))
	for _, set := range sets {
		sort.Slice(set.Nodes, func(i, j int) bool {
			return set.Nodes[i].IP < set.Nodes[j].IP
		})
		router.Sets = append(router.Sets, set)
	}
	sort.Slice(router.Sets, func(i, j int) bool {
		return router.Sets[i].ID < router.Sets[j].ID
	})
	return router, nil
}

func (c *etcdv3Storage) IncrAndGetMax(ctx context.Context, ID uint32, nowVal, step uint64) (uint64, error) {
	key := fmt.Sprintf("/sections/%d", ID)
	nextVal := nowVal + step
	txnRes, err := c.etcdv3Cli.Txn(ctx).
		If(
			clientv3.Compare(clientv3.CreateRevision(key), "!=", 0),
			clientv3.Compare(clientv3.Value(key), ">", fmt.Sprintf("%d", nowVal)),
		).
		Then(clientv3.OpGet(key)).
		Else(clientv3.OpPut(key, fmt.Sprintf("%d", nextVal))).
		Commit()
	if err != nil {
		return 0, err
	}
	if txnRes.Succeeded {
		rangeRes := txnRes.Responses[0].GetResponseRange()
		nextVal, _ = strconv.ParseUint(string(rangeRes.Kvs[0].Value), 10, 64)
	}
	return nextVal, nil
}
