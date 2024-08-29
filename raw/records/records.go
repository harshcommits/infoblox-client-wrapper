package raw

import (
	"strconv"

	ibclient "github.com/infobloxopen/infoblox-go-client/v2"
)

const (
	Type_A     = "record:a"
	Type_CNAME = "record:cname"
	Type_AAAA  = "record:aaaa"
	Type_TXT   = "record:txt"
)

type Base_Record interface {
	GetId() string
	GetType() string
	GetValue() string
	GetDNSName() string
	GetSetIdentifier() string
	GetZone() string
	GetTTL() int
	SetTTL(int)
	Copy() Base_Record
}

type Record interface {
	Base_Record
	PrepareUpdate() Base_Record
}

type RecordA ibclient.RecordA

func (r *RecordA) GetType() string          { return Type_A }
func (r *RecordA) GetId() string            { return r.Ref }
func (r *RecordA) GetDNSName() string       { return r.Name }
func (r *RecordA) GetZone() string          { return r.Zone }
func (r *RecordA) GetSetIdentifier() string { return "" }
func (r *RecordA) GetValue() string         { return r.Ipv4Addr }
func (r *RecordA) GetTTL() int              { return int(r.Ttl) }
func (r *RecordA) SetTTL(ttl int)           { r.Ttl = uint32(ttl); r.UseTtl = ttl != 0 }
func (r *RecordA) Copy() Base_Record        { n := *r; return &n }
func (r *RecordA) PrepareUpdate() Base_Record {
	n := *r
	n.Zone = ""
	n.Name = ""
	n.View = ""
	return &n
}

type RecordAAAA ibclient.RecordAAAA

func (r *RecordAAAA) GetType() string          { return Type_AAAA }
func (r *RecordAAAA) GetId() string            { return r.Ref }
func (r *RecordAAAA) GetDNSName() string       { return r.Name }
func (r *RecordAAAA) GetZone() string          { return r.Zone }
func (r *RecordAAAA) GetSetIdentifier() string { return "" }
func (r *RecordAAAA) GetValue() string         { return r.Ipv6Addr }
func (r *RecordAAAA) GetTTL() int              { return int(r.Ttl) }
func (r *RecordAAAA) SetTTL(ttl int)           { r.Ttl = uint32(ttl); r.UseTtl = ttl != 0 }
func (r *RecordAAAA) Copy() Base_Record        { n := *r; return &n }
func (r *RecordAAAA) PrepareUpdate() Base_Record {
	n := *r
	n.Zone = ""
	n.Name = ""
	n.View = ""
	return &n
}

type RecordCNAME ibclient.RecordCNAME

func (r *RecordCNAME) GetType() string            { return Type_CNAME }
func (r *RecordCNAME) GetId() string              { return r.Ref }
func (r *RecordCNAME) GetDNSName() string         { return r.Name }
func (r *RecordCNAME) GetZone() string            { return r.Zone }
func (r *RecordCNAME) GetSetIdentifier() string   { return "" }
func (r *RecordCNAME) GetValue() string           { return r.Canonical }
func (r *RecordCNAME) GetTTL() int                { return int(r.Ttl) }
func (r *RecordCNAME) SetTTL(ttl int)             { r.Ttl = uint32(ttl); r.UseTtl = ttl != 0 }
func (r *RecordCNAME) Copy() Base_Record          { n := *r; return &n }
func (r *RecordCNAME) PrepareUpdate() Base_Record { n := *r; n.Zone = ""; n.View = ""; return &n }

type RecordTXT ibclient.RecordTXT

func (r *RecordTXT) GetType() string            { return Type_TXT }
func (r *RecordTXT) GetId() string              { return r.Ref }
func (r *RecordTXT) GetDNSName() string         { return r.Name }
func (r *RecordTXT) GetZone() string            { return r.Zone }
func (r *RecordTXT) GetSetIdentifier() string   { return "" }
func (r *RecordTXT) GetValue() string           { return EnsureQuotedText(r.Text) }
func (r *RecordTXT) GetTTL() int                { return int(r.Ttl) }
func (r *RecordTXT) SetTTL(ttl int)             { r.Ttl = uint(ttl); r.UseTtl = ttl != 0 }
func (r *RecordTXT) Copy() Base_Record          { n := *r; return &n }
func (r *RecordTXT) PrepareUpdate() Base_Record { n := *r; n.Zone = ""; n.View = ""; return &n }

var _ Base_Record = (*RecordA)(nil)
var _ Base_Record = (*RecordAAAA)(nil)
var _ Base_Record = (*RecordCNAME)(nil)
var _ Base_Record = (*RecordTXT)(nil)

type RecordNS ibclient.RecordNS

func EnsureQuotedText(v string) string {
	if _, err := strconv.Unquote(v); err != nil {
		v = strconv.Quote(v)
	}
	return v
}
