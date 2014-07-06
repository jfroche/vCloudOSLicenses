
package vcloudoslicenses 

import (
    "strings"
    "sync"
    "time"
    "strconv"

    "log"
)

type ReportTotal struct {
    VDCs            uint 
    VApps           uint 
    VMs             uint
}

type ReportDocument struct {
    Timestamp       string 
    Year            string 
    Month           string 
    Day             string 
    Organisation    string
    VDC             string 
    VApp            string
    MSWindows       uint 
    RHEL            uint 
    CentOS          uint 
    Ubuntu          uint 
    Unknown         uint

    Totals          ReportTotal
}

type WorkerJob struct {
    Waiter          *sync.WaitGroup
    ResultsChannel  chan <- *ReportDocument
    Organisation    *Organisation
}

func (v *VCloudSession) ReportWorker (job *WorkerJob) {
    for _, link := range job.Organisation.Links {
        if link.Type == "application/vnd.vmware.vcloud.vdc+xml" {
            v.Counters.VDCs++

            vdc := &VDC{}
            vdc.Get(v, link)

            for _, entity := range vdc.ResourceEntities.ResourceEntity {
                if entity.Type == "application/vnd.vmware.vcloud.vApp+xml" {
                    v.Counters.VApps++

                    vapp := &VDCVApp{}
                    vapp.Get(v, entity)

                    now := time.Now()
                    report := &ReportDocument{
                        Timestamp:      now.String(),
                        Year:           strconv.Itoa(now.Year()),
                        Month:          now.Month().String(),
                        Day:            strconv.Itoa(now.Day()),
                        Organisation:   job.Organisation.Name,
                        VDC:            vdc.Name,
                        VApp:           vapp.Name,
                        MSWindows:      0,
                        RHEL:           0,
                        CentOS:         0,
                        Ubuntu:         0,
                        Unknown:        0,
                    }

                    for _, vm := range vapp.VMs.VM {
                        v.Counters.VMs++

                        if strings.Contains(vm.OperatingSystemSection.OSType, "windows") {
                            report.MSWindows++
                        } else if strings.Contains(vm.OperatingSystemSection.OSType, "rhel") {
                            report.RHEL++
                        } else if strings.Contains(vm.OperatingSystemSection.OSType, "centos") {
                            report.CentOS++
                        } else if strings.Contains(vm.OperatingSystemSection.OSType, "ubuntu") {
                            report.Ubuntu++
                        } else {
                            report.Unknown++
                        }
                    }

                    job.ResultsChannel <- report
                }
            }
        }
    }

    log.Printf("VDCs: %d, vApps: %d, and VMs: %d", vdc_counter, vapp_counter, vm_counter)
    job.Waiter.Done() 
}

func (v *VCloudSession) LicenseReport (max_organisations, max_pages int) (report []*ReportDocument) {
    var reports []*ReportDocument
    
    if max_organisations <= 0 {
        max_organisations = 10
    }

    if max_pages <= 0 {
        max_pages = 1
    }

    waiter  := &sync.WaitGroup{}
    results := make(chan *ReportDocument)

    waiter.Add(max_organisations)

    orgs, _ := FindOrganisations(v, max_organisations, max_pages)
    v.Counters.Orgs = len(orgs)

    for _, org := range orgs {
        job := &WorkerJob{
            Waiter:         waiter,
            ResultsChannel: results,
            Organisation:   org,
        }

        go v.ReportWorker(job)
    }

    go func() {
        waiter.Wait()
        close(results)
    }()

    for report := range results {
        reports = append(reports, report)
    }
  
    return reports 
}