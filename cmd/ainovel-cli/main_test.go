package main

import (
	"testing"
)

func TestParseCLIOptions_Flags(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantHead   bool
		wantTrends bool
		wantReview bool
		wantNext   bool
		wantErr    bool
	}{
		{
			name:       "trends and review",
			args:       []string{"--trends", "--review"},
			wantTrends: true,
			wantReview: true,
		},
		{
			name:     "headless next",
			args:     []string{"--headless", "--next"},
			wantHead: true,
			wantNext: true,
		},
		{
			name:    "next without headless fails",
			args:    []string{"--next"},
			wantErr: true,
		},
		{
			name:       "headless trends review",
			args:       []string{"--headless", "--trends", "--review"},
			wantHead:   true,
			wantTrends: true,
			wantReview: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			opts, _, err := parseCLIOptions(tc.args)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v, got err=%v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if opts.Headless != tc.wantHead {
				t.Errorf("Headless want=%v, got=%v", tc.wantHead, opts.Headless)
			}
			if opts.Trends != tc.wantTrends {
				t.Errorf("Trends want=%v, got=%v", tc.wantTrends, opts.Trends)
			}
			if opts.Review != tc.wantReview {
				t.Errorf("Review want=%v, got=%v", tc.wantReview, opts.Review)
			}
			if opts.Next != tc.wantNext {
				t.Errorf("Next want=%v, got=%v", tc.wantNext, opts.Next)
			}
		})
	}
}

func TestParseCLIOptions_WebSecurityFlags(t *testing.T) {
	args := []string{"web", "--host", "0.0.0.0", "--port", "9090", "--token", "mysecrettoken", "--external"}
	opts, _, err := parseCLIOptions(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.Web {
		t.Errorf("expected Web=true")
	}
	if opts.Host != "0.0.0.0" {
		t.Errorf("expected Host=0.0.0.0, got %s", opts.Host)
	}
	if opts.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", opts.Port)
	}
	if opts.Token != "mysecrettoken" {
		t.Errorf("expected Token=mysecrettoken, got %s", opts.Token)
	}
	if !opts.External {
		t.Errorf("expected External=true")
	}
}

func TestParseCLIOptions_HelpFlags(t *testing.T) {
	for _, flag := range []string{"--help", "-h", "help"} {
		t.Run("Flag_"+flag, func(t *testing.T) {
			opts, _, err := parseCLIOptions([]string{flag})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !opts.Help {
				t.Fatalf("expected Help=true for flag %s", flag)
			}
		})
	}
}
