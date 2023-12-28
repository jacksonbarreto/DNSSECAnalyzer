package dnsrecords

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// NSECRecord represents a single NSEC (Next SECure) resource record in DNSSEC.
// NSEC records are used to provide authenticated denial of existence for DNS records,
// indicating which domain names do not exist in a zone. They also enumerate the types
// of records that are available for a given domain name.
//
// Fields:
//
//	TTL: An unsigned 32-bit integer indicating the time-to-live (TTL) value of the NSEC record.
//	     Specifies the duration in seconds that the record may be cached before it should be
//	     discarded or refreshed.
//
//	NextDomainName: A string representing the next domain name in the zone according to canonical
//	                ordering. Used in proving the non-existence of a name in the zone.
//
//	Types: A string listing the types of DNS resource records that exist for the domain name.
//	       This field helps in understanding the resource records available for the domain.
//
//	Validated: A boolean flag indicating whether the NSEC record has been validated
//	           using DNSSEC validation procedures. True if validated, false otherwise.
//
//	RRSIG: A pointer to an RRSIGRecord struct containing the DNSSEC signature for the NSEC record.
//	       This field may be nil if DNSSEC is not used or if the record is not signed.
type NSECRecord struct {
	TTL            uint32
	NextDomainName string
	Types          string
	Validated      bool
	RRSIG          *RRSIGRecord
}

// NewNSECRecord parses a raw DNS response string and creates a new NSECRecord struct.
// This function is designed to work with the output of the 'delv' command-line tool
// for NSEC queries in the context of DNSSEC. NSEC (Next SECure) records provide authenticated
// denial of existence for DNS records, indicating which domain names do not exist in a zone.
// The parsed information from the 'delv' response is used to populate an NSECRecord struct,
// which represents a structured interpretation of the NSEC record and associated data.
//
// Parameters:
//
//	response: A string containing the raw textual response from the 'delv' command-line tool.
//	          This response should be the result of an NSEC query for a specific domain.
//	          The response string is expected to include the NSEC record, along with
//	          any associated RRSIG records and other relevant DNSSEC information.
//
// Return Value:
//
//	*NSECRecord: A pointer to an NSECRecord struct that contains the parsed NSEC record details,
//	             including the next domain name, the types of records available for the domain,
//	             the validation status, and any associated RRSIG record.
//
//	error: An error object that indicates any issues encountered during the parsing of the
//	       response string. If the parsing is successful, the error is nil. If parsing fails,
//	       the error provides details about the cause of the failure.
//
// Example Usage:
//
//	nsecRecord, err := NewNSECRecord(rawDelvResponse)
//	if err != nil {
//	    // Handle error
//	}
//	// Use nsecRecord for DNSSEC-related tasks or analysis
//
// Note:
//
//	This function is specifically designed to parse the output of the 'delv' command-line tool,
//	which is commonly used for DNSSEC diagnostics and troubleshooting. The function assumes that the input
//	string is in the format provided by 'delv' and may not work correctly with responses from
//	other tools or in different formats.
func NewNSECRecord(response string) (*NSECRecord, error) {
	lines := strings.Split(response, "\n")
	if strings.Contains(response, "resolution failed") {
		return nil, fmt.Errorf("resolution failed: %s", lines[0])
	}

	record := &NSECRecord{}
	nsecRegex := regexp.MustCompile(`\bIN\s+NSEC\b`)
	rrsigRegex := regexp.MustCompile(`\bRRSIG\s+NSEC\b`)

	for _, line := range lines {
		if strings.HasPrefix(line, "; fully validated") {
			record.Validated = true
		} else if strings.HasPrefix(line, "; unsigned answer") {
			record.Validated = false
		} else if nsecRegex.MatchString(line) {
			parts := strings.Fields(line)
			if len(parts) < 6 {
				return nil, fmt.Errorf("invalid NSEC record: %s", line)
			}
			ttl, err := strconv.ParseUint(parts[1], 10, 32)
			if err != nil {
				return nil, fmt.Errorf("invalid TTL '%s' in NSEC record: %v", parts[1], err)
			}
			record.TTL = uint32(int(ttl))

			record.NextDomainName = parts[4]
			record.Types = strings.Join(parts[5:], ";")

		} else if rrsigRegex.MatchString(line) {
			rrsigRecord, err := NewRRSIGRecord(line)
			if err != nil {
				return nil, err
			}
			record.RRSIG = rrsigRecord
		}
	}
	return record, nil
}
