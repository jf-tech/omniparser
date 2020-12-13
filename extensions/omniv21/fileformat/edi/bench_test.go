package edi

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/jf-tech/go-corelib/testlib"
)

// Adding a benchmark for RawSeg operation to ensure there is no alloc:
// BenchmarkRawSeg-8   	81410766	        13.9 ns/op	       0 B/op	       0 allocs/op
func BenchmarkRawSeg(b *testing.B) {
	rawSegName := "test"
	rawSegData := []byte("test data")
	r := ediReader{
		unprocessedRawSeg: newRawSeg(),
	}
	for i := 0; i < b.N; i++ {
		r.resetRawSeg()
		r.unprocessedRawSeg.valid = true
		r.unprocessedRawSeg.Name = rawSegName
		r.unprocessedRawSeg.Raw = rawSegData
		r.unprocessedRawSeg.Elems = append(
			r.unprocessedRawSeg.Elems,
			RawSegElem{1, 1, rawSegData[0:4]}, RawSegElem{2, 1, rawSegData[5:]})
	}
}

// Adding a benchmark for stack operation to ensure there is no alloc:
// BenchmarkGrownShrinkStack-8    	12901227	        89.0 ns/op	       0 B/op	       0 allocs/op
func BenchmarkGrownShrinkStack(b *testing.B) {
	r := ediReader{
		stack: newStack(),
	}
	for i := 0; i < b.N; i++ {
		for j := 0; j < 20; j++ {
			r.growStack(stackEntry{})
		}
		for r.shrinkStack() != nil {
		}
	}
}

