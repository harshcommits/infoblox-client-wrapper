package raw

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	raw_records "dnsclient-poc/raw/records"

	ibclient "github.com/infobloxopen/infoblox-go-client/v2"
)

type DNSClient interface {
	GetManagedZones(ctx context.Context) (map[string]string, error)
	CreateOrUpdateRecordSet(ctx context.Context, view, name, zone string, record_type string, values []string, ttl int64) error
	DeleteRecordSet(ctx context.Context, zone, name, recordType string) error
	// NewRecord(name string, view string, zone string, value string, ttl int64, record_type string) raw_records.Record
	// CreateRecord(r raw_records.Record, zone string) error
	GetRecordSet(name bool, record_type string, zone string) (RecordSet, error)
}

type dnsClient struct {
	client ibclient.IBConnector
}

type RecordSet []raw_records.Base_Record

type InfobloxConfig struct {
	Host            *string `json:"host,omitempty"`
	Port            *int    `json:"port,omitempty"`
	SSLVerify       *bool   `json:"sslVerify,omitempty"`
	Version         *string `json:"version,omitempty"`
	View            *string `json:"view,omitempty"`
	PoolConnections *int    `json:"httpPoolConnections,omitempty"`
	RequestTimeout  *int    `json:"httpRequestTimeout,omitempty"`
	CaCert          *string `json:"caCert,omitempty"`
	MaxResults      int     `json:"maxResults,omitempty"`
	ProxyURL        *string `json:"proxyUrl,omitempty"`
}

func assignDefaultValues() (InfobloxConfig, error) {

	port := 443
	view := "default"
	poolConnections := 10
	requestTimeout := 60
	version := "2.10"

	return InfobloxConfig{
		Port:            &port,
		View:            &view,
		PoolConnections: &poolConnections,
		RequestTimeout:  &requestTimeout,
		Version:         &version,
	}, nil

}

// NewDNSClient creates a new dns client based on the Infoblox config provided
func NewDNSClient(ctx context.Context, username string, password string, host string) (DNSClient, error) {

	// infobloxConfig := &InfobloxConfig{}

	// set default values
	infobloxConfig, err := assignDefaultValues()
	if err != nil {
		fmt.Println(err)
	}

	// define hostConfig
	hostConfig := ibclient.HostConfig{
		Host:     host,
		Port:     strconv.Itoa(*infobloxConfig.Port),
		Version:  *infobloxConfig.Version,
		Username: username,
		Password: password,
	}

	verify := "false"
	if infobloxConfig.SSLVerify != nil {
		verify = strconv.FormatBool(*infobloxConfig.SSLVerify)
	}

	if infobloxConfig.CaCert != nil && verify == "true" {
		tmpfile, err := ioutil.TempFile("", "cacert")
		if err != nil {
			return nil, fmt.Errorf("cannot create temporary file for cacert: %w", err)
		}
		defer os.Remove(tmpfile.Name())
		if _, err := tmpfile.Write([]byte(*infobloxConfig.CaCert)); err != nil {
			return nil, fmt.Errorf("cannot write temporary file for cacert: %w", err)
		}
		if err := tmpfile.Close(); err != nil {
			return nil, fmt.Errorf("cannot close temporary file for cacert: %w", err)
		}
		verify = tmpfile.Name()
	}

	// define transportConfig
	transportConfig := ibclient.NewTransportConfig(verify, *infobloxConfig.RequestTimeout, *infobloxConfig.PoolConnections)

	var requestBuilder ibclient.HttpRequestBuilder = &ibclient.WapiRequestBuilder{}

	dns_client, err := ibclient.NewConnector(hostConfig, transportConfig, requestBuilder, &ibclient.WapiHttpRequestor{})
	if err != nil {
		fmt.Println(err)
	}

	return &dnsClient{
		client: dns_client,
	}, nil
}

