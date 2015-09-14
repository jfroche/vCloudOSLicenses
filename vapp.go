package vcloudoslicenses

import (
	"encoding/xml"
	"net/url"
)

type AdminVAppRecord struct {
	XMLName string `xml:"AdminVAppRecord"`

	VDCName            string `xml:"vdcName,attr"`
	VDC                string `xml:"vdc,attr"`
	StorageKB          string `xml:"storageKB,attr"`
	Status             string `xml:"status,attr"`
	OwnerName          string `xml:"ownerName,attr"`
	Org                string `xml:"org,attr"`
	NumberOfVMs        uint   `xml:"numberOfVMs,attr"`
	Name               string `xml:"name,attr"`
	MemoryAllocationMB string `xml:"memoryAllocationMB,attr"`
	IsExpired          string `xml:"isExpired,attr"`
	IsEnabled          string `xml:"isEnabled,attr"`
	IsDeployed         string `xml:"isDeployed,attr"`
	CreationDate       string `xml:"creationDate,attr"`
	Href               string `xml:"href,attr"`
	NumberOfCpus       string `xml:"numberOfCpus,attr"`
}

type VAppQueryResultsRecords struct {
	XMLName string             `xml:"QueryResultRecords"`
	Records []*AdminVAppRecord `xml:"AdminVAppRecord"`
}

type VAppChildren struct {
	XMLName string    `xml:"Children"`
	VM      []*VAppVm `xml:"Vm"`
}

type VDCVApp struct {
	XMLName string `xml:"VApp"`

	Deployed string `xml:"deployed,attr"`
	Status   string `xml:"status,attr"`
	Name     string `xml:"name,attr"`
	Id       string `xml:"id,attr"`
	Type     string `xml:"type,attr"`
	Href     string `xml:"href,attr"`

	VMs VAppChildren `xml:"Children"`
}

func (a *VDCVApp) Get(session *VCloudSession, vdc *VdcResourceEntity) {
	r, _ := session.Get(vdc.Href)

	_ = xml.NewDecoder(r.Body).Decode(a)

	for k1, v1 := range a.VMs.VM {
		u, _ := url.Parse(v1.Href)
		a.VMs.VM[k1].Href = u.Path

		u, _ = url.Parse(v1.OperatingSystemSection.Href)
		a.VMs.VM[k1].OperatingSystemSection.Href = u.Path
	}
}
