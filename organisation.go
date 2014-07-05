
package main 

import (
    "fmt"

    "encoding/xml"
    "net/url"

    // "io"
    // "io/ioutil"
)

type OrganisationReference struct {
    Name    string `xml:"name,attr"`
    Id      string `xml:"id,attr"`
    Href    string `xml:"href,attr"`
}

type Organisations struct {
    Records []*OrganisationReference `xml:"OrgReference"`
}

func (o *Organisations) GetAll (session *vCloudSession, format string, max int) {
    uri := fmt.Sprintf("/api/query?type=organization&format=%v&pageSize=%v", format, max)
    
    r := session.Get(uri)
    defer r.Close()

    _ = xml.NewDecoder(r).Decode(o)
    // xml_decoder.Decode(o)

    // Loop over URLs and reduce the HREFs to URIs
    // We don't need the whole URL
    for k, v := range o.Records {
        u, _ := url.Parse(v.Href)
        o.Records[k].Href = u.Path 
    }
}
