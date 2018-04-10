package datanode

import (
	"fmt"
	"net"
	"os"
	"sync"

	"github.com/tiglabs/containerfs/logger"
	"github.com/tiglabs/containerfs/proto/dp"
	"github.com/tiglabs/containerfs/proto/vp"
	"github.com/tiglabs/containerfs/utils"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var VolMgrHosts []string

// DataNodeServer ...
type DataNodeServer struct {
	ClientStreamID                 uint64
	C2MReplServerStreamCacheLocker sync.RWMutex
	C2MReplServerStreamCache       map[uint64]*C2MReplServerStream

	M2SReplClientStreamCacheLocker sync.RWMutex
	M2SReplClientStreamCache       map[uint64]*M2SReplClientStream
}

// Start DataNode Server
func StartDataService() {

	lis, err := net.Listen("tcp", DtAddr.Host)
	if err != nil {
		panic(fmt.Sprintf("Failed to listen on:%v", DtAddr.Host))
	}
	s := grpc.NewServer()
	dp.RegisterDataNodeServer(s, &DataNodeServer{M2SReplClientStreamCache: make(map[uint64]*M2SReplClientStream), C2MReplServerStreamCache: make(map[uint64]*C2MReplServerStream)})
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		panic(fmt.Sprintf("Failed to start Serve on:%v", DtAddr.Host))
	}
}

// DataNodeServerAddr ...
type DataNodeServerAddr struct {
	Host string
	Path string
	Flag string
	Log  string
	Tier string
}

var MetaNodePeers []string
var MetaNodeAddr string

var DtAddr DataNodeServerAddr

// Registry DataNode to VolMgr Server
func RegistryToVolMgr() {
	_, conn, err := utils.DialVolMgr(VolMgrHosts)
	if err != nil {
		logger.Error("DataNode[%v]: registryToVolMgr: Dail VolMgr Failed err:%v", DtAddr.Host, err)
		os.Exit(1)
	}
	defer conn.Close()
	vc := vp.NewVolMgrClient(conn)

	var datanodeRegistryReq vp.DataNode
	datanodeRegistryReq.Host = DtAddr.Host
	diskInfo := utils.DiskUsage(DtAddr.Path)
	datanodeRegistryReq.Capacity = int32(float64(diskInfo.All) / float64(1024*1024*1024))
	datanodeRegistryReq.Free = int32(float64(diskInfo.Free) / float64(1024*1024*1024))
	datanodeRegistryReq.Used = int32(float64(diskInfo.Used) / float64(1024*1024*1024))
	datanodeRegistryReq.MountPoint = DtAddr.Path
	datanodeRegistryReq.Tier = DtAddr.Tier
	datanodeRegistryReq.Status = 0

	ack, err := vc.DataNodeRegistry(context.Background(), &datanodeRegistryReq)
	if err != nil {
		logger.Error("DataNode[%v]: register to VolMgr failed! err %v", DtAddr.Host, err)
		os.Exit(1)
	}
	if ack.Ret == 0 {
		logger.Debug("DataNode[%v]: register to VolMgr success!", DtAddr.Host)
	} else if ack.Ret == 3 {
		logger.Debug("DataNode[%v]: register to VolMgr success, already register!", DtAddr.Host)
	} else {
		logger.Error("DataNode[%v]: register to VolMgr failed! ret %v", DtAddr.Host, ack.Ret)
		os.Exit(1)
	}

	return
}

// DatanodeHealthCheck: check datanode if alive
func (s *DataNodeServer) DataNodeHealthCheck(ctx context.Context, in *dp.DataNodeHealthCheckReq) (*dp.DataNodeHealthCheckAck, error) {
	ack := dp.DataNodeHealthCheckAck{}
	f, err := os.OpenFile(DtAddr.Path+"/health", os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		logger.Error("DataNode[%v]: Open datanode check health file error:%v", DtAddr.Host, err)
		ack.Status = 2
	} else {
		_, err = f.WriteString("ok")
		if err != nil {
			logger.Error("DataNode[%v]: Write datanode check health file error:%v", DtAddr.Host, err)
			ack.Status = 2
		}
	}

	diskInfo := utils.DiskUsage(DtAddr.Path)
	ack.Used = int32(float64(diskInfo.Used) / float64(1024*1024*1024))
	ack.Ret = 0
	return &ack, nil
}