const (
	benchInputNoCompNoReleaseChar = `
ISA*00*          *00*          *02*CPC            *ZZ*00602679321    *191103*1800*U*00401*000001644*0*P*>
GS*QM*CPC*00602679321*20191103*1800*000001644*X*004010
ST*214*000000001
B10*4343638097845589              *4343638097845589              *CPCC
L11*4343638097845589*97
L11*0000*86
N1*SF*00602679321
N1*ST*The Rock
N3*264040 DONUT-TOWN 473*
N4*HAMMER*AB*T0C1Z0*CA
LX*000001
AT7*XB*NS***20191103*1534*
MS1*WRENCH*ON*CA
AT8*G*K*4
SE*0000000013*000000001
ST*214*000000002
B10*4343638098050296              *4343638098050296              *CPCC
L11*4343638098050296*97
L11*0156*86
N1*SF*00602679321
N1*ST*The Terminator
N3*PO BOX 12345 RPO 21ST AVE*MARKET
N4*DRILL*BC*V5L0B3*CA
LX*000001
AT7*AH*AG***20191102*1625*PT
MS1*DRILL*BC*CA
AT8*G*K*0
SE*0000000013*000000002
ST*214*000000003
B10*4343638931638575              *4343638931638575              *CPCC
L11*3001100072*97
L11*1303*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 42*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*CA*BT***20191102*1752*PT
MS1*SAW*BC*CA
AT8*G*K*0
SE*0000000013*000000003
ST*214*000000004
B10*4343638098146166              *4343638098146166              *CPCC
L11*4343638098146166*97
L11*1498*86
N1*SF*00602679321
N1*ST*The Joker
N3*RR 1*
N4*SWITCH*AB*T0C0J0*CA
LX*000001
AT7*D1*NS***20191102*1648*MT
MS1*SWITCH*AB*CA
L11**CI
AT8*G*K*15
SE*0000000014*000000004
ST*214*000000005
B10*4343638098181877              *4343638098181877              *CPCC
L11*4343638098181877*97
L11*0410*86
N1*SF*00602679321
N1*ST*The Autobot
N3*PO BOX 314*159 26 ST-GRINDER
N4*SCREWDRIVER*SK*S0M0E0*CA
LX*000001
AT7*AF*NS***20191103*1509*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*1
SE*0000000013*000000005
ST*214*000000006
B10*4343638098181891              *4343638098181891              *CPCC
L11*4343638098181891*97
L11*0410*86
N1*SF*00602679321
N1*ST*The Hobbit
N3*PO BOX 42*
N4*MITERSAW*SK*S0M1C0*CA
LX*000001
AT7*AF*NS***20191103*1509*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*19
SE*0000000013*000000006
ST*214*000000007
B10*4343638098181921              *4343638098181921              *CPCC
L11*4343638098181921*97
L11*0100*86
N1*SF*00602679321
N1*ST*The Avengers
N3*PO BOX 135*HOUSE 9 IMPACTDRILL
N4*WETVAC*MB*R0A1E0*CA
LX*000001
AT7*XB*NS***20191103*1131*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*0
SE*0000000013*000000007
ST*214*000000008
B10*4343638098186995              *4343638098186995              *CPCC
L11*4343638098186995*97
L11*0410*86
N1*SF*00602679321
N1*ST*The Hulk
N3*PO BOX 000*
N4*PRESSUREWASHER*SK*S0K2L0*CA
LX*000001
AT7*AF*NS***20191103*1509*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*0
SE*0000000013*000000008
ST*214*000000009
B10*4343638151403540              *4343638151403540              *CPCC
L11*5653512*97
L11*0175*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*AF*NS***20191103*1125*MT
MS1*LASERLEVEL*AB*CA
AT8*G*K*0
SE*0000000013*000000009
ST*214*000000010
B10*4343638151403540              *4343638151403540              *CPCC
L11*5653512*97
L11*0100*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*XB*NS***20191103*0853*MT
MS1*LASERLEVEL*AB*CA
AT8*G*K*0
SE*0000000013*000000010
ST*214*000000011
B10*4343638316026577              *4343638316026577              *CPCC
L11*300521101*97
L11*0410*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*AF*NS***20191103*0837*MT
MS1*LASERLEVEL*AB*CA
AT8*G*K*0
SE*0000000013*000000011
ST*214*000000012
B10*4343638316026577              *4343638316026577              *CPCC
L11*300521101*97
L11*0405*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*AP*AG***20191103*0815*MT
MS1*LASERLEVEL*AB*CA
AT8*G*K*0
SE*0000000013*000000012
ST*214*000000013
B10*4343638316026577              *4343638316026577              *CPCC
L11*300521101*97
L11*0410*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*AF*NS***20191102*1555*PT
MS1*NAILGUN*BC*CA
AT8*G*K*0
SE*0000000013*000000013
ST*214*000000014
B10*4343638672340607              *4343638672340607              *CPCC
L11*300633104*97
L11*0405*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*AP*AG***20191103*0733*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*0
SE*0000000013*000000014
ST*214*000000015
B10*4343638098171441              *4343638098171441              *CPCC
L11*4343638098171441*97
L11*0100*86
N1*SF*00602679321
N1*ST*The Dragon Ball
N3*PO BOX 314*159 2 ST
N4*WOODGLUE*AB*T0K2E0*CA
LX*000001
AT7*XB*NS***20191103*0904*MT
MS1*LASERLEVEL*AB*CA
AT8*G*K*16
SE*0000000013*000000015
ST*214*000000016
B10*4343638098171472              *4343638098171472              *CPCC
L11*4343638098171472*97
L11*0410*86
N1*SF*00602679321
N1*ST*The X-men
N3*PO BOX 314*1592 WIREBOX PL
N4*DRYWALL*SK*S0L0P0*CA
LX*000001
AT7*AF*NS***20191103*1509*CS
MS1*PAINTBRUSH*MB*CA
AT8*G*K*3
SE*0000000013*000000016
ST*214*000000017
B10*4343638098176088              *4343638098176088              *CPCC
L11*4343638098176088*97
L11*0405*86
N1*SF*00602679321
N1*ST*The Matrix
N3*PO BOX 314*159 UP ST
N4*FUSE*SK*S0C1S0*CA
LX*000001
AT7*AP*AG***20191103*0631*CS
MS1*SHEETROCK*SK*CA
AT8*G*K*0
SE*0000000013*000000017
ST*214*000000018
B10*4343638458862606              *4343638458862606              *CPCC
L11*3001570460*97
L11*0104*86
N1*SF*00602679321
N1*ST*
N3*PO BOX 1122*
N4*WRENCH*ON*N5Y5W3*CA
LX*000001
AT7*XB*NS***20191103*1728*ET
MS1*PVCPIPE*ON*CA
AT8*G*K*0
SE*0000000013*000000018
ST*214*000000019
B10*4343638098196574              *4343638098196574              *CPCC
L11*4343638098196574*97
L11*0104*86
N1*SF*00602679321
N1*ST*The Beast
N3*PO BOX 314*1592 HWY 551
N4*WIRECUTTER*ON*P0P1S0*CA
LX*000001
AT7*XB*NS***20191103*1618*ET
MS1*TORCH*ON*CA
AT8*G*K*1
SE*0000000013*000000019
GE*000019*000001644
IEA*00001*000001644
`

	benchDeclNoCompNoReleaseCharJSON = `
{
	"segment_delimiter": "\n",
	"element_delimiter": "*",
	"segment_declarations": [
		{
			"name": "ISA",
			"child_segments": [
				{
					"name": "GS",
					"child_segments": [
						{
							"name": "scanInfo", "type": "segment_group", "min": 0, "max": -1, "is_target": true,
							"child_segments": [
								{ "name": "ST" },
								{ "name": "B10", "elements": [ { "name": "shipment_id", "index": 2 } ] },
								{ "name": "L11" },
								{ "name": "L11" },
								{ "name": "N1" },
								{ "name": "N1" },
								{ "name": "N3" },
								{
									"name": "N4",
									"elements": [
										{ "name": "city", "index": 1 },
										{ "name": "province", "index": 2 },
										{ "name": "zip", "index": 3 },
										{ "name": "country", "index": 4 }
									]
								},
								{ "name": "LX" },
								{
									"name": "AT7",
									"elements": [
										{ "name": "status_code", "index": 1 },
										{ "name": "reason_code", "index": 2 },
										{ "name": "date", "index": 5 },
										{ "name": "time", "index": 6 },
										{ "name": "time_code", "index": 7 }
									]
								},
								{
									"name": "MS1",
									"elements": [
										{ "name": "city", "index": 1 },
										{ "name": "province", "index": 2 },
										{ "name": "country", "index": 3 }
									]
								},
								{ "name": "L11", "min": 0 },
								{ "name": "AT8" },
								{ "name": "SE" }
							]
						}
					]
				},
				{ "name": "GE" }
			]
		},
		{ "name": "IEA" }
	]
}`
)

