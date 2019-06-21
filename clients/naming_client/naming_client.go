package naming_client

import (
	"github.com/nacos-group/nacos-sdk-go/clients/nacos_client"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/model"
	"github.com/nacos-group/nacos-sdk-go/utils"
	"github.com/nacos-group/nacos-sdk-go/vo"
	"strings"
)

type NamingClient struct {
	nacos_client.INacosClient
	hostReactor  HostReactor
	serviceProxy NamingProxy
	subCallback  SubscribeCallback
	beatReactor  BeatReactor
}

func NewNamingClient(nc nacos_client.INacosClient) (NamingClient, error) {
	naming := NamingClient{}
	clientConfig, err :=
		nc.GetClientConfig()
	if err != nil {
		return naming, err
	}
	serverConfig, err := nc.GetServerConfig()
	if err != nil {
		return naming, err
	}
	httpAgent, err := nc.GetHttpAgent()
	if err != nil {
		return naming, err
	}
	naming.subCallback = NewSubscribeCallback()
	naming.serviceProxy = NewNamingProxy(clientConfig, serverConfig, httpAgent)
	naming.hostReactor = NewHostReactor(naming.serviceProxy, clientConfig.CacheDir, clientConfig.UpdateThreadNum, clientConfig.NotLoadCacheAtStart, naming.subCallback)
	naming.beatReactor = NewBeatReactor(naming.serviceProxy, clientConfig.BeatInterval)
	return naming, nil
}

// 注册服务实例
func (sc *NamingClient) RegisterServiceInstance(param vo.RegisterServiceInstanceParam) (bool, error) {
	if param.GroupName == "" {
		param.GroupName = constant.DEFAULT_GROUP
	}
	instance := model.ServiceInstance{
		Ip:          param.Ip,
		Port:        param.Port,
		Metadata:    param.Metadata,
		ClusterName: param.ClusterName,
		Healthy:     param.Healthy,
		Enable:      param.Enable,
		Weight:      param.Weight,
		Ephemeral:   true,
	}
	beatInfo := model.BeatInfo{
		Ip:          param.Ip,
		Port:        param.Port,
		Metadata:    param.Metadata,
		ServiceName: utils.GetGroupName(param.ServiceName, param.GroupName),
		Cluster:     param.ClusterName,
		Weight:      param.Weight,
	}
	_, err := sc.serviceProxy.RegisterService(utils.GetGroupName(param.ServiceName, param.GroupName), param.GroupName, instance)
	if err != nil {
		return false, err
	}
	sc.beatReactor.AddBeatInfo(utils.GetGroupName(param.ServiceName, param.GroupName), beatInfo)
	return true, nil

}

// 注销服务实例
func (sc *NamingClient) LogoutServiceInstance(param vo.LogoutServiceInstanceParam) (bool, error) {
	if param.GroupName == "" {
		param.GroupName = constant.DEFAULT_GROUP
	}
	_, err := sc.serviceProxy.DeristerService(utils.GetGroupName(param.ServiceName, param.GroupName), param.Ip, param.Port, param.Cluster, true)
	if err != nil {
		return false, err
	}
	sc.beatReactor.RemoveBeatInfo(utils.GetGroupName(param.ServiceName, param.GroupName), param.Ip, param.Port)
	return true, nil
}

// 获取服务列表
func (sc *NamingClient) GetService(param vo.GetServiceParam) (model.Service, error) {
	if param.GroupName == "" {
		param.GroupName = constant.DEFAULT_GROUP
	}
	service := sc.hostReactor.GetServiceInfo(utils.GetGroupName(param.ServiceName, param.GroupName), strings.Join(param.Clusters, ","))
	return service, nil
}

// 获取服务某个实例
func (sc *NamingClient) GetServiceInstance(param vo.GetServiceInstanceParam) (model.ServiceInstance, error) {
	return model.ServiceInstance{}, nil
}

// 获取service的基本信息
func (sc *NamingClient) GetServiceDetail(param vo.GetServiceDetailParam) (model.ServiceDetail, error) {
	return model.ServiceDetail{}, nil
}

// 服务监听
func (sc *NamingClient) Subscribe(param *vo.SubscribeParam) error {
	if param.GroupName == "" {
		param.GroupName = constant.DEFAULT_GROUP
	}
	serviceParam := vo.GetServiceParam{
		ServiceName: param.ServiceName,
		GroupName:   param.GroupName,
		Clusters:    param.Clusters,
	}

	sc.subCallback.AddCallbackFuncs(utils.GetGroupName(param.ServiceName, param.GroupName), strings.Join(param.Clusters, ","), &param.SubscribeCallback)
	_, err := sc.GetService(serviceParam)
	if err != nil {
		return err
	}
	return nil
}

//取消服务监听
func (sc *NamingClient) Unsubscribe(param *vo.SubscribeParam) error {
	sc.subCallback.RemoveCallbackFuncs(utils.GetGroupName(param.ServiceName, param.GroupName), strings.Join(param.Clusters, ","), &param.SubscribeCallback)
	return nil
}