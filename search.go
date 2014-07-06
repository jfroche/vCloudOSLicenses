
package vcloudoslicenses

import (
    "log"
    "fmt"

    "encoding/xml"
    "net/url"
    "errors"
)

type OrgReference struct {
    XMLName string `xml:"OrgReference"`
    Name    string `xml:"name,attr"`
    Id      string `xml:"id,attr"`
    Href    string `xml:"href,attr"`
}

type OrgReferences struct {
    XMLName     string  `xml:"OrgReferences"`
    Records     []*OrgReference `xml:"OrgReference"`
}

func FindOrganisations (session *VCloudSession, max_page_size, max_pages int) (Orgs []*Organisation, err error) {

    if max_page_size <= 0 {
        max_page_size = 1
    }

    if max_pages <= 0 {
        max_pages = 1
    }

    log.Print("Looking for Orgs...")

    for i := 1; i <= max_pages; i++ {
    	page := &OrgReferences{}

        uri := fmt.Sprintf("/api/query?type=organization&format=references&pageSize=%v&page=%v", max_page_size, i)

        log.Printf("Page: %d URI: %s", i, uri)

        r := session.Get(uri)
        defer r.Body.Close()

        if r.StatusCode == 400 {
            break 
        }

        _ = xml.NewDecoder(r.Body).Decode(page)
     
        for _, v := range page.Records {
            u, _ := url.Parse(v.Href)

            log.Printf("Looking for: %s", u.Path)

            r := session.Get(u.Path)
            defer r.Body.Close()

            if r.StatusCode != 200 {
            	continue 
            }

            new_org := &Organisation{}
            _ = xml.NewDecoder(r.Body).Decode(new_org)
            Orgs = append(Orgs, new_org)
        }
    }

    if len(Orgs) <= 0 {
    	return Orgs, errors.New("No organisations returned.")
    } else {
    	return Orgs, nil
    }
}