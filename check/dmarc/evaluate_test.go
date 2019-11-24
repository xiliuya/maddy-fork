package dmarc

import (
	"bufio"
	"strings"
	"testing"

	"github.com/emersion/go-message/textproto"
	"github.com/emersion/go-msgauth/authres"
	"github.com/emersion/go-msgauth/dmarc"
)

func TestEvaluateAlignment(t *testing.T) {
	type tCase struct {
		fromDomain string
		record     *dmarc.Record
		results    []authres.Result

		output authres.ResultValue
	}
	test := func(i int, c tCase) {
		out := EvaluateAlignment(c.fromDomain, c.record, c.results)
		t.Logf("%d - %+v", i, out)
		if out.Authres.Value != c.output {
			t.Errorf("%d: Wrong eval result, want '%s', got '%s' (%+v)", i, c.output, out.Authres.Value, out)
		}
	}

	cases := []tCase{
		{ // 0
			fromDomain: "example.org",
			record:     &dmarc.Record{},

			output: authres.ResultNone,
		},
		{ // 1
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultFail,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 2
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 3
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultFail,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 4
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 5
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 6
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment: dmarc.AlignmentStrict,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 7
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 8
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 9
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.com",
				},
			},
			output: authres.ResultPass,
		},
		{ // 10
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment: dmarc.AlignmentRelaxed,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "example.com",
				},
			},
			output: authres.ResultPass,
		},
		{ // 11
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment: dmarc.AlignmentStrict,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.com",
				},
			},
			output: authres.ResultPass,
		},
		{ // 12
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment:  dmarc.AlignmentStrict,
				DKIMAlignment: dmarc.AlignmentStrict,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "cbg.bounces.example.com",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "cbg.example.com",
				},
			},
			output: authres.ResultFail,
		},
		{ // 13
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.net",
				},
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "example.com",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 14
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 15
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment: dmarc.AlignmentStrict,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 16
			fromDomain: "example.com",
			record: &dmarc.Record{
				SPFAlignment: dmarc.AlignmentStrict,
			},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 17
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultTempError,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
			},
			output: authres.ResultTempError,
		},
		{ // 18
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultTempError,
					Domain: "example.com",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultTempError,
		},
		{ // 19
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultTempError,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.com",
				},
			},
			output: authres.ResultPass,
		},
		{ // 20
			fromDomain: "example.com",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.SPFResult{
					Value: authres.ResultPass,
					From:  "",
					Helo:  "mx.example.com",
				},
				&authres.DKIMResult{
					Value:  authres.ResultTempError,
					Domain: "example.com",
				},
			},
			output: authres.ResultPass,
		},
		{ // 21
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultTempError,
					Domain: "example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultPass,
		},
		{ // 22
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultFail,
					Domain: "example.org",
				},
				&authres.DKIMResult{
					Value:  authres.ResultTempError,
					Domain: "example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultTempError,
		},
		{ // 23
			fromDomain: "example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultNone,
					Domain: "example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultFail,
		},
		{ // 21
			fromDomain: "sub.example.org",
			record:     &dmarc.Record{},
			results: []authres.Result{
				&authres.DKIMResult{
					Value:  authres.ResultPass,
					Domain: "mx.example.org",
				},
				&authres.SPFResult{
					Value: authres.ResultNone,
					From:  "example.org",
					Helo:  "mx.example.org",
				},
			},
			output: authres.ResultPass,
		},
	}
	for i, case_ := range cases {
		test(i, case_)
	}
}

func TestExtractDomains(t *testing.T) {
	type tCase struct {
		hdr string

		orgDomain  string
		fromDomain string
	}
	test := func(i int, c tCase) {
		hdr, err := textproto.ReadHeader(bufio.NewReader(strings.NewReader(c.hdr + "\n\n")))
		if err != nil {
			panic(err)
		}

		orgDomain, fromDomain, err := ExtractDomains(hdr)
		if c.orgDomain == "" && err == nil {
			t.Errorf("%d: expected failure, got orgDomain = %s, fromDomain = %s", i, orgDomain, fromDomain)
			return
		}
		if c.orgDomain != "" && err != nil {
			t.Errorf("%d: unexpected error: %v", i, err)
			return
		}
		if orgDomain != c.orgDomain {
			t.Errorf("%d: want orgDomain = %v but got %s", i, c.orgDomain, orgDomain)
		}
		if fromDomain != c.fromDomain {
			t.Errorf("%d: want fromDomain = %v but got %s", i, c.fromDomain, fromDomain)
		}
	}

	cases := []tCase{
		{
			hdr:        `From: <test@example.org>`,
			orgDomain:  "example.org",
			fromDomain: "example.org",
		},
		{
			hdr:        `From: <test@foo.example.org>`,
			orgDomain:  "example.org",
			fromDomain: "foo.example.org",
		},
		{
			hdr: `From: <test@foo.example.org>, <test@bar.example.org>`,
		},
		{
			hdr: `From: <test@foo.example.org>,
From: <test@bar.example.org>`,
		},
		{
			hdr: `From: <test@>`,
		},
		{
			hdr: `From: `,
		},
		{
			hdr: `From: foo`,
		},
		{
			hdr: `From: <test@org>`,
		},
	}
	for i, case_ := range cases {
		test(i, case_)
	}
}
