
package vcloudoslicenses 

import (
    "strings"
    "sync"
    "time"
    "strconv"
    "encoding/xml"

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
    VApp            *AdminVAppRecord
}

type VAppWorkerJob struct {
    Waiter          *sync.WaitGroup
    ResultsChannel  chan <- *ReportDocument
    VApp            *AdminVAppRecord
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

    job.Waiter.Done() 
}

func (v *VCloudSession) VAppReportWorker (job *WorkerJob) {

    if job.VApp.Href == "/api/vApp/vapp-918f87d0-5c7c-4b15-b30f-583692623f36" {
        log.Printf("Here I am! %s", job.VApp.Name)
    }

    vdc := &VDCVApp{}

    log.Printf("I am %s, and I'm about to request: %s%s.", job.VApp.Name, v.Host, job.VApp.Href)

    r := v.Get(job.VApp.Href)
    defer r.Body.Close()

    _ = xml.NewDecoder(r.Body).Decode(vdc)

    now := time.Now()
    report := &ReportDocument{
        Timestamp:      now.String(),
        Year:           strconv.Itoa(now.Year()),
        Month:          now.Month().String(),
        Day:            strconv.Itoa(now.Day()),
        Organisation:   job.VApp.Org,
        VDC:            job.VApp.VDC,
        VApp:           job.VApp.Name,
        MSWindows:      0,
        RHEL:           0,
        CentOS:         0,
        Ubuntu:         0,
        Unknown:        0,
    }

    for _, vm := range vdc.VMs.VM {
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
    job.Waiter.Done() 
}

// func (v *VCloudSession) VAppReportWorker (jobs <- chan *VAppWorkerJob) {

//     for job := range jobs {
//         log.Printf("Job: %+v", job)

//         vdc := &VDCVApp{}

//         r := v.Get(job.VApp.Href)
//         defer r.Body.Close()

//         if r.StatusCode != 200 {
//             continue 
//         }

//         _ = xml.NewDecoder(r.Body).Decode(vdc)

//         now := time.Now()
//         report := &ReportDocument{
//             Timestamp:      now.String(),
//             Year:           strconv.Itoa(now.Year()),
//             Month:          now.Month().String(),
//             Day:            strconv.Itoa(now.Day()),
//             Organisation:   job.VApp.OwnerName,
//             VDC:            job.VApp.VDCName,
//             VApp:           job.VApp.Name,
//             MSWindows:      0,
//             RHEL:           0,
//             CentOS:         0,
//             Ubuntu:         0,
//             Unknown:        0,
//         }
 
//         for _, vm := range vdc.VMs.VM {
//             v.Counters.VMs++

//             if strings.Contains(vm.OperatingSystemSection.OSType, "windows") {
//                 report.MSWindows++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "rhel") {
//                 report.RHEL++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "centos") {
//                 report.CentOS++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "ubuntu") {
//                 report.Ubuntu++
//             } else {
//                 report.Unknown++
//             }
//         }

//         log.Printf("Report: %+v", report)

//         job.ResultsChannel <- report
//     }

//     log.Print("Worker finished...")
// }

func (v *VCloudSession) VAppReport (max_vapps, max_pages int) (reports []*ReportDocument) {
    // var reports []*ReportDocument
    
    if max_vapps <= 0 {
        max_vapps = 10
    }

    if max_pages <= 0 {
        max_pages = 1
    }

    waiter  := &sync.WaitGroup{}
    results := make(chan *ReportDocument)
    vapps, _ := v.FindVApps(max_vapps, max_pages)
    // v.Counters.Orgs = len(orgs)

    waiter.Add(v.Counters.VApps)

    for _, vapp := range vapps {
        job := &WorkerJob{
            Waiter:         waiter,
            ResultsChannel: results,
            VApp:           vapp,
        }

        go v.VAppReportWorker(job)
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

// func (v *VCloudSession) VAppReport (max_vapps, max_pages int) (reports []*ReportDocument) {
//     vapps, _ := v.FindVApps(max_vapps, max_pages)
//     for _, vapp := range vapps {
//         vdc := &VDCVApp{}

//         r := v.Get(vapp.Href)
//         defer r.Body.Close()

//         if r.StatusCode != 200 {
//             continue 
//         }

//         _ = xml.NewDecoder(r.Body).Decode(vdc)

//         now := time.Now()
//         report := &ReportDocument{
//             Timestamp:      now.String(),
//             Year:           strconv.Itoa(now.Year()),
//             Month:          now.Month().String(),
//             Day:            strconv.Itoa(now.Day()),
//             Organisation:   vapp.OwnerName,
//             VDC:            vapp.VDCName,
//             VApp:           vapp.Name,
//             MSWindows:      0,
//             RHEL:           0,
//             CentOS:         0,
//             Ubuntu:         0,
//             Unknown:        0,
//         }
 
//         for _, vm := range vdc.VMs.VM {
//             v.Counters.VMs++

//             if strings.Contains(vm.OperatingSystemSection.OSType, "windows") {
//                 report.MSWindows++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "rhel") {
//                 report.RHEL++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "centos") {
//                 report.CentOS++
//             } else if strings.Contains(vm.OperatingSystemSection.OSType, "ubuntu") {
//                 report.Ubuntu++
//             } else {
//                 report.Unknown++
//             }
//         }

//         log.Printf("Report: %+v", report) 
//         reports = append(reports, report)
//     }
  
//     return reports 
// }

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

    orgs, _ := v.FindOrganisations(max_organisations, max_pages)
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