var (
	benchDeclNoCompNoReleaseChar = func() *FileDecl {
		var fd FileDecl
		err := json.Unmarshal([]byte(benchDeclNoCompNoReleaseCharJSON), &fd)
		if err != nil {
			panic(err)
		}
		return &fd
	}()
)

// BenchmarkGetUnprocessedRawSeg_NoCompNoReleaseChar-8   	   10000	    103945 ns/op	   27080 B/op	     515 allocs/op
func BenchmarkGetUnprocessedRawSeg_NoCompNoReleaseChar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader, err := NewReader(
			"test", strings.NewReader(benchInputNoCompNoReleaseChar), benchDeclNoCompNoReleaseChar, "")
		if err != nil {
			b.FailNow()
		}
		for {
			_, err := reader.getUnprocessedRawSeg()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
			reader.resetRawSeg()
		}
	}
}

const (
	benchInputWithCompAndRelease    = `UNA:+.? 'UNB+UNOC:3+7080000714520:14+BPGARMIN:ZZZ+202502:1715+20202502171538'UNH+1+IFTSTA:D:04A:UN:BIG07'BGM+77+70726206369762052+9'DTM+137:202025021715:203'NAD+CZ+20000181402::87'CNI+1+70726206369762052'STS+1+S::87:REFORWARDED'RFF+CU:171577008'DTM+78:202002251728:203'LOC+14+01530:16::VANTAA+FI:162'GID+1+1'PCI+30+370726206369762060'STS+1+V::87:ERROR OCCURRED IN PROCESSING+99::87:Other / unknown reason+Z::87:Other / unknown reason'RFF+CU:171577008'DTM+78:202002251729:203'LOC+14+01530:16::VANTAA+FI:162'GID+1+1'PCI+30+370726206369762060'STS+1+S::87:REFORWARDED'RFF+CU:171577008'DTM+78:202002251732:203'LOC+14+01530:16::VANTAA+FI:162'GID+1+1'PCI+30+370726206369762060'UNT+24+1'UNH+2+IFTSTA:D:04A:UN:BIG07'BGM+77+70726206369728447+9'DTM+137:202025021715:203'NAD+CZ+20000181410::87'CNI+1+70726206369728447'STS+1+I::87:DELIVERED'RFF+CU:171452065'DTM+78:202002251709:203'LOC+14+0701:16::OSLO+NO:162'GID+1+1'PCI+30+370726206369728455'UNT+12+2'UNH+3+IFTSTA:D:04A:UN:BIG07'BGM+77+70726206369754293+9'DTM+137:202025021715:203'NAD+CZ+20000181410::87'CNI+1+70726206369754293'STS+1+I::87:DELIVERED'RFF+CU:171559197'DTM+78:202002251704:203'LOC+14+6101:16::VOLDA+NO:162'GID+1+1'PCI+30+370726206369754300'UNT+12+3'UNH+4+IFTSTA:D:04A:UN:BIG07'BGM+77+70726206369762663+9'DTM+137:202025021715:203'NAD+CZ+20000181410::87'CNI+1+70726206369762663'STS+1+Q::87:ARRIVED AT POST OFFICE'RFF+CU:171577030'DTM+78:202002251709:203'LOC+14+7480:16::TRONDHEIM+NO:162'GID+1+1'PCI+30+370726206369762671'UNT+12+4'UNH+5+IFTSTA:D:04A:UN:BIG07'BGM+77+70726206369768672+9'DTM+137:202025021715:203'NAD+CZ+20000181410::87'CNI+1+70726206369768672'STS+1+Q::87:ARRIVED AT POST OFFICE'RFF+CU:171608297'DTM+78:202002251709:203'LOC+14+1442:16::DRBAK+NO:162'GID+1+1'PCI+30+370726206369768680'UNT+12+5'UNZ+5+20202502171538'`
	benchDeclWithCompAndReleaseJSON = `
{
	"release_character": "?",
	"element_delimiter": "+",
	"component_delimiter": ":",
	"segment_delimiter": "'",
	"segment_declarations": [
		{
			"name": "UNA",
			"child_segments": [
				{ "name": "UNB" },
				{
					"name": "SG0", "type": "segment_group", "min": 0, "max": -1,
					"child_segments": [
						{ "name": "UNH" },
						{ "name": "BGM" },
						{ "name": "DTM", "min": 0 },
						{
							"name": "SG1_1", "type": "segment_group", "min": 0,
							"child_segments": [
								{ "name": "NAD" },
								{
									"name": "SG2", "type": "segment_group", "min": 0,
									"child_segments": [
										{ "name": "CTA" },
										{ "name": "COM" }
									]
								}
							]
						},
						{
							"name": "SG1_2", "type": "segment_group", "min": 0,
							"child_segments": [
								{ "name": "NAD" }
							]
						},
						{
							"name": "SG4", "type": "segment_group", "is_target": true, "min": 0, "max": -1,
							"child_segments": [
								{
									"name": "CNI",
									"elements": [
										{ "name": "tracking_number", "index": 2 }
									]
								},
								{
									"name": "SG5", "type": "segment_group", "max": -1,
									"child_segments": [
										{
											"name": "STS", "min": 0,
											"elements": [
												{ "name": "status_code", "index":  2, "component_index": 1 },
												{ "name": "description", "index":  2, "component_index":  4, "empty_if_missing": true }
											]
										},
										{ "name": "RFF", "min": 0 },
										{
											"name": "DTM", "min": 0,
											"elements": [
												{ "name": "event_datetime", "index": 1, "component_index": 2, "empty_if_missing": true },
												{ "name": "event_datetime_format", "index": 1, "component_index": 3, "empty_if_missing": true }
											]
										},
										{ "name": "FTX", "min": 0 },
										{
											"name": "SG6", "type": "segment_group", "min": 0,
											"child_segments": [
												{ "name": "NAD" }
											]
										},
										{
											"name": "LOC", "min": 0,
											"elements": [
												{ "name": "city", "index": 2, "component_index": 4, "empty_if_missing": true },
												{ "name": "country", "index": 3, "empty_if_missing": true }
											]
										},
										{
											"name": "SG14", "type": "segment_group", "min": 0,
											"child_segments": [
												{ "name": "GID" },
												{
													"name": "SG17", "type": "segment_group", "min": 0, "max": -1,
													"child_segments": [
														{ "name": "PCI" },
														{ "name": "GIN", "min": 0 }
													]
												}
											]
										}
									]
								}
							]
						},
						{ "name": "UNT" }
					]
				},
				{ "name": "UNZ" }
			]
		}
	]
}`
)