// GetManagedZones returns a map of all managed zone DNS names mapped to their IDs, composed of the project ID and
// their user assigned resource names.
func (c *dnsClient) GetManagedZones(ctx context.Context) (map[string]string, error) {

	rt := ibclient.NewZoneAuth(ibclient.ZoneAuth{})
	// urlStr := conn.RequestBuilder.BuildUrl(ibclient.GET, "allrecords", [], "", &ibclient.QueryParams{})
	urlStr := c.client.(*ibclient.Connector).RequestBuilder.BuildUrl(ibclient.GET, rt.ObjectType(), "", rt.ReturnFields(), &ibclient.QueryParams{})

	req, err := http.NewRequest("GET", urlStr, new(bytes.Buffer))
	if err != nil {
		fmt.Println(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.client.(*ibclient.Connector).HostConfig.Username, c.client.(*ibclient.Connector).HostConfig.Password)

	resp, err := c.client.(*ibclient.Connector).Requestor.SendRequest(req)
	if err != nil {
		fmt.Println(err)
	}

	rs := []ibclient.ZoneAuth{}
	err = json.Unmarshal(resp, &rs)
	if err != nil {
		fmt.Println(err)
	}

	ZoneList := make(map[string]string)

	for _, zone := range rs {
		// zone_list = append(zone_list, zone.Fqdn)
		ZoneList[zone.Ref] = zone.Fqdn
	}

	return ZoneList, nil
}

// CreateOrUpdateRecordSet creates or updates the resource recordset with the given name, record type, rrdatas, and ttl
// in the managed zone with the given name or ID.
func (c *dnsClient) CreateOrUpdateRecordSet(ctx context.Context, view, name, zone, record_type string, values []string, ttl int64) error {
	records, err := c.GetRecordSet(true, record_type, zone)
	if err != nil {
		return err
	}

	for _, record := range records {
		if record.GetDNSName() == name {
			err_del := c.DeleteRecord(record.(raw_records.Record), zone)
			if err_del != nil {
				return err_del
			}
		}
	}

	for _, value := range values {
		// r0 := c.NewRecord(name, view, zone, value, ttl, record_type)
		_, err := c.createRecord(name, view, value, ttl, record_type)
		// err := c.CreateRecord(r0)
		if err != nil {
			return err
		}
	}

	return err
}

// DeleteRecordSet deletes the resource recordset with the given name and record type
// in the managed zone with the given name or ID.
func (c *dnsClient) DeleteRecordSet(ctx context.Context, zone, name, record_type string) error {
	records, err := c.GetRecordSet(false, record_type, zone)
	if err != nil {
		return err
	}

	for _, rec := range records {
		if rec.GetId() != "" {
			err := c.DeleteRecord(rec.(raw_records.Record), zone)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// create DNS record for the Infoblox DDI setup
// func (c *dnsClient) NewRecord(name string, view string, zone string, value string, ttl int64, record_type string) raw_records.Record {

// 	var record raw_records.Record

// 	// fqdn := strings.TrimSpace(name + "." + zone)

// 	switch record_type {
// 	case raw_records.Type_A:
// 		r := ibclient.NewEmptyRecordA()
// 		r.View = view
// 		// r.Name = fqdn
// 		r.Name = name
// 		r.Ipv4Addr = value
// 		r.Ttl = uint32(ttl)
// 		record = (*raw_records.RecordA)(r)
// 	case raw_records.Type_AAAA:
// 		r := ibclient.NewEmptyRecordAAAA()
// 		r.View = view
// 		// r.Name = fqdn
// 		r.Name = name
// 		r.Ipv6Addr = value
// 		r.Ttl = uint32(ttl)
// 		record = (*raw_records.RecordAAAA)(r)
// 	case raw_records.Type_CNAME:
// 		r := ibclient.NewEmptyRecordCNAME()
// 		r.View = view
// 		// r.Name = fqdn
// 		r.Name = name
// 		r.Canonical = value
// 		r.Ttl = uint32(ttl)
// 		record = (*raw_records.RecordCNAME)(r)
// 	case raw_records.Type_TXT:
// 		if n, err := strconv.Unquote(value); err == nil && !strings.Contains(value, " ") {
// 			value = n
// 		}
// 		record = (*raw_records.RecordTXT)(ibclient.NewRecordTXT(ibclient.RecordTXT{
// 			// Name: fqdn,
// 			Name: name,
// 			Text: value,
// 			View: view,
// 		}))
// 	}

// 	return record
// }

// // func (c *dnsClient) CreateRecord(r raw_records.Record, zone string) {
// func (c *dnsClient) CreateRecord(r raw_records.Record) error {

// 	_, err := c.client.CreateObject(r.(ibclient.IBObject))
// 	return err

// }

func (c *dnsClient) createRecord(name string, view string, value string, ttl int64, record_type string) (string, error) {

	var record string
	var err error
	var rec ibclient.IBObject

	switch record_type {
	case raw_records.Type_A:
		rec = ibclient.NewRecordA(view, "", name, value, uint32(ttl), false, "", nil, "")

	case raw_records.Type_AAAA:
		rec = ibclient.NewRecordAAAA(view, name, value, false, uint32(ttl), "", nil, "")

	case raw_records.Type_CNAME:
		rec = ibclient.NewRecordCNAME(view, value, name, true, uint32(ttl), "", nil, "")

	case raw_records.Type_TXT:
		rec = ibclient.NewRecordTXT(ibclient.RecordTXT{
			Name: name,
			View: view,
			Text: value,
		})
	}

	record, err = c.client.CreateObject(rec)
	if err != nil {
		return "", err
	}

	return record, nil
}

func (c *dnsClient) DeleteRecord(record raw_records.Record, zone string) error {

	_, err := c.client.DeleteObject(record.GetId())
	if err != nil {
		return err
	}

	return nil

}

// func (c *dnsClient) GetRecordSet(name string, record_type string, zone string) (RecordSet, error) {
func (c *dnsClient) GetRecordSet(fetchAll bool, record_type string, zone string) (RecordSet, error) {

	results := c.client.(*ibclient.Connector)

	// if record_type != raw_records.Type_TXT && record_type != raw_records.Type_A && record_type != raw_records.Type_CNAME {
	// 	return nil, fmt.Errorf("record type %s not supported for GetRecord", record_type)
	// }

	execRequest := func(forceProxy bool, zone string, getall bool, recordType string) ([]byte, error) {

		var rec ibclient.IBObject
		var rt string

		if !getall {

			switch record_type {
			case "record:txt":
				rec = ibclient.NewRecordTXT(ibclient.RecordTXT{})
			case "record:a":
				rec = ibclient.NewEmptyRecordA()
			case "record:cname":
				rec = ibclient.NewEmptyRecordCNAME()
			}

			rt = rec.ObjectType()

		} else {
			rt = "allrecords"
		}

		// urlStr := results.RequestBuilder.BuildUrl(ibclient.GET, rt.ObjectType(), "", rt.ReturnFields(), &ibclient.QueryParams{})
		record_map := make(map[string]string)
		record_map["zone"] = zone
		query_params := ibclient.NewQueryParams(false, record_map)

		// urlStr := conn.RequestBuilder.BuildUrl(ibclient.GET, rt.ObjectType(), "", rt.ReturnFields(), &ibclient.QueryParams{})
		urlStr := results.RequestBuilder.BuildUrl(ibclient.GET, rt, "", nil, query_params)
		// urlStr += "&name=" + name
		if forceProxy {
			urlStr += "&_proxy_search=GM"
		}
		req, err := http.NewRequest("GET", urlStr, new(bytes.Buffer))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth(results.HostConfig.Username, results.HostConfig.Password)

		return results.Requestor.SendRequest(req)
	}

	resp, err := execRequest(false, zone, fetchAll, record_type)
	if err != nil {
		// Forcing the request to redirect to Grid Master by making forcedProxy=true
		resp, err = execRequest(true, zone, fetchAll, record_type)
	}
	if err != nil {
		return nil, err
	}

	// var rec raw_records.Base_Record

	rs := []raw_records.RecordTXT{}
	err = json.Unmarshal(resp, &rs)
	if err != nil {
		return nil, err
	}
	rs2 := RecordSet{}
	for _, r := range rs {
		rs2 = append(rs2, r.Copy())
	}

	return rs2, nil

	// switch record_type {
	// case raw_records.Type_TXT:
	// 	rs := []raw_records.RecordTXT{}
	// 	err = json.Unmarshal(resp, &rs)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	rs2 := RecordSet{}
	// 	for _, r := range rs {
	// 		rs2 = append(rs2, r.Copy())
	// 	}

	// 	return rs2, nil

	// case raw_records.Type_A:
	// 	rs := []raw_records.RecordA{}
	// 	err = json.Unmarshal(resp, &rs)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	rs2 := RecordSet{}
	// 	for _, r := range rs {
	// 		rs2 = append(rs2, r.Copy())
	// 	}

	// 	return rs2, nil

	// case raw_records.Type_CNAME:
	// 	rs := []raw_records.RecordCNAME{}
	// 	err = json.Unmarshal(resp, &rs)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	rs2 := RecordSet{}
	// 	for _, r := range rs {
	// 		rs2 = append(rs2, r.Copy())
	// 	}

	// 	return rs2, nil
	// }

	return nil, nil
}
