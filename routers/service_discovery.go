package routers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Database struct {
	User     string
	Password string
	Host     string
	Name     string
}

type ZoneInfo struct {
	Datacenter string `uri:"dc" binding:"required" json:"datacenter"`
	Namespace  string `uri:"ns" binding:"required" json:"namespace"`
}

type DeServiceInstance struct {
	Datacenter string `uri:"dc" binding:"required" json:"datacenter"`
	Namespace  string `uri:"ns" binding:"required" json:"namespace"`
	InstanceId string `uri:"instanceId" json:"instanceId" binding:"required"`
}

type ServiceInstance struct {
	InstanceId string                 `form:"instanceId" json:"instanceId" binding:"required"`
	ServiceId  string                 `form:"serviceId" json:"serviceId" binding:"required"`
	Host       string                 `form:"host" json:"host" binding:"required"`
	Port       int                    `form:"port" json:"port" binding:"required"`
	Metadata   map[string]interface{} `form:"metadata" json:"metadata" binding:"required"`
}

/**
 {
	"Datacenter": {
		"Namespace": {
			"serverName": {
				"instanceId": {
					"instanceId": "radljfkei",
					"serviceId": "user_server",
					"host": "127.0.0.1",
					"port": 8080,
					"metadata": {
						"abc": "dbc"
					}
				}
			}
		}
	}
}
*/

var globalServiceInstanceMap = make(map[string]map[string]map[string]map[string]ServiceInstance)

func init() {
	globalServiceInstanceMap["default"] = make(map[string]map[string]map[string]ServiceInstance)
	globalServiceInstanceMap["default"]["default"] = make(map[string]map[string]ServiceInstance)
}

func DeregisterServiceInstance(c *gin.Context) {
	var deServiceInstance DeServiceInstance

	if err := c.ShouldBindUri(&deServiceInstance); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	if datacenterMap, ok := globalServiceInstanceMap[deServiceInstance.Datacenter]; ok {
		if namespaceMap, ok := datacenterMap[deServiceInstance.Namespace]; ok {
			for _, serviceInfo := range namespaceMap {
				delete(serviceInfo, deServiceInstance.InstanceId)
			}
		}
	}
	c.AbortWithStatus(200)
}

func GetServiceInstance(c *gin.Context) {
	c.JSON(200, globalServiceInstanceMap)
}

func RegisterServiceInstance(c *gin.Context) {

	var zoneInfo ZoneInfo
	var serviceInstance ServiceInstance

	if err := c.ShouldBindUri(&zoneInfo); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}

	if err := c.ShouldBindJSON(&serviceInstance); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	datacenterMap, ok := globalServiceInstanceMap[zoneInfo.Datacenter]
	if !ok {
		globalServiceInstanceMap[zoneInfo.Datacenter] = make(map[string]map[string]map[string]ServiceInstance)
		datacenterMap = globalServiceInstanceMap[zoneInfo.Datacenter]
	}

	namespaceMap, ok := datacenterMap[zoneInfo.Namespace]
	if !ok {
		datacenterMap[zoneInfo.Namespace] = make(map[string]map[string]ServiceInstance)
		namespaceMap = datacenterMap[zoneInfo.Namespace]
	}

	serviceInstanceNode, ok := namespaceMap[serviceInstance.ServiceId]
	if !ok {
		namespaceMap[serviceInstance.ServiceId] = make(map[string]ServiceInstance)
		serviceInstanceNode = namespaceMap[serviceInstance.ServiceId]
	}

	serviceInstanceNode[serviceInstance.InstanceId] = serviceInstance
	c.JSON(200, globalServiceInstanceMap)
}