var (
	benchDeclWithCompAndRelease = func() *FileDecl {
		var fd FileDecl
		err := json.Unmarshal([]byte(benchDeclWithCompAndReleaseJSON), &fd)
		if err != nil {
			panic(err)
		}
		return &fd
	}()
)

// BenchmarkGetUnprocessedRawSeg_WithCompAndRelease-8    	   14174	     84115 ns/op	  102880 B/op	     385 allocs/op
func BenchmarkGetUnprocessedRawSeg_WithCompAndRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader, err := NewReader(
			"test", strings.NewReader(benchInputWithCompAndRelease), benchDeclWithCompAndRelease, "")
		if err != nil {
			b.FailNow()
		}
		for {
			_, err := reader.getUnprocessedRawSeg()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
			reader.resetRawSeg()
		}
	}
}

var (
	benchRawSegToNodeRawSeg = RawSeg{
		valid: true,
		Name:  "ISA",
		Raw:   []byte("ISA*0*1:2*3*"),
		Elems: []RawSegElem{
			{0, 1, []byte("ISA")},
			{1, 1, []byte("0")},
			{2, 1, []byte("1")},
			{2, 2, []byte("2")},
			{3, 1, []byte("3")},
		},
	}
	benchRawSegToNodeDecl = &SegDecl{
		Elems: []Elem{
			{Name: "e1", Index: 1},
			{Name: "e2c1", Index: 2, CompIndex: testlib.IntPtr(1)},
			{Name: "e2c2", Index: 2, CompIndex: testlib.IntPtr(2)},
			{Name: "e3", Index: 3},
		},
		fqdn: "ISA",
	}
	// we can do this (reusing reader and its RawSeg again & again in benchmark
	// because there is no release-char thus there is no data modification in
	// RawSeg.Elems
	benchRawSegToNodeReader = &ediReader{unprocessedRawSeg: benchRawSegToNodeRawSeg}
)

// BenchmarkRawSegToNode-8                               	  674935	      1777 ns/op	     864 B/op	       9 allocs/op
func BenchmarkRawSegToNode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := benchRawSegToNodeReader.rawSegToNode(benchRawSegToNodeDecl)
		if err != nil {
			b.FailNow()
		}
	}
}

// BenchmarkRead_NoCompNoReleaseChar-8                   	    6943	    177692 ns/op	   29620 B/op	     766 allocs/op
func BenchmarkRead_NoCompNoReleaseChar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader, err := NewReader(
			"test", strings.NewReader(benchInputNoCompNoReleaseChar), benchDeclNoCompNoReleaseChar, "")
		if err != nil {
			b.FailNow()
		}
		for {
			_, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
		}
	}
}

// BenchmarkRead_WithCompAndRelease-8                    	    9795	    120079 ns/op	  107591 B/op	     464 allocs/op
func BenchmarkRead_WithCompAndRelease(b *testing.B) {
	for i := 0; i < b.N; i++ {
		reader, err := NewReader(
			"test", strings.NewReader(benchInputWithCompAndRelease), benchDeclWithCompAndRelease, "")
		if err != nil {
			b.FailNow()
		}
		for {
			_, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				b.FailNow()
			}
		}
	}
}
