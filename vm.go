package vcloudoslicenses

import (
	"encoding/xml"
	"fmt"
	"log"
	"strings"
)

type VMRecord struct {
	XMLName string `xml:"VMRecord"`
	Name    string `xml:"name,attr"`
	Href    string `xml:"href,attr"`
}

type TypedValue struct {
	XMLName string `xml:"TypedValue"`

	Value string `xml:"Value"`
}

type Metadata struct {
	XMLName string `xml:"MetadataEntry"`

	Key    string        `xml:"Key"`
	Values []*TypedValue `xml:"TypedValue"`
}

type VmMetadata struct {
	XMLName string `xml:"Metadata"`

	Metadatas []*Metadata `xml:"MetadataEntry,omitempty"`
}

type VMQueryResultsRecords struct {
	XMLName string      `xml:"QueryResultRecords"`
	Records []*VMRecord `xml:"VMRecord"`
}

type NetworkConnection struct {
	XMLName string `xml:"NetworkConnection"`

	NetworkConnectionIndex uint   `xml:"NetworkConnectionIndex"`
	IpAddress              string `xml:"IpAddress"`
}

type VmNets struct {
	XMLName            string               `xml:"NetworkConnectionSection"`
	NetworkConnections []*NetworkConnection `xml:"NetworkConnection"`
}

type VmOS struct {
	XMLName string `xml:"OperatingSystemSection"`

	Id     string `xml:"id,attr"`
	Type   string `xml:"type,attr"`
	Href   string `xml:"href,attr"`
	OSType string `xml:"osType,attr"`
}

type VAppVm struct {
	XMLName string `xml:"Vm"`

	Deployed string `xml:"deployed,attr"`
	Status   string `xml:"status,attr"`
	Name     string `xml:"name,attr"`
	Id       string `xml:"id,attr"`
	Type     string `xml:"type,attr"`
	Href     string `xml:"href,attr"`

	OperatingSystemSection *VmOS `xml:"OperatingSystemSection"`

	NetworkConnectionSection *VmNets `xml:"NetworkConnectionSection"`

	Metadata map[string][]string
}

func (v *VAppVm) GetMetadata(session *VCloudSession) map[string][]string {
	vmid := strings.Replace(v.Id, "urn:vcloud:vm:", "", 1)
	metadata_href := fmt.Sprint("/api/vApp/vm-", vmid, "/metadata")
	resp, err := session.Get(metadata_href)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	vmm := &VmMetadata{}

	_ = xml.NewDecoder(resp.Body).Decode(vmm)

	results := make(map[string][]string)
	for _, metadata := range vmm.Metadatas {
		for _, value := range metadata.Values {
			results[metadata.Key] = append(results[metadata.Key], value.Value)
		}
	}
	return results
}
