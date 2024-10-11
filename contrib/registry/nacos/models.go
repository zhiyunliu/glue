package nacos

type ServiceCatalogInfo struct {
	NamespaceId      string            `json:"namespaceId"`
	GroupName        string            `json:"groupName"`
	Name             string            `json:"name"`
	ProtectThreshold float64           `json:"protectThreshold"`
	Metadata         map[string]string `json:"metadata"`
